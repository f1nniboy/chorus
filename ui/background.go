package ui

import (
	"context"
	"log/slog"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/art"
)

const fadeMs = 1000

type Background struct {
	*gtk.Overlay
	back       *gtk.Picture
	front      *gtk.Picture
	dim        *gtk.Box
	resolver   *art.Resolver
	cancel     context.CancelFunc
	fadeAnim   *adw.TimedAnimation
	lastArtURL string
}

func NewBackground(resolver *art.Resolver) *Background {
	overlay := gtk.NewOverlay()

	newLayer := func() *gtk.Picture {
		pic := gtk.NewPicture()
		pic.SetContentFit(gtk.ContentFitCover)
		pic.SetCanShrink(true)
		pic.AddCSSClass("cover")
		pic.SetHExpand(true)
		pic.SetVExpand(true)
		pic.SetOpacity(0)
		return pic
	}

	back := newLayer()
	front := newLayer()
	overlay.SetChild(back)
	overlay.AddOverlay(front)

	dim := gtk.NewBox(gtk.OrientationVertical, 0)
	dim.AddCSSClass("cover")
	dim.AddCSSClass("dim")
	dim.SetHExpand(true)
	dim.SetVExpand(true)
	overlay.AddOverlay(dim)

	return &Background{Overlay: overlay, back: back, front: front, dim: dim, resolver: resolver}
}

func (b *Background) SetArtURL(artURL string) {
	if artURL == b.lastArtURL {
		return
	}
	b.lastArtURL = artURL

	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}

	if artURL == "" {
		b.transitionTo(nil)
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
			b.transitionTo(texture)
		})
	}()
}

func (b *Background) transitionTo(texture *gdk.Texture) {
	if b.fadeAnim != nil {
		b.fadeAnim.Pause()
	}

	if texture == nil {
		b.front.SetOpacity(0)
		b.front.SetPaintable(nil)

		anim := adw.NewTimedAnimation(b.back, b.back.Opacity(), 0, fadeMs,
			adw.NewCallbackAnimationTarget(func(value float64) {
				b.back.SetOpacity(value)
			}),
		)
		anim.ConnectDone(func() { b.back.SetPaintable(nil) })
		b.fadeAnim = anim
		anim.Play()
		return
	}

	from := b.front.Opacity()
	b.front.SetPaintable(texture)

	anim := adw.NewTimedAnimation(b.front, from, 1, fadeMs,
		adw.NewCallbackAnimationTarget(func(value float64) {
			b.front.SetOpacity(value)
		}),
	)
	anim.ConnectDone(func() {
		b.back.SetPaintable(texture)
		b.back.SetOpacity(1)
		b.front.SetOpacity(0)
		b.front.SetPaintable(nil)
	})
	b.fadeAnim = anim
	anim.Play()
}
