package header

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/art"
	"github.com/f1nniboy/chorus/internal/locale"
	"github.com/f1nniboy/chorus/ui/picker"
)

type Header struct {
	Revealer *gtk.Revealer
	Bar      *adw.HeaderBar
	Picker   *picker.Picker
}

func New(artResolver *art.Resolver) *Header {
	bar := adw.NewHeaderBar()
	bar.SetShowTitle(false)
	bar.AddCSSClass("flat")

	p := picker.New(artResolver)
	bar.PackStart(p)

	menu := gio.NewMenu()
	menu.Append(locale.Get("Preferences"), "app.preferences")
	menu.Append(locale.Get("About"), "app.about")

	menuButton := gtk.NewMenuButton()
	menuButton.SetIconName("open-menu-symbolic")
	menuButton.SetMenuModel(menu)
	bar.PackEnd(menuButton)

	revealer := gtk.NewRevealer()
	revealer.SetTransitionType(gtk.RevealerTransitionTypeCrossfade)
	revealer.SetChild(bar)
	revealer.SetRevealChild(false)
	revealer.SetHAlign(gtk.AlignFill)
	revealer.SetVAlign(gtk.AlignStart)

	return &Header{Revealer: revealer, Bar: bar, Picker: p}
}

func (h *Header) SetRevealed(revealed bool) {
	h.Revealer.SetRevealChild(revealed)
}
