package settings

import (
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/locale"
	"github.com/f1nniboy/chorus/internal/providers"
)

func (s *Settings) buildProviderPage(page *adw.PreferencesPage) {
	providerGroup := adw.NewPreferencesGroup()
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

	combo.NotifyProperty("selected", func() {
		id := s.providerIDs[combo.Selected()]
		if id == s.cfg.ProviderName() {
			return
		}
		s.cfg.SetProviderName(id)
		s.cfg.SetProviderConfig(nil)
		s.renderConfig()
	})

	providerGroup.Add(combo)

	s.configGroup = adw.NewPreferencesGroup()
	page.Add(s.configGroup)

	s.renderConfig()
}

func (s *Settings) renderConfig() {
	for _, w := range s.configWidgets {
		s.configGroup.Remove(w)
	}
	s.configWidgets = s.configWidgets[:0]

	name := s.cfg.ProviderName()
	fields := providers.Fields(name)
	s.configGroup.SetVisible(len(fields) > 0)
	if len(fields) == 0 {
		return
	}

	s.configGroup.SetTitle(locale.Get("Configuration"))
	cfg := s.cfg.ProviderConfig()

	for _, f := range fields {
		var val any
		if cfg != nil {
			val = cfg[f.Key]
		}

		if f.Type == reflect.String {
			s.addStringRow(f, val)
		}
	}
}

func (s *Settings) addStringRow(f providers.ConfigField, val any) {
	def, _ := f.Default.(string)

	row := adw.NewEntryRow()
	row.SetTitle(locale.Get(f.Label))
	row.SetShowApplyButton(true)

	current := def
	if v, ok := val.(string); ok && v != "" {
		current = v
	}
	row.SetText(current)

	s.configGroup.Add(row)
	s.configWidgets = append(s.configWidgets, row)

	row.ConnectApply(func() {
		switch text := row.Text(); text {
		case "":
			s.clearField(f.Key)
			if root := row.Root(); root != nil {
				root.SetFocus(nil)
			}
			row.SetText(def)
		case def:
			s.clearField(f.Key)
		default:
			s.saveField(f.Key, text)
		}
	})
}

func (s *Settings) saveField(key string, val any) {
	cfg := s.cfg.ProviderConfig()
	if cfg == nil {
		cfg = make(map[string]any)
	}
	cfg[key] = val
	s.cfg.SetProviderConfig(cfg)
}

func (s *Settings) clearField(key string) {
	cfg := s.cfg.ProviderConfig()
	delete(cfg, key)
	s.cfg.SetProviderConfig(cfg)
}
