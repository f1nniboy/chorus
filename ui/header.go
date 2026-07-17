package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/art"
)

type Header struct {
	Revealer *gtk.Revealer

	Bar    *adw.HeaderBar
	Picker *Picker
}

func NewHeader(artResolver *art.Resolver) *Header {
	bar := adw.NewHeaderBar()
	bar.SetShowTitle(false)
	bar.SetShowStartTitleButtons(false)
	bar.SetShowEndTitleButtons(true)
	bar.AddCSSClass("flat")

	picker := NewPicker(artResolver)
	bar.PackStart(picker)

	menu := gio.NewMenu()
	menu.Append("Settings", "app.settings")
	menu.Append("About", "app.about")

	menuButton := gtk.NewMenuButton()
	menuButton.SetIconName("open-menu-symbolic")
	menuButton.SetMenuModel(menu)
	bar.PackEnd(menuButton)

	revealer := gtk.NewRevealer()
	revealer.SetTransitionType(gtk.RevealerTransitionTypeCrossfade)
	revealer.SetChild(bar)
	revealer.SetRevealChild(false)

	return &Header{Revealer: revealer, Bar: bar, Picker: picker}
}

func (h *Header) SetRevealed(revealed bool) {
	h.Revealer.SetRevealChild(revealed)
}
