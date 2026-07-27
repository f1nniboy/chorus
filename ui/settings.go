package ui

import (
	"fmt"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/cache"
	"github.com/f1nniboy/chorus/internal/config"
	"github.com/f1nniboy/chorus/internal/locale"
	"github.com/f1nniboy/chorus/internal/providers"
)

type Settings struct {
	dialog        *adw.PreferencesDialog
	cfg           *config.Config
	diskCache     *cache.Cache
	providerGroup *adw.PreferencesGroup
	sizeRow       *adw.ActionRow
	onChanged     func()
	configWidgets []gtk.Widgetter
	providerIDs   []string
	dirty         bool
}

// row types (EntryRow, SpinRow, SwitchRow) that support a suffix widget
type configRow interface {
	gtk.Widgetter
	AddSuffix(widget gtk.Widgetter)
}

func NewSettings(cfg *config.Config, diskCache *cache.Cache, onChanged func()) *Settings {
	s := &Settings{cfg: cfg, diskCache: diskCache, onChanged: onChanged}
	s.build()
	return s
}

func (s *Settings) Present(parent gtk.Widgetter) {
	go s.refreshCacheSize(s.sizeRow)
	s.dialog.Present(parent)
}

func (s *Settings) build() {
	s.dialog = adw.NewPreferencesDialog()
	s.dialog.SetTitle(locale.Get("Settings"))
	s.dialog.ConnectClosed(func() {
		if s.dirty {
			s.dirty = false
			s.onChanged()
		}
	})

	page := adw.NewPreferencesPage()
	s.dialog.Add(page)

	providerGroup := adw.NewPreferencesGroup()
	providerGroup.SetTitle(locale.Get("Provider"))
	page.Add(providerGroup)

	var labels []string
	for _, p := range providers.Available() {
		s.providerIDs = append(s.providerIDs, p.ID)
		labels = append(labels, p.Name)
	}

	combo := adw.NewComboRow()
	combo.SetTitle(locale.Get("Provider"))
	combo.SetModel(gtk.NewStringList(labels))

	current := s.cfg.ProviderName()
	for i, id := range s.providerIDs {
		if id == current {
			combo.SetSelected(uint(i))
			break
		}
	}

	providerGroup.Add(combo)
	s.providerGroup = providerGroup

	s.renderConfig()

	cacheGroup := adw.NewPreferencesGroup()
	cacheGroup.SetTitle(locale.Get("Cache"))
	page.Add(cacheGroup)

	sizeRow := adw.NewActionRow()
	sizeRow.SetTitle(locale.Get("Disk usage"))
	cacheGroup.Add(sizeRow)
	s.sizeRow = sizeRow

	clearButton := gtk.NewButton()
	clearButton.SetIconName("user-trash-symbolic")
	clearButton.AddCSSClass("circular")
	clearButton.AddCSSClass("destructive-action")
	clearButton.SetTooltipText(locale.Get("Clear cache"))
	clearButton.SetVAlign(gtk.AlignCenter)
	clearButton.SetSizeRequest(32, 32)
	sizeRow.AddSuffix(clearButton)

	clearButton.ConnectClicked(func() {
		clearButton.SetSensitive(false)
		go func() {
			s.diskCache.Clear()
			s.refreshCacheSize(sizeRow)
			glib.IdleAdd(func() { clearButton.SetSensitive(true) })
		}()
	})
}

func (s *Settings) renderConfig() {
	for _, w := range s.configWidgets {
		s.providerGroup.Remove(w)
	}
	s.configWidgets = s.configWidgets[:0]

	name := s.cfg.ProviderName()
	fields := providers.Fields(name)
	if len(fields) == 0 {
		return
	}

	cfg := s.cfg.ProviderConfig(name)

	for _, f := range fields {
		val := any(nil)
		if cfg != nil {
			val = cfg[f.Key]
		}
		if val == nil {
			val = f.Default
		}

		if f.Type == reflect.String {
			s.addStringRow(f, val)
		}
	}
}

func (s *Settings) addConfigRow(row configRow, key string, isDefault func() bool) *gtk.Button {
	btn := gtk.NewButton()
	btn.SetIconName("edit-undo-symbolic")
	btn.AddCSSClass("flat")
	btn.SetTooltipText(locale.Get("Reset to default"))
	btn.SetVAlign(gtk.AlignCenter)
	btn.SetVisible(!isDefault())
	btn.ConnectClicked(func() {
		s.resetField(key)
	})
	row.AddSuffix(btn)

	s.providerGroup.Add(row)
	s.configWidgets = append(s.configWidgets, row)
	return btn
}

func (s *Settings) addStringRow(f providers.ConfigField, val any) {
	row := adw.NewEntryRow()
	row.SetTitle(locale.Get(f.Label))
	if v, ok := val.(string); ok {
		row.SetText(v)
	}

	def, _ := f.Default.(string)
	isDefault := func() bool { return row.Text() == def }

	btn := s.addConfigRow(row, f.Key, isDefault)
	row.ConnectChanged(func() {
		s.saveField(f.Key, row.Text())
		btn.SetVisible(!isDefault())
	})
}

func (s *Settings) saveField(key string, val any) {
	providerID := s.cfg.ProviderName()
	cfg := s.cfg.ProviderConfig(providerID)
	if cfg == nil {
		cfg = make(map[string]any)
	}
	cfg[key] = val
	s.cfg.SetProviderConfig(providerID, cfg)
	s.dirty = true
}

func (s *Settings) resetField(key string) {
	providerID := s.cfg.ProviderName()
	cfg := s.cfg.ProviderConfig(providerID)
	delete(cfg, key)
	s.cfg.SetProviderConfig(providerID, cfg)
	s.dirty = true
	s.renderConfig()
}

func (s *Settings) refreshCacheSize(row *adw.ActionRow) {
	size, err := s.diskCache.Size()
	glib.IdleAdd(func() {
		if err != nil {
			return
		}
		row.SetSubtitle(formatSize(size))
	})
}

func formatSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
