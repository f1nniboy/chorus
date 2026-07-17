package providers

import (
	"fmt"
	"net/http"

	"github.com/f1nniboy/chorus/internal/lyrics"
	"github.com/f1nniboy/chorus/internal/providers/lrcmux"
)

// fresh instance per call, New runs again whenever settings change
var factories = []func() lyrics.Provider{
	func() lyrics.Provider { return &lrcmux.Provider{} },
}

type Info struct {
	ID, Name string
}

func Available() []Info {
	infos := make([]Info, len(factories))
	for i, f := range factories {
		p := f()
		infos[i] = Info{ID: p.ID(), Name: p.Name()}
	}
	return infos
}

func New(name string, cfg map[string]any, client *http.Client) (lyrics.Provider, error) {
	for _, f := range factories {
		p := f()
		if p.ID() != name {
			continue
		}
		decodeConfig(p, cfg)
		p.SetDeps(client)
		p.Init()
		return p, nil
	}
	return nil, fmt.Errorf("unknown provider: %s", name)
}
