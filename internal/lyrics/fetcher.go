package lyrics

import (
	"context"
	"encoding/json"

	"github.com/f1nniboy/chorus/internal/cache"
)

type Fetcher struct {
	provider Provider
	cache    *cache.Cache
}

func NewFetcher(p Provider, c *cache.Cache) *Fetcher {
	return &Fetcher{provider: p, cache: c}
}

func (f *Fetcher) Get(ctx context.Context, q TrackQuery) (Result, error) {
	key := cache.Key(f.provider.ID(), q.Artist, q.Title, q.Album)

	if data, ok := f.cache.Get(key); ok {
		var res Result
		if err := json.Unmarshal(data, &res); err == nil {
			return res, nil
		}
	}

	res, err := f.provider.Fetch(ctx, q)
	if err != nil {
		return Result{}, err
	}

	if data, err := json.Marshal(res); err == nil {
		_ = f.cache.Set(key, data)
	}
	return res, nil
}
