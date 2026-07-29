package mpris

import (
	"context"
	"fmt"
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

func (m *Manager) SeekTo(pos time.Duration) error {
	m.mu.Lock()
	busName := m.current.busName
	track := m.tracks[busName]
	p := m.pos
	m.mu.Unlock()

	if busName == "" {
		return nil
	}

	obj := m.conn.Object(busName, objectPath)

	// some players (e.g. Amberol) never set mpris:trackid, so we have to use
	// relative Seek for them
	var call *dbus.Call
	if track.ID != "" {
		call = obj.Call(playerInterface+".SetPosition", 0, track.ID, int64(pos/time.Microsecond))
	} else {
		call = obj.Call(playerInterface+".Seek", 0, int64((pos-p.interpolated())/time.Microsecond))
	}
	if call.Err != nil {
		return call.Err
	}

	// otherwise the seek looks like it did nothing to the user
	if !p.playing {
		if err := obj.Call(playerInterface+".Play", 0).Err; err != nil {
			return fmt.Errorf("mpris: seeked but failed to resume playback: %w", err)
		}
	}

	return nil
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

	rate, err := player.Rate()
	if err != nil || rate == 0 {
		rate = 1.0
	}

	canSeek, _ := player.CanSeek()

	positionMicros, _ := player.Position()
	pos := time.Duration(positionMicros) * time.Microsecond

	m.setTrack(busName, track)

	m.mu.Lock()
	if busName != m.current.busName {
		m.mu.Unlock()
		return
	}
	m.pos = posState{base: pos, baseAt: time.Now(), rate: rate, playing: status == gompris.PlaybackStatusPlaying}
	sendLatest(m.playbackCh, Playback{Player: info, Status: status, Track: track, Position: pos, CanSeek: canSeek})
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

func trackFromMetadata(m gompris.Metadata) Track {
	if m == nil {
		return Track{}
	}

	title, _ := m.XESAMTitle()
	album, _ := m.XESAMAlbum()
	artURL, _ := m.MPRISArtURL()
	lengthMicros, _ := m.MPRISLength()
	artists, _ := m.XESAMArtist()
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
		ID:     trackID(m),
	}
}

// we can't use go-mpris' MPRISTrackID directly, because it only strictly
// accepts a dbus.ObjectPath, and some clients (like Spotify...) send it
// as a plain string instead
func trackID(md gompris.Metadata) dbus.ObjectPath {
	switch v := md["mpris:trackid"].Value().(type) {
	case dbus.ObjectPath:
		return v
	case string:
		return dbus.ObjectPath(v)
	default:
		return ""
	}
}
