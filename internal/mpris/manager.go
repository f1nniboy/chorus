package mpris

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	gompris "github.com/leberKleber/go-mpris"
)

const (
	busNamePrefix = "org.mpris.MediaPlayer2."

	objectPath      = "/org/mpris/MediaPlayer2"
	propsInterface  = "org.freedesktop.DBus.Properties"
	playerInterface = "org.mpris.MediaPlayer2.Player"
	appInterface    = "org.mpris.MediaPlayer2"

	positionTickInterval = 200 * time.Millisecond
)

type Player struct {
	BusName  string
	Identity string
}

type Track struct {
	Title  string
	Artist string
	Album  string
	ArtURL string
	Length time.Duration
}

func (t Track) Valid() bool {
	return t.Artist != ""
}

func (t Track) Key() string {
	return t.Artist + t.Title + t.Album
}

type State struct {
	Player   Player
	Status   gompris.PlaybackStatus
	Track    Track
	Position time.Duration
}

func (s State) IsIdle() bool {
	return s.Player.BusName == "" || s.Status == gompris.PlaybackStatusStopped
}

// a single background (non-current) player's refreshed track
type TrackUpdate struct {
	BusName string
	Track   Track
}

// the currently-selected player's identity plus its detach hook
type attachment struct {
	busName string
	cancel  context.CancelFunc
}

// interpolates live playback position between MPRIS polls
type posState struct {
	base    time.Duration
	baseAt  time.Time
	rate    float64
	playing bool
}

type Manager struct {
	conn *dbus.Conn

	playersCh  chan []Player
	stateCh    chan State
	positionCh chan time.Duration
	tracksCh   chan TrackUpdate

	mu         sync.Mutex
	players    map[string]Player
	busByOwner map[string]string
	current    attachment
	preferred  string
	pos        posState

	onPreferredChanged func(identity string)
}

func New(conn *dbus.Conn, preferredIdentity string) *Manager {
	return &Manager{
		conn:       conn,
		playersCh:  make(chan []Player, 4),
		stateCh:    make(chan State, 4),
		positionCh: make(chan time.Duration, 16),
		tracksCh:   make(chan TrackUpdate, 8),
		players:    map[string]Player{},
		busByOwner: map[string]string{},
		preferred:  preferredIdentity,
		pos:        posState{rate: 1.0},
	}
}

func (m *Manager) Players() <-chan []Player       { return m.playersCh }
func (m *Manager) State() <-chan State            { return m.stateCh }
func (m *Manager) Position() <-chan time.Duration { return m.positionCh }
func (m *Manager) Tracks() <-chan TrackUpdate     { return m.tracksCh }

func (m *Manager) Start(ctx context.Context) error {
	if err := m.conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
	); err != nil {
		return err
	}

	sigCh := make(chan *dbus.Signal, 32)
	m.conn.Signal(sigCh)

	m.rescanPlayers()

	ticker := time.NewTicker(positionTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigCh:
			m.handleSignal(sig)
		case <-ticker.C:
			m.tickPosition()
		}
	}
}

func (m *Manager) SelectPlayer(busName string) {
	m.selectBusName(busName, nil)
}

// fires whenever SelectPlayerManually updates the remembered identity
func (m *Manager) OnPreferredChanged(f func(identity string)) {
	m.onPreferredChanged = f
}

func (m *Manager) SelectPlayerManually(p Player) {
	m.mu.Lock()
	m.preferred = p.Identity
	m.mu.Unlock()

	if m.onPreferredChanged != nil {
		m.onPreferredChanged(p.Identity)
	}

	m.SelectPlayer(p.BusName)
}

func (m *Manager) selectBusName(busName string, known *knownState) {
	m.mu.Lock()
	_, ok := m.players[busName]
	if !ok || busName == m.current.busName {
		m.mu.Unlock()
		return
	}
	if m.current.cancel != nil {
		m.current.cancel()
	}
	m.current.busName = busName
	m.mu.Unlock()

	m.attachPlayer(busName, known)
}

func (m *Manager) handleSignal(sig *dbus.Signal) {
	switch sig.Name {
	case "org.freedesktop.DBus.NameOwnerChanged":
		if len(sig.Body) != 3 {
			return
		}
		name, _ := sig.Body[0].(string)
		newOwner, _ := sig.Body[2].(string)
		if !strings.HasPrefix(name, busNamePrefix) {
			return
		}
		if newOwner == "" {
			m.playerVanished(name)
		} else {
			m.playerAppeared(name)
		}
	case propsInterface + ".PropertiesChanged":
		m.handlePropertiesChanged(sig)
	case playerInterface + ".Seeked":
		if len(sig.Body) != 1 {
			return
		}
		m.mu.Lock()
		busName, known := m.busByOwner[sig.Sender]
		isCurrent := known && busName == m.current.busName
		m.mu.Unlock()
		if !isCurrent {
			return
		}
		micros, ok := sig.Body[0].(int64)
		if !ok {
			return
		}
		m.mu.Lock()
		m.pos.base = time.Duration(micros) * time.Microsecond
		m.pos.baseAt = time.Now()
		m.mu.Unlock()
		m.emitPosition()
	}
}

func (m *Manager) rescanPlayers() {
	var names []string
	if err := m.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
		slog.Error("mpris: ListNames failed", "err", err)
		return
	}

	m.mu.Lock()
	m.players = map[string]Player{}
	m.mu.Unlock()

	for _, name := range names {
		if strings.HasPrefix(name, busNamePrefix) {
			m.playerAppeared(name)
		}
	}
}

func propsMatchOpts(busName string) []dbus.MatchOption {
	return []dbus.MatchOption{
		dbus.WithMatchObjectPath(objectPath),
		dbus.WithMatchInterface(propsInterface),
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchSender(busName),
	}
}

func (m *Manager) playerAppeared(busName string) {
	identity, _ := m.getStringProp(busName, appInterface, "Identity")
	if identity == "" {
		identity = strings.TrimPrefix(busName, busNamePrefix)
	}

	var owner string
	if err := m.conn.BusObject().Call("org.freedesktop.DBus.GetNameOwner", 0, busName).Store(&owner); err != nil {
		slog.Error("mpris: GetNameOwner failed", "player", busName, "err", err)
	}

	if err := m.conn.AddMatchSignal(propsMatchOpts(busName)...); err != nil {
		slog.Error("mpris: failed to watch PropertiesChanged", "player", busName, "err", err)
	}

	m.mu.Lock()
	m.players[busName] = Player{BusName: busName, Identity: identity}
	m.busByOwner[owner] = busName
	m.mu.Unlock()

	m.emitPlayers()
	m.autoSelect()

	go func() {
		trySend(m.tracksCh, TrackUpdate{BusName: busName, Track: m.Snapshot(busName)})
	}()
}

func (m *Manager) playerVanished(busName string) {
	m.mu.Lock()
	for owner, bus := range m.busByOwner {
		if bus == busName {
			delete(m.busByOwner, owner)
			break
		}
	}
	delete(m.players, busName)
	wasCurrent := m.current.busName == busName
	if wasCurrent {
		if m.current.cancel != nil {
			m.current.cancel()
		}
		m.current = attachment{}
	}
	m.mu.Unlock()

	m.conn.RemoveMatchSignal(propsMatchOpts(busName)...)

	m.emitPlayers()

	if wasCurrent && !m.autoSelect() {
		sendLatest(m.stateCh, State{})
	}
}

// prefers the remembered identity, then whatever's playing, then any valid player
func (m *Manager) autoSelect() bool {
	m.mu.Lock()
	if m.current.busName != "" {
		m.mu.Unlock()
		return false
	}
	preferred := m.preferred
	players := slices.SortedFunc(maps.Values(m.players), func(a, b Player) int {
		return strings.Compare(a.BusName, b.BusName)
	})
	m.mu.Unlock()

	if preferred != "" {
		for _, p := range players {
			if p.Identity == preferred {
				m.SelectPlayer(p.BusName)
				return true
			}
		}
	}

	var fallback string
	var fallbackStatus gompris.PlaybackStatus
	for _, p := range players {
		name := p.BusName
		var status gompris.PlaybackStatus
		if s, err := m.getStringProp(name, playerInterface, "PlaybackStatus"); err == nil {
			status = gompris.PlaybackStatus(s)
		}

		if fallback == "" {
			fallback = name
			fallbackStatus = status
		}

		if status != gompris.PlaybackStatusPlaying {
			continue
		}
		track := m.Snapshot(name)
		if !track.Valid() {
			continue
		}
		m.selectBusName(name, &knownState{status: status, track: track})
		return true
	}
	if fallback != "" {
		track := m.Snapshot(fallback)
		if track.Valid() {
			m.selectBusName(fallback, &knownState{status: fallbackStatus, track: track})
			return true
		}
	}
	return false
}

func (m *Manager) emitPlayers() {
	m.mu.Lock()
	list := make([]Player, 0, len(m.players))
	for _, p := range m.players {
		list = append(list, p)
	}
	m.mu.Unlock()

	sort.Slice(list, func(i, j int) bool { return list[i].Identity < list[j].Identity })

	sendLatest(m.playersCh, list)
}

func (m *Manager) getStringProp(busName, iface, prop string) (string, error) {
	v, err := m.conn.Object(busName, objectPath).GetProperty(iface + "." + prop)
	if err != nil {
		return "", err
	}
	s, _ := v.Value().(string)
	return s, nil
}

func sendLatest[T any](ch chan T, v T) {
	for {
		select {
		case ch <- v:
			return
		default:
			select {
			case <-ch:
			default:
			}
		}
	}
}

func trySend[T any](ch chan T, v T) {
	select {
	case ch <- v:
	default:
	}
}
