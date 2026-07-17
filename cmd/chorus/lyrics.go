package main

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
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

	fetcherMu sync.Mutex
	fetcher   *lyrics.Fetcher

	lastTrackKey string
	currentTrack mpris.Track
	lastPosition time.Duration
	fetchKey     string
	cancel       context.CancelFunc
}

func newLyricsController(cfg *config.Config, httpClient *http.Client, diskCache *cache.Cache, view *ui.LyricsView) (*lyricsController, error) {
	p, err := providers.New(cfg.ProviderName(), cfg.ProviderConfig(cfg.ProviderName()), httpClient)
	if err != nil {
		return nil, err
	}
	return &lyricsController{
		cfg:        cfg,
		httpClient: httpClient,
		diskCache:  diskCache,
		view:       view,
		fetcher:    lyrics.NewFetcher(p, diskCache),
	}, nil
}

func (c *lyricsController) RebuildProvider() {
	p, err := providers.New(c.cfg.ProviderName(), c.cfg.ProviderConfig(c.cfg.ProviderName()), c.httpClient)
	if err != nil {
		slog.Error("providers: rebuild failed", "err", err)
		return
	}
	c.fetcherMu.Lock()
	c.fetcher = lyrics.NewFetcher(p, c.diskCache)
	c.fetcherMu.Unlock()

	// a different provider may have different lyrics for the same track
	if c.lastTrackKey != "" {
		c.fetch(c.currentTrack)
	}
}

func (c *lyricsController) TrackChanged(track mpris.Track, pos time.Duration) {
	key := track.Key()
	if key == c.lastTrackKey {
		return
	}
	c.lastTrackKey = key
	c.currentTrack = track
	c.lastPosition = pos
	c.fetch(track)
}

func (c *lyricsController) UpdatePosition(pos time.Duration) {
	c.lastPosition = pos
}

func (c *lyricsController) Idle() {
	c.lastTrackKey = ""
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

		c.fetcherMu.Lock()
		f := c.fetcher
		c.fetcherMu.Unlock()

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
