package ui

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

func NewBackground(resolver *art.Resolver) *Background {
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
		b.stack.SetVisibleChild(b.empty)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel

	go func() {
		raw, err := b.resolver.Load(ctx, artURL)
		if err != nil && ctx.Err() == nil {
			slog.Warn("art: load failed", "url", artURL, "err", err)
		}

		var texture *gdk.Texture
		if err == nil && raw != nil {
			texture, err = art.Background(raw)
			if err != nil {
				slog.Warn("art: process failed", "url", artURL, "err", err)
			}
		}

		glib.IdleAdd(func() {
			if ctx.Err() != nil {
				return
			}
			if texture == nil {
				b.stack.SetVisibleChild(b.empty)
				return
			}
			b.show(texture)
		})
	}()
}

func (b *Background) show(texture *gdk.Texture) {
	layer := b.layers[b.next]
	b.next = 1 - b.next

	layer.SetPaintable(texture)
	b.stack.SetVisibleChild(layer)
}
