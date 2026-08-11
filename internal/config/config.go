package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gio/v2"

	"github.com/f1nniboy/chorus/internal/meta"
)

type Config struct {
	*gio.Settings
}

type Appearance struct {
	Size              float64
	Blur              float64
	Padding           float64
	BackgroundBlur    float64
	BackgroundOpacity float64
}

func New() (*Config, error) {
	source := gio.SettingsSchemaSourceGetDefault()
	if source == nil {
		return nil, errors.New("config: no GSettings schema source available")
	}
	schema := source.Lookup(meta.AppID, true)
	if schema == nil {
		return nil, fmt.Errorf("config: GSettings schema %q not found; run glib-compile-schemas data/", meta.AppID)
	}
	return &Config{Settings: gio.NewSettings(meta.AppID)}, nil
}

func (c *Config) ProviderName() string {
	return c.String("provider")
}

func (c *Config) SetProviderName(name string) {
	c.SetString("provider", name)
}

func (c *Config) Appearance() Appearance {
	return Appearance{
		Size:              c.Double("lyrics-size"),
		Blur:              c.Double("lyrics-blur"),
		Padding:           c.Double("lyrics-padding"),
		BackgroundBlur:    c.Double("background-blur"),
		BackgroundOpacity: c.Double("background-opacity"),
	}
}

func (c *Config) ProviderConfig() map[string]any {
	var cfg map[string]any
	json.Unmarshal([]byte(c.String("provider-config")), &cfg) //nolint:errcheck // zero value is fine
	return cfg
}

func (c *Config) SetProviderConfig(cfg map[string]any) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	c.SetString("provider-config", string(data))
}
