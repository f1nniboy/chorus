package lrcmux

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/f1nniboy/chorus/internal/lyrics"
	"github.com/f1nniboy/chorus/internal/providers/base"
)

type Provider struct {
	base.Common

	BaseURL string `config:"base_url" default:"https://api.lrcmux.dev" label:"Base URL" type:"string"`
}

func (p *Provider) ID() string   { return "lrcmux" }
func (p *Provider) Name() string { return "lrcmux" }

func (p *Provider) Init() {
	p.BaseURL = strings.TrimRight(p.BaseURL, "/")
}

func (p *Provider) Fetch(ctx context.Context, q lyrics.TrackQuery) (lyrics.Result, error) {
	u, err := url.Parse(p.BaseURL + "/get")
	if err != nil {
		return lyrics.Result{}, fmt.Errorf("lrcmux: invalid base URL: %w", err)
	}

	query := u.Query()
	query.Set("artist", q.Artist)
	query.Set("title", q.Title)
	query.Set("format", "json")
	if q.Album != "" {
		query.Set("album", q.Album)
	}
	if q.Duration > 0 {
		query.Set("duration", strconv.FormatInt(int64(q.Duration/time.Second), 10))
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return lyrics.Result{}, err
	}

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return lyrics.Result{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return lyrics.Result{}, lyrics.ErrNotFound
	case http.StatusTooManyRequests:
		return lyrics.Result{}, fmt.Errorf("lrcmux: rate limited (retry after %s)", resp.Header.Get("Retry-After"))
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<12))
		return lyrics.Result{}, fmt.Errorf("lrcmux: HTTP %d: %s", resp.StatusCode, body)
	}

	var parsed response
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return lyrics.Result{}, fmt.Errorf("lrcmux: decode: %w", err)
	}

	return parsed.toResult(), nil
}

type response struct {
	Meta struct {
		Level string `json:"level"`
	} `json:"meta"`
	Lines []struct {
		Text  string `json:"text"`
		Start int64  `json:"start"`
		End   int64  `json:"end"`
	} `json:"lines"`
}

func (r response) toResult() lyrics.Result {
	lines := make([]lyrics.Line, 0, len(r.Lines))
	for _, l := range r.Lines {
		lines = append(lines, lyrics.Line{
			Text:  l.Text,
			Start: time.Duration(l.Start) * time.Millisecond,
			End:   time.Duration(l.End) * time.Millisecond,
		})
	}

	return lyrics.Result{Level: lyrics.Level(r.Meta.Level), Lines: lines}
}
