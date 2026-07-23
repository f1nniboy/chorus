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

const playerRowArtSize = 40

type playerRow struct {
	box     *gtk.ListBoxRow
	art     *gtk.Picture
	title   *gtk.Label
	artist  *gtk.Label
	artURL  string
	busName string
}

type Picker struct {
	*gtk.MenuButton
	popover   *gtk.Popover
	listBox   *gtk.ListBox
	resolver  *art.Resolver
	players   map[string]mpris.Player
	rowsByBus map[string]*playerRow
	onSelect  func(info mpris.Player)
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
		players:    map[string]mpris.Player{},
		rowsByBus:  map[string]*playerRow{},
	}

	listBox.ConnectRowActivated(func(activated *gtk.ListBoxRow) {
		for _, row := range pp.rowsByBus {
			if row.box.Object.Native() != activated.Object.Native() {
				continue
			}
			p, ok := pp.players[row.busName]
			if !ok {
				return
			}
			pp.popover.Popdown()
			if pp.onSelect != nil {
				pp.onSelect(p)
			}
			return
		}
	})

	return pp
}

func (pp *Picker) Popover() *gtk.Popover { return pp.popover }

func (pp *Picker) OnSelect(f func(info mpris.Player)) {
	pp.onSelect = f
}

func (pp *Picker) SetPlayers(players []mpris.Player) {
	pp.players = make(map[string]mpris.Player, len(players))
	for _, p := range players {
		pp.players[p.BusName] = p
	}

	for busName := range pp.rowsByBus {
		if _, ok := pp.players[busName]; !ok {
			pp.removeRow(busName)
		}
	}

	pp.MenuButton.SetSensitive(len(players) > 0)
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

func (pp *Picker) SetPlayerTrack(busName string, track mpris.Track) {
	if !track.Valid() {
		pp.removeRow(busName)
		return
	}

	row, exists := pp.rowsByBus[busName]
	if !exists {
		info, ok := pp.players[busName]
		if !ok {
			return
		}
		row = pp.buildRow(info)
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
	if track.ArtURL == "" {
		return
	}

	go func() {
		raw, err := pp.resolver.Load(context.Background(), track.ArtURL)
		if err != nil || raw == nil {
			return
		}
		texture, err := art.Thumbnail(raw, playerRowArtSize)
		if err != nil {
			return
		}
		glib.IdleAdd(func() {
			if row.artURL != track.ArtURL {
				return
			}
			row.art.SetPaintable(texture)
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
	pic.SetSizeRequest(playerRowArtSize, playerRowArtSize)
	pic.SetOverflow(gtk.OverflowHidden)
	pic.AddCSSClass("player-row-art")

	title := gtk.NewLabel(p.Identity)
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
	content.Append(pic)
	content.Append(text)

	row := gtk.NewListBoxRow()
	row.AddCSSClass("player-row")
	row.SetChild(content)
	row.SetActivatable(true)

	pp.listBox.Append(row)

	return &playerRow{box: row, art: pic, title: title, artist: artist, busName: p.BusName}
}
