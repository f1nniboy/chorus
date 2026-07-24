package mpris

import (
	"context"
	"log/slog"
	"time"

	"github.com/godbus/dbus/v5"
	gompris "github.com/leberKleber/go-mpris"
)

func (m *Manager) attachPlayer(busName string) {
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

	m.refreshState(busName)
}

func (m *Manager) Snapshot(busName string) Track {
	player := gompris.NewPlayerWithConnection(busName, m.conn)
	metadata, _ := player.Metadata()
	return trackFromMetadata(metadata)
}

func (m *Manager) refreshState(busName string) {
	player := gompris.NewPlayerWithConnection(busName, m.conn)

	m.mu.Lock()
	info := m.players[busName]
	m.mu.Unlock()

	status, err := player.PlaybackStatus()
	if err != nil {
		slog.Warn("mpris: PlaybackStatus read failed", "player", busName, "err", err)
	}

	metadata, err := player.Metadata()
	if err != nil {
		slog.Warn("mpris: Metadata read failed", "player", busName, "err", err)
	}
	track := trackFromMetadata(metadata)

	positionMicros, err := player.Position()
	if err != nil {
		positionMicros = 0
	}

	rate, err := player.Rate()
	if err != nil || rate == 0 {
		rate = 1.0
	}

	pos := time.Duration(positionMicros) * time.Microsecond

	m.setTrack(busName, track)

	m.mu.Lock()
	if busName != m.current.busName {
		m.mu.Unlock()
		return
	}
	m.pos = posState{base: pos, baseAt: time.Now(), rate: rate, playing: status == gompris.PlaybackStatusPlaying}
	sendLatest(m.playbackCh, Playback{Player: info, Status: status, Track: track, Position: pos})
	m.mu.Unlock()
}

func (m *Manager) handlePropertiesChanged(sig *dbus.Signal) {
	m.mu.Lock()
	busName, known := m.busByOwner[sig.Sender]
	isCurrent := known && busName == m.current.busName
	m.mu.Unlock()
	if !known {
		return
	}

	var iface string
	var changed map[string]dbus.Variant
	var invalidated []string
	if err := dbus.Store(sig.Body, &iface, &changed, &invalidated); err != nil {
		return
	}
	if iface != playerInterface {
		return
	}

	_, statusChanged := changed["PlaybackStatus"]
	_, metadataChanged := changed["Metadata"]
	_, rateChanged := changed["Rate"]

	if !isCurrent {
		if metadataChanged {
			m.setTrack(busName, m.Snapshot(busName))
		}
		return
	}

	if statusChanged || metadataChanged || rateChanged {
		m.refreshState(busName)
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
