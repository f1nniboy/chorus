package settings

import (
	"math"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/config"
	"github.com/f1nniboy/chorus/internal/locale"
)

const scaleWidthPx = 200

type Settings struct {
	dialog        *adw.PreferencesDialog
	cfg           *config.Config
	configGroup   *adw.PreferencesGroup
	configWidgets []gtk.Widgetter
	providerIDs   []string
}

func New(cfg *config.Config) *Settings {
	s := &Settings{cfg: cfg}
	s.build()
	return s
}

func (s *Settings) Present(parent gtk.Widgetter) {
	s.dialog.Present(parent)
}

func (s *Settings) build() {
	s.dialog = adw.NewPreferencesDialog()
	s.dialog.SetTitle(locale.Get("Preferences"))

	appearancePage := adw.NewPreferencesPage()
	appearancePage.SetTitle(locale.Get("Appearance"))
	appearancePage.SetIconName("preferences-desktop-appearance-symbolic")
	s.dialog.Add(appearancePage)
	s.buildAppearancePage(appearancePage)

	providerPage := adw.NewPreferencesPage()
	providerPage.SetTitle(locale.Get("Provider"))
	providerPage.SetIconName("network-transmit-receive-symbolic")
	s.dialog.Add(providerPage)
	s.buildProviderPage(providerPage)
}

func (s *Settings) buildAppearancePage(page *adw.PreferencesPage) {
	textGroup := adw.NewPreferencesGroup()
	textGroup.SetTitle(locale.Get("Text"))
	page.Add(textGroup)

	s.addScaleRow(textGroup, locale.Get("Size"), "font-select-symbolic", "lyrics-size", scaleSpec{0.5, 2, 0.05, 1})
	s.addScaleRow(textGroup, locale.Get("Blur"), "weather-fog-symbolic", "lyrics-blur", scaleSpec{0, 1, 0.05, 0.35})
	s.addScaleRow(textGroup, locale.Get("Padding"), "format-justify-fill-symbolic", "lyrics-padding", scaleSpec{0, 1.5, 0.1, 0.5})

	bgGroup := adw.NewPreferencesGroup()
	bgGroup.SetTitle(locale.Get("Background"))
	page.Add(bgGroup)

	s.addScaleRow(bgGroup, locale.Get("Blur"), "weather-fog-symbolic", "background-blur", scaleSpec{0, 1, 0.05, 0.9})
	s.addScaleRow(bgGroup, locale.Get("Opacity"), "view-reveal-symbolic", "background-opacity", scaleSpec{0, 1, 0.1, 1})

	previewGroup := adw.NewPreferencesGroup()
	previewGroup.SetTitle(locale.Get("Preview"))
	previewGroup.Add(s.newLyricsPreview())
	page.Add(previewGroup)
}

type scaleSpec struct {
	min, max, step, def float64
}

func (s *Settings) addScaleRow(group *adw.PreferencesGroup, title, icon, key string, spec scaleSpec) {
	row := adw.NewActionRow()
	row.SetTitle(title)
	row.AddPrefix(gtk.NewImageFromIconName(icon))

	scale := gtk.NewScaleWithRange(gtk.OrientationHorizontal, spec.min, spec.max, spec.step)
	scale.SetSizeRequest(scaleWidthPx, -1)
	scale.AddMark(spec.def, gtk.PosBottom, "")
	scale.ConnectChangeValue(func(_ gtk.ScrollType, value float64) bool {
		scale.SetValue(spec.min + math.Round((value-spec.min)/spec.step)*spec.step)
		return true
	})
	row.AddSuffix(scale)

	s.cfg.Bind(key, scale.Adjustment().Object, "value", gio.SettingsBindDefault)
	group.Add(row)
}
