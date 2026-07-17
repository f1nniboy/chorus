package art

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/f1nniboy/chorus/internal/cache"
)

type Resolver struct {
	client *http.Client
	cache  *cache.Cache
}

func NewResolver(client *http.Client, c *cache.Cache) *Resolver {
	return &Resolver{client: client, cache: c}
}

func (r *Resolver) Load(ctx context.Context, artURL string) ([]byte, error) {
	if artURL == "" {
		return nil, nil
	}

	u, err := url.Parse(artURL)
	if err != nil {
		return nil, fmt.Errorf("art: parse art URL: %w", err)
	}

	switch u.Scheme {
	case "file", "":
		return os.ReadFile(u.Path)
	case "http", "https":
		return r.loadRemote(ctx, artURL)
	default:
		return nil, fmt.Errorf("art: unsupported art URL scheme %q", u.Scheme)
	}
}

func (r *Resolver) loadRemote(ctx context.Context, artURL string) ([]byte, error) {
	key := cache.Key(artURL)
	if data, ok := r.cache.Get(key); ok {
		return data, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, artURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("art: fetch %s: HTTP %d", artURL, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(key, data)
	return data, nil
}
