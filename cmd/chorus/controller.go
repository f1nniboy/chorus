package main

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"

	"github.com/f1nniboy/chorus/internal/cache"
	"github.com/f1nniboy/chorus/internal/config"
	"github.com/f1nniboy/chorus/internal/lyrics"
	"github.com/f1nniboy/chorus/internal/mpris"
	"github.com/f1nniboy/chorus/internal/providers"
	"github.com/f1nniboy/chorus/ui"
)

const fetchTimeout = 30 * time.Second

type controller struct {
	cfg        *config.Config
	httpClient *http.Client
	diskCache  *cache.Cache
	view       *ui.LyricsView
	mgr        *mpris.Manager
	fetcher    atomic.Pointer[lyrics.Fetcher]
	cancel     context.CancelFunc
	fetchKey   string
	playback   mpris.Playback
}

func newController(cfg *config.Config, httpClient *http.Client, diskCache *cache.Cache, view *ui.LyricsView, mgr *mpris.Manager) (*controller, error) {
	name := cfg.ProviderName()
	p, err := providers.New(name, cfg.ProviderConfig(name), httpClient)
	if err != nil {
		return nil, err
	}
	c := &controller{
		cfg:        cfg,
		httpClient: httpClient,
		diskCache:  diskCache,
		view:       view,
		mgr:        mgr,
	}
	c.fetcher.Store(lyrics.NewFetcher(p, diskCache))
	return c, nil
}

func (c *controller) SeekTo(pos time.Duration) {
	if err := c.mgr.SeekTo(pos); err != nil {
		slog.Warn("mpris: seek failed", "err", err)
	}
}

func (c *controller) RebuildProvider() {
	name := c.cfg.ProviderName()
	cfg := c.cfg.ProviderConfig(name)

	go func() {
		p, err := providers.New(name, cfg, c.httpClient)
		if err != nil {
			slog.Error("providers: rebuild failed", "err", err)
			return
		}
		c.fetcher.Store(lyrics.NewFetcher(p, c.diskCache))

		glib.IdleAdd(func() {
			// a different provider may have different lyrics for the same track
			if c.playback.Track.Key() != "" {
				c.fetch(c.playback.Track)
			}
		})
	}()
}

func (c *controller) ChangeTrack(pb mpris.Playback) {
	changed := pb.Track.Key() != c.playback.Track.Key()
	c.playback = pb
	if changed {
		c.fetch(pb.Track)
	}
}

func (c *controller) UpdatePosition(pos time.Duration) {
	c.playback.Position = pos
	c.view.SetPosition(pos)
}

func (c *controller) Idle() {
	c.playback = mpris.Playback{}
	c.fetchKey = ""
	if c.cancel != nil {
		c.cancel()
	}
	c.view.SetIdle()
}

func (c *controller) fetch(track mpris.Track) {
	c.view.SetLoading()

	if c.cancel != nil {
		c.cancel()
	}

	key := c.cfg.ProviderName() + track.Key()
	c.fetchKey = key

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	c.cancel = cancel

	go func() {
		defer cancel()

		f := c.fetcher.Load()

		res, err := f.Get(ctx, lyrics.TrackQuery{
			Artist:   track.Artist,
			Title:    track.Title,
			Album:    track.Album,
			Duration: track.Length,
		})

		glib.IdleAdd(func() {
			if key != c.fetchKey {
				return
			}
			c.view.SetResult(res, err, c.playback.Position, c.playback.CanSeek)
		})
	}()
}
