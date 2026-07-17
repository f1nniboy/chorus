package mpris

import (
	"context"
	"log/slog"
	"time"

	"github.com/godbus/dbus/v5"
	gompris "github.com/leberKleber/go-mpris"
)

// lets autoSelect skip refreshState re-fetching what it already probed
type knownState struct {
	status gompris.PlaybackStatus
	track  Track
}

func (m *Manager) attachPlayer(busName string, known *knownState) {
	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.current.cancel = cancel
	m.mu.Unlock()

	seekedOpts := []dbus.MatchOption{
		dbus.WithMatchInterface(playerInterface),
		dbus.WithMatchMember("Seeked"),
		dbus.WithMatchSender(busName),
	}

	if err := m.conn.AddMatchSignal(seekedOpts...); err != nil {
		slog.Error("mpris: failed to watch Seeked", "player", busName, "err", err)
	}

	go func() {
		<-ctx.Done()
		m.conn.RemoveMatchSignal(seekedOpts...)
	}()

	m.refreshState(busName, known)
}

func (m *Manager) Snapshot(busName string) Track {
	player := gompris.NewPlayerWithConnection(busName, m.conn)
	metadata, _ := player.Metadata()
	return trackFromMetadata(metadata)
}

func (m *Manager) refreshState(busName string, known *knownState) {
	player := gompris.NewPlayerWithConnection(busName, m.conn)

	m.mu.Lock()
	info := m.players[busName]
	m.mu.Unlock()

	var status gompris.PlaybackStatus
	var track Track
	if known != nil {
		status, track = known.status, known.track
	} else {
		var err error
		status, err = player.PlaybackStatus()
		if err != nil {
			slog.Warn("mpris: PlaybackStatus read failed", "player", busName, "err", err)
		}

		metadata, err := player.Metadata()
		if err != nil {
			slog.Warn("mpris: Metadata read failed", "player", busName, "err", err)
		}
		track = trackFromMetadata(metadata)
	}

	positionMicros, err := player.Position()
	if err != nil {
		positionMicros = 0
	}

	rate, err := player.Rate()
	if err != nil || rate == 0 {
		rate = 1.0
	}

	pos := time.Duration(positionMicros) * time.Microsecond

	m.mu.Lock()
	m.pos = posState{base: pos, baseAt: time.Now(), rate: rate, playing: status == gompris.PlaybackStatusPlaying}
	m.mu.Unlock()

	sendLatest(m.stateCh, State{Player: info, Status: status, Track: track, Position: pos})
}

func (m *Manager) handlePropertiesChanged(sig *dbus.Signal) {
	m.mu.Lock()
	busName, known := m.busByOwner[sig.Sender]
	isCurrent := known && busName == m.current.busName
	m.mu.Unlock()
	if !known || len(sig.Body) < 2 {
		return
	}

	iface, _ := sig.Body[0].(string)
	if iface != playerInterface {
		return
	}
	changed, _ := sig.Body[1].(map[string]dbus.Variant)

	_, statusChanged := changed["PlaybackStatus"]
	_, metadataChanged := changed["Metadata"]

	if !isCurrent {
		if metadataChanged {
			trySend(m.tracksCh, TrackUpdate{BusName: busName, Track: m.Snapshot(busName)})
		}
		return
	}

	if statusChanged || metadataChanged {
		m.refreshState(busName, nil)
		return
	}

	if v, ok := changed["Rate"]; ok {
		if rate, ok := v.Value().(float64); ok && rate != 0 {
			m.mu.Lock()
			elapsed := time.Since(m.pos.baseAt)
			m.pos.base += time.Duration(float64(elapsed) * m.pos.rate)
			m.pos.baseAt = time.Now()
			m.pos.rate = rate
			m.mu.Unlock()
		}
	}
}

func (m *Manager) tickPosition() {
	m.mu.Lock()
	p := m.pos
	m.mu.Unlock()

	if !p.playing {
		return
	}

	elapsed := time.Since(p.baseAt)
	interpolated := p.base + time.Duration(float64(elapsed)*p.rate)

	sendLatest(m.positionCh, interpolated)
}

func (m *Manager) emitPosition() {
	m.mu.Lock()
	pos := m.pos.base
	m.mu.Unlock()

	sendLatest(m.positionCh, pos)
}

func trackFromMetadata(md gompris.Metadata) Track {
	if md == nil {
		return Track{}
	}

	title, _ := md.XESAMTitle()
	album, _ := md.XESAMAlbum()
	artURL, _ := md.MPRISArtURL()
	lengthMicros, _ := md.MPRISLength()
	artists, _ := md.XESAMArtist()
	var artist string
	if len(artists) > 0 {
		artist = artists[0]
	}

	return Track{
		Title:  title,
		Artist: artist,
		Album:  album,
		ArtURL: artURL,
		Length: time.Duration(lengthMicros) * time.Microsecond,
	}
}
