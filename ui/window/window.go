package window

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/art"
	"github.com/f1nniboy/chorus/internal/config"
	"github.com/f1nniboy/chorus/internal/meta"
	"github.com/f1nniboy/chorus/ui/appearance"
	"github.com/f1nniboy/chorus/ui/background"
	"github.com/f1nniboy/chorus/ui/header"
	"github.com/f1nniboy/chorus/ui/lyricsview"
)

type Window struct {
	*adw.ApplicationWindow

	cfg        *config.Config
	Background *background.Background
	Lyrics     *lyricsview.View
	Header     *header.Header
}

func New(app *adw.Application, cfg *config.Config, resolver *art.Resolver) *Window {
	win := &Window{
		ApplicationWindow: adw.NewApplicationWindow(&app.Application),
		Background:        background.New(resolver),
		Lyrics:            lyricsview.New(),
		Header:            header.New(resolver),
		cfg:               cfg,
	}

	win.SetTitle(meta.AppName)
	win.AddCSSClass("app-window")

	obj := coreglib.BaseObject(win)
	cfg.Bind("window-width", obj, "default-width", gio.SettingsBindDefault)
	cfg.Bind("window-height", obj, "default-height", gio.SettingsBindDefault)
	cfg.Bind("window-maximized", obj, "maximized", gio.SettingsBindDefault)

	sm := adw.StyleManagerGetDefault()
	win.updateTheme(sm.Dark())
	sm.NotifyProperty("dark", func() {
		win.updateTheme(sm.Dark())
	})

	win.Lyrics.SetHExpand(true)
	win.Lyrics.SetVExpand(true)

	overlay := gtk.NewOverlay()
	overlay.SetChild(win.Background)
	overlay.AddOverlay(win.Lyrics)
	overlay.AddOverlay(win.Header.Revealer)

	handle := gtk.NewWindowHandle()
	handle.SetChild(overlay)

	win.SetContent(handle)

	win.NotifyProperty("is-active", func() {
		win.Header.SetRevealed(win.IsActive())
	})

	win.applyAppearance()
	cfg.ConnectChanged(func(key string) {
		switch key {
		case "lyrics-size", "lyrics-blur", "lyrics-padding", "background-blur", "background-opacity":
			win.applyAppearance()
		}
	})

	return win
}

func (win *Window) applyAppearance() {
	a := win.cfg.Appearance()
	appearance.Apply(a)
	win.Background.SetBlur(a.BackgroundBlur)
	// TODO: set blur (and cover art) for preview?
}

func (win *Window) updateTheme(dark bool) {
	if dark {
		win.AddCSSClass("dark")
	} else {
		win.RemoveCSSClass("dark")
	}
}
