package ui

import (
	"context"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"github.com/f1nniboy/chorus/internal/art"
	"github.com/f1nniboy/chorus/internal/locale"
	"github.com/f1nniboy/chorus/internal/mpris"
)

const rowArtSize = 40

type playerRow struct {
	box    *gtk.ListBoxRow
	stack  *gtk.Stack
	art    *gtk.Picture
	title  *gtk.Label
	artist *gtk.Label
	artURL string
	player mpris.Player
}

type Picker struct {
	*gtk.MenuButton
	popover   *gtk.Popover
	listBox   *gtk.ListBox
	resolver  *art.Resolver
	rowsByBus map[string]*playerRow
	onSelect  func(busName string)
	current   string
}

func NewPicker(resolver *art.Resolver) *Picker {
	button := gtk.NewMenuButton()
	button.SetIconName("multimedia-player-symbolic")
	button.SetTooltipText(locale.Get("Choose player"))
	button.SetSensitive(false)

	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.SetActivateOnSingleClick(true)

	popover := gtk.NewPopover()
	popover.SetChild(listBox)
	button.SetPopover(popover)

	pp := &Picker{
		MenuButton: button,
		popover:    popover,
		listBox:    listBox,
		resolver:   resolver,
		rowsByBus:  map[string]*playerRow{},
	}

	listBox.ConnectRowActivated(func(activated *gtk.ListBoxRow) {
		for _, row := range pp.rowsByBus {
			if row.box.Object.Native() != activated.Object.Native() {
				continue
			}
			pp.popover.Popdown()
			if pp.onSelect != nil {
				pp.onSelect(row.player.BusName)
			}
			return
		}
	})

	return pp
}

func (pp *Picker) Popover() *gtk.Popover { return pp.popover }

func (pp *Picker) OnSelect(f func(busName string)) {
	pp.onSelect = f
}

func (pp *Picker) SetRoster(entries []mpris.Entry) {
	pp.MenuButton.SetSensitive(len(entries) > 0)

	seen := make(map[string]bool, len(entries))
	for _, e := range entries {
		seen[e.Player.BusName] = true
		pp.upsertRow(e)
	}

	for busName := range pp.rowsByBus {
		if !seen[busName] {
			pp.removeRow(busName)
		}
	}
}

func (pp *Picker) SetCurrent(busName string) {
	pp.current = busName
	for bus, row := range pp.rowsByBus {
		if bus == pp.current {
			row.title.AddCSSClass("current")
		} else {
			row.title.RemoveCSSClass("current")
		}
	}
}

func (pp *Picker) upsertRow(e mpris.Entry) {
	busName := e.Player.BusName
	track := e.Track

	row, exists := pp.rowsByBus[busName]
	if !exists {
		row = pp.buildRow(e.Player)
		pp.rowsByBus[busName] = row
		if busName == pp.current {
			row.title.AddCSSClass("current")
		}
	}

	if track.Title != "" {
		row.title.SetLabel(track.Title)
	}
	row.artist.SetLabel(track.Artist)

	if track.ArtURL == row.artURL {
		return
	}
	row.artURL = track.ArtURL
	row.art.SetPaintable(nil)
	row.stack.SetVisibleChildName("placeholder")
	if track.ArtURL == "" {
		return
	}

	go func() {
		raw, err := pp.resolver.Load(context.Background(), track.ArtURL)
		if err != nil || raw == nil {
			return
		}
		texture, err := art.Thumbnail(raw, rowArtSize)
		if err != nil {
			return
		}
		glib.IdleAdd(func() {
			if row.artURL != track.ArtURL {
				return
			}
			row.art.SetPaintable(texture)
			row.stack.SetVisibleChildName("art")
		})
	}()
}

func (pp *Picker) removeRow(busName string) {
	row, ok := pp.rowsByBus[busName]
	if !ok {
		return
	}
	pp.listBox.Remove(row.box)
	delete(pp.rowsByBus, busName)
}

func (pp *Picker) buildRow(p mpris.Player) *playerRow {
	pic := gtk.NewPicture()
	pic.SetContentFit(gtk.ContentFitCover)
	pic.SetCanShrink(true)

	placeholder := gtk.NewImage()
	placeholder.SetFromIconName("folder-music-symbolic")
	placeholder.SetPixelSize(rowArtSize / 2)
	placeholder.AddCSSClass("dim-label")

	stack := gtk.NewStack()
	stack.SetSizeRequest(rowArtSize, rowArtSize)
	stack.SetOverflow(gtk.OverflowHidden)
	stack.AddCSSClass("player-row-art")
	stack.AddNamed(pic, "art")
	stack.AddNamed(placeholder, "placeholder")
	stack.SetVisibleChildName("placeholder")

	title := gtk.NewLabel("")
	title.SetXAlign(0)
	title.SetEllipsize(pango.EllipsizeEnd)
	title.AddCSSClass("player-row-title")

	artist := gtk.NewLabel("")
	artist.SetXAlign(0)
	artist.SetEllipsize(pango.EllipsizeEnd)
	artist.AddCSSClass("player-row-artist")

	text := gtk.NewBox(gtk.OrientationVertical, 2)
	text.SetVAlign(gtk.AlignCenter)
	text.Append(title)
	text.Append(artist)

	content := gtk.NewBox(gtk.OrientationHorizontal, 10)
	content.SetMarginTop(8)
	content.SetMarginBottom(8)
	content.SetMarginStart(10)
	content.SetMarginEnd(16)
	content.Append(stack)
	content.Append(text)

	row := gtk.NewListBoxRow()
	row.AddCSSClass("player-row")
	row.SetChild(content)
	row.SetActivatable(true)

	pp.listBox.Append(row)

	return &playerRow{box: row, stack: stack, art: pic, title: title, artist: artist, player: p}
}
