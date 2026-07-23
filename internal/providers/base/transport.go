package base

import (
	"net/http"
	"time"

	"github.com/f1nniboy/chorus/internal/meta"
)

const userAgent = meta.AppName + "/" + meta.Version + " (" + meta.AppRepo + ")"

type uaTransport struct{ inner http.RoundTripper }

func (t uaTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		clone := r.Clone(r.Context())
		clone.Header.Set("User-Agent", userAgent)
		r = clone
	}
	return t.inner.RoundTrip(r)
}

func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: uaTransport{inner: http.DefaultTransport},
	}
}
