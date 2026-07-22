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

const fetchTimeout = 15 * time.Second

type lyricsController struct {
	cfg        *config.Config
	httpClient *http.Client
	diskCache  *cache.Cache
	view       *ui.LyricsView

	fetcher atomic.Pointer[lyrics.Fetcher]

	currentTrack mpris.Track
	lastPosition time.Duration
	fetchKey     string
	cancel       context.CancelFunc
}

func newLyricsController(cfg *config.Config, httpClient *http.Client, diskCache *cache.Cache, view *ui.LyricsView) (*lyricsController, error) {
	name := cfg.ProviderName()
	p, err := providers.New(name, cfg.ProviderConfig(name), httpClient)
	if err != nil {
		return nil, err
	}
	c := &lyricsController{
		cfg:        cfg,
		httpClient: httpClient,
		diskCache:  diskCache,
		view:       view,
	}
	c.fetcher.Store(lyrics.NewFetcher(p, diskCache))
	return c, nil
}

func (c *lyricsController) RebuildProvider() {
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
			if c.currentTrack.Key() != "" {
				c.fetch(c.currentTrack)
			}
		})
	}()
}

func (c *lyricsController) TrackChanged(track mpris.Track, pos time.Duration) {
	if track.Key() == c.currentTrack.Key() {
		return
	}
	c.currentTrack = track
	c.lastPosition = pos
	c.fetch(track)
}

func (c *lyricsController) UpdatePosition(pos time.Duration) {
	c.lastPosition = pos
}

func (c *lyricsController) Idle() {
	c.currentTrack = mpris.Track{}
	c.fetchKey = ""
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *lyricsController) fetch(track mpris.Track) {
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
			c.view.SetResult(res, err, c.lastPosition)
		})
	}()
}
