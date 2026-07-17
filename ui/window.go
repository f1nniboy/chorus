package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/art"
	"github.com/f1nniboy/chorus/internal/config"
)

type Window struct {
	*adw.ApplicationWindow

	cfg        *config.Config
	Background *Background
	Lyrics     *LyricsView
	Header     *Header
}

func NewWindow(app *adw.Application, cfg *config.Config, artResolver *art.Resolver) *Window {
	win := &Window{
		ApplicationWindow: adw.NewApplicationWindow(&app.Application),
		cfg:               cfg,
		Background:        NewBackground(artResolver),
		Lyrics:            NewLyricsView(),
		Header:            NewHeader(artResolver),
	}

	win.SetTitle("chorus")
	width, height := cfg.WindowSize()
	win.SetDefaultSize(width, height)
	win.AddCSSClass("chorus-window")

	sm := adw.StyleManagerGetDefault()
	win.updateDarkClass(sm.Dark())
	sm.NotifyProperty("dark", func() {
		win.updateDarkClass(sm.Dark())
	})

	win.Lyrics.SetHExpand(true)
	win.Lyrics.SetVExpand(true)

	win.Header.Revealer.SetHAlign(gtk.AlignFill)
	win.Header.Revealer.SetVAlign(gtk.AlignStart)

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

	win.ConnectCloseRequest(func() bool {
		w, h := win.DefaultSize()
		cfg.SetWindowSize(w, h)
		return false
	})

	return win
}

func (win *Window) updateDarkClass(dark bool) {
	if dark {
		win.AddCSSClass("dark")
	} else {
		win.RemoveCSSClass("dark")
	}
}
