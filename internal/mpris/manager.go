package mpris

import (
	"context"
	"log/slog"
	"maps"
	"slices"
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

// one player in the roster, with its last known track
type Entry struct {
	Player Player
	Track  Track
}

// the selected player's full playback state
type Playback struct {
	Player   Player
	Status   gompris.PlaybackStatus
	Track    Track
	Position time.Duration
}

func (p Playback) IsIdle() bool {
	return p.Player.BusName == "" || p.Status == gompris.PlaybackStatusStopped
}

// the currently-selected player's identity plus its detach hook
type attachment struct {
	cancel  context.CancelFunc
	busName string
}

// interpolates live playback position between MPRIS polls
type posState struct {
	baseAt  time.Time
	base    time.Duration
	rate    float64
	playing bool
}

type Manager struct {
	current    attachment
	conn       *dbus.Conn
	rosterCh   chan []Entry
	playbackCh chan Playback
	positionCh chan time.Duration
	players    map[string]Player
	tracks     map[string]Track
	busByOwner map[string]string
	preferred  string
	pos        posState
	mu         sync.Mutex
}

func New(conn *dbus.Conn, preferredIdentity string) *Manager {
	return &Manager{
		conn:       conn,
		rosterCh:   make(chan []Entry, 1),
		playbackCh: make(chan Playback, 1),
		positionCh: make(chan time.Duration, 1),
		players:    map[string]Player{},
		tracks:     map[string]Track{},
		busByOwner: map[string]string{},
		preferred:  preferredIdentity,
		pos:        posState{rate: 1.0},
	}
}

func (m *Manager) Roster() <-chan []Entry         { return m.rosterCh }
func (m *Manager) Playback() <-chan Playback      { return m.playbackCh }
func (m *Manager) Position() <-chan time.Duration { return m.positionCh }

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
	m.selectBusName(busName, false)
}

func (m *Manager) SelectPlayerManually(p Player) {
	m.mu.Lock()
	m.preferred = p.Identity
	m.mu.Unlock()

	m.SelectPlayer(p.BusName)
}

func (m *Manager) selectBusName(busName string, onlyIfEmpty bool) {
	m.mu.Lock()
	_, ok := m.players[busName]
	if !ok || busName == m.current.busName || (onlyIfEmpty && m.current.busName != "") {
		m.mu.Unlock()
		return
	}
	if m.current.cancel != nil {
		m.current.cancel()
	}
	m.current.busName = busName
	m.mu.Unlock()

	m.attachPlayer(busName)
}

func (m *Manager) handleSignal(sig *dbus.Signal) {
	switch sig.Name {
	case "org.freedesktop.DBus.NameOwnerChanged":
		var name, oldOwner, newOwner string
		if err := dbus.Store(sig.Body, &name, &oldOwner, &newOwner); err != nil {
			return
		}
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
		var micros int64
		if err := dbus.Store(sig.Body, &micros); err != nil {
			return
		}

		m.mu.Lock()
		busName, known := m.busByOwner[sig.Sender]
		isCurrent := known && busName == m.current.busName
		m.mu.Unlock()
		if !isCurrent {
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
	m.tracks = map[string]Track{}
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

	m.emitRoster()
	m.autoSelect()

	go func() {
		m.setTrack(busName, m.Snapshot(busName))
	}()
}

func (m *Manager) setTrack(busName string, track Track) {
	m.mu.Lock()
	if _, ok := m.players[busName]; !ok {
		m.mu.Unlock()
		return
	}
	m.tracks[busName] = track
	m.mu.Unlock()

	m.emitRoster()
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
	delete(m.tracks, busName)
	wasCurrent := m.current.busName == busName
	if wasCurrent {
		if m.current.cancel != nil {
			m.current.cancel()
		}
		m.current = attachment{}
	}
	m.mu.Unlock()

	m.conn.RemoveMatchSignal(propsMatchOpts(busName)...)

	m.emitRoster()

	if wasCurrent && !m.autoSelect() {
		m.mu.Lock()
		if m.current.busName == "" {
			sendLatest(m.playbackCh, Playback{})
		}
		m.mu.Unlock()
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

	var (
		best      string
		bestScore = -1
	)
	for _, p := range players {
		if !m.Snapshot(p.BusName).Valid() {
			continue
		}
		status, err := gompris.NewPlayerWithConnection(p.BusName, m.conn).PlaybackStatus()
		if err != nil {
			status = ""
		}

		score := 0
		if status == gompris.PlaybackStatusPlaying {
			score += 2
		}
		if p.Identity == preferred {
			score++
		}
		if score > bestScore {
			best, bestScore = p.BusName, score
		}
	}

	if best == "" {
		return false
	}
	m.selectBusName(best, true)
	return true
}

// runs from several goroutines, an older snapshot could replace a newer
// snapshot with an older one, so we have to lock
func (m *Manager) emitRoster() {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := make([]Entry, 0, len(m.players))
	for bus, p := range m.players {
		track := m.tracks[bus]
		if !track.Valid() {
			continue
		}
		list = append(list, Entry{Player: p, Track: track})
	}
	slices.SortFunc(list, func(a, b Entry) int { return strings.Compare(a.Player.Identity, b.Player.Identity) })

	sendLatest(m.rosterCh, list)
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
