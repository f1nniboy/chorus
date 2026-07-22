package ui

import (
	"fmt"
	"strconv"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/cache"
	"github.com/f1nniboy/chorus/internal/config"
	"github.com/f1nniboy/chorus/internal/providers"
)

type Settings struct {
	dialog        *adw.PreferencesDialog
	cfg           *config.Config
	diskCache     *cache.Cache
	providerGroup *adw.PreferencesGroup
	configWidgets []gtk.Widgetter
	providerIDs   []string
	onChanged     func()
	dirty         bool
}

func NewSettings(cfg *config.Config, diskCache *cache.Cache, onChanged func()) *Settings {
	s := &Settings{cfg: cfg, diskCache: diskCache, onChanged: onChanged}
	s.build()
	return s
}

func (s *Settings) Present(parent gtk.Widgetter) {
	s.dialog.Present(parent)
}

func (s *Settings) build() {
	s.dialog = adw.NewPreferencesDialog()
	s.dialog.SetTitle("Settings")
	s.dialog.ConnectClosed(func() {
		if s.dirty {
			s.dirty = false
			s.onChanged()
		}
	})

	page := adw.NewPreferencesPage()
	s.dialog.Add(page)

	providerGroup := adw.NewPreferencesGroup()
	providerGroup.SetTitle("Provider")
	page.Add(providerGroup)

	var labels []string
	for _, p := range providers.Available() {
		s.providerIDs = append(s.providerIDs, p.ID)
		labels = append(labels, p.Name)
	}

	combo := adw.NewComboRow()
	combo.SetTitle("Provider")
	combo.SetModel(gtk.NewStringList(labels))

	current := s.cfg.ProviderName()
	for i, id := range s.providerIDs {
		if id == current {
			combo.SetSelected(uint(i))
			break
		}
	}

	combo.NotifyProperty("selected", func() {
		idx := combo.Selected()
		if int(idx) < len(s.providerIDs) {
			s.cfg.SetProviderName(s.providerIDs[idx])
			s.renderConfig()
			s.dirty = true
		}
	})

	providerGroup.Add(combo)

	s.providerGroup = providerGroup

	s.renderConfig()

	cacheGroup := adw.NewPreferencesGroup()
	cacheGroup.SetTitle("Cache")
	page.Add(cacheGroup)

	sizeRow := adw.NewActionRow()
	sizeRow.SetTitle("Disk usage")
	cacheGroup.Add(sizeRow)

	clearButton := gtk.NewButton()
	clearButton.SetIconName("user-trash-symbolic")
	clearButton.AddCSSClass("circular")
	clearButton.AddCSSClass("destructive-action")
	clearButton.SetTooltipText("Clear cache")
	clearButton.SetVAlign(gtk.AlignCenter)
	clearButton.SetSizeRequest(34, 34)
	sizeRow.AddSuffix(clearButton)

	clearButton.ConnectClicked(func() {
		clearButton.SetSensitive(false)
		go func() {
			s.diskCache.Clear()
			s.refreshCacheSize(sizeRow)
			glib.IdleAdd(func() { clearButton.SetSensitive(true) })
		}()
	})

	go s.refreshCacheSize(sizeRow)
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

		switch f.Type {
		case "string":
			s.addStringRow(f, val)
		case "int":
			s.addIntRow(f, val)
		case "bool":
			s.addBoolRow(f, val)
		}
	}
}

func (s *Settings) addStringRow(f providers.ConfigField, val any) {
	row := adw.NewEntryRow()
	row.SetTitle(f.Label)
	if v, ok := val.(string); ok {
		row.SetText(v)
	}
	row.ConnectChanged(func() {
		s.saveField(f.Key, row.Text())
	})
	s.providerGroup.Add(row)
	s.configWidgets = append(s.configWidgets, row)
}

func (s *Settings) addIntRow(f providers.ConfigField, val any) {
	row := adw.NewSpinRowWithRange(0, 60000, 1)
	row.SetTitle(f.Label)
	var n float64
	switch v := val.(type) {
	case float64:
		n = v
	case string:
		n, _ = strconv.ParseFloat(v, 64)
	}
	row.SetValue(n)
	row.NotifyProperty("value", func() {
		s.saveField(f.Key, row.Value())
	})
	s.providerGroup.Add(row)
	s.configWidgets = append(s.configWidgets, row)
}

func (s *Settings) addBoolRow(f providers.ConfigField, val any) {
	row := adw.NewSwitchRow()
	row.SetTitle(f.Label)
	switch v := val.(type) {
	case bool:
		row.SetActive(v)
	case string:
		b, _ := strconv.ParseBool(v)
		row.SetActive(b)
	}
	row.NotifyProperty("active", func() {
		s.saveField(f.Key, row.Active())
	})
	s.providerGroup.Add(row)
	s.configWidgets = append(s.configWidgets, row)
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

func (s *Settings) refreshCacheSize(row *adw.ActionRow) {
	size, err := s.diskCache.Size()
	glib.IdleAdd(func() {
		if err != nil {
			row.SetSubtitle("unknown")
			return
		}
		row.SetSubtitle(formatSize(size))
	})
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
