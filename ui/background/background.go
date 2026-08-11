package background

import (
	"context"
	"log/slog"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/art"
)

const fadeMs = 750

type Background struct {
	layers [2]*gtk.Picture
	*gtk.Overlay
	stack    *gtk.Stack
	empty    *gtk.Box
	tint     *gtk.Box
	resolver *art.Resolver
	cancel   context.CancelFunc
	lastURL  string
	raw      []byte
	blur     float64
	next     int
}

func newBackgroundLayer() *gtk.Picture {
	pic := gtk.NewPicture()
	pic.SetContentFit(gtk.ContentFitCover)
	pic.SetCanShrink(true)
	pic.AddCSSClass("cover")
	pic.SetHExpand(true)
	pic.SetVExpand(true)
	return pic
}

func New(resolver *art.Resolver) *Background {
	b := &Background{resolver: resolver}
	b.layers[0] = newBackgroundLayer()
	b.layers[1] = newBackgroundLayer()
	b.empty = gtk.NewBox(gtk.OrientationVertical, 0)

	stack := gtk.NewStack()
	stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)
	stack.SetTransitionDuration(fadeMs)
	stack.SetHExpand(true)
	stack.SetVExpand(true)
	stack.AddChild(b.layers[0])
	stack.AddChild(b.layers[1])
	stack.AddChild(b.empty)
	stack.SetVisibleChild(b.empty)
	b.stack = stack

	tint := gtk.NewBox(gtk.OrientationVertical, 0)
	tint.AddCSSClass("cover")
	tint.AddCSSClass("tint")
	tint.SetHExpand(true)
	tint.SetVExpand(true)
	b.tint = tint

	overlay := gtk.NewOverlay()
	overlay.AddCSSClass("cover-backdrop")
	overlay.SetChild(stack)
	overlay.AddOverlay(tint)
	b.Overlay = overlay

	return b
}

func (b *Background) SetArtURL(artURL string) {
	if artURL == b.lastURL {
		return
	}
	b.lastURL = artURL

	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}

	if artURL == "" {
		b.raw = nil
		b.stack.SetVisibleChild(b.empty)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel

	blur := b.blur
	go func() {
		raw, err := b.resolver.Load(ctx, artURL)
		if err != nil && ctx.Err() == nil {
			slog.Warn("art: load failed", "url", artURL, "err", err)
		}

		var texture *gdk.Texture
		if err == nil && raw != nil {
			texture, err = art.Background(raw, blur)
			if err != nil {
				slog.Warn("art: process failed", "url", artURL, "err", err)
			}
		}

		glib.IdleAdd(func() {
			if ctx.Err() != nil {
				return
			}
			b.raw = raw
			if texture == nil {
				b.stack.SetVisibleChild(b.empty)
				return
			}
			b.Show(texture)
		})
	}()
}

func (b *Background) SetBlur(radius float64) {
	if radius == b.blur {
		return
	}
	b.blur = radius

	if b.raw == nil {
		return
	}
	raw := b.raw
	go func() {
		texture, err := art.Background(raw, radius)
		if err != nil {
			return
		}
		glib.IdleAdd(func() {
			if radius != b.blur {
				return
			}
			b.Show(texture)
		})
	}()
}

func (b *Background) Show(p gdk.Paintabler) {
	layer := b.layers[b.next]
	b.next = 1 - b.next

	layer.SetPaintable(p)
	b.stack.SetVisibleChild(layer)
}
