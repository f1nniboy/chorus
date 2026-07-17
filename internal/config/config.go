package config

import (
	"encoding/json"
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

const SchemaID = "space.f1nn.chorus"

type Config struct {
	*gio.Settings
}

func New() (*Config, error) {
	source := gio.SettingsSchemaSourceGetDefault()
	if source == nil {
		return nil, fmt.Errorf("config: no GSettings schema source available")
	}
	schema := source.Lookup(SchemaID, true)
	if schema == nil {
		return nil, fmt.Errorf("config: GSettings schema %q not found; run glib-compile-schemas data/", SchemaID)
	}
	return &Config{Settings: gio.NewSettings(SchemaID)}, nil
}

func (c *Config) LastPlayerIdentity() string {
	return c.String("last-player-identity")
}

func (c *Config) SetLastPlayerIdentity(identity string) {
	c.SetString("last-player-identity", identity)
}

func (c *Config) WindowSize() (width, height int) {
	return int(c.Int("window-width")), int(c.Int("window-height"))
}

func (c *Config) SetWindowSize(width, height int) {
	c.SetInt("window-width", width)
	c.SetInt("window-height", height)
}

func (c *Config) ProviderName() string {
	return c.String("provider")
}

func (c *Config) allProviderConfigs() map[string]map[string]any {
	var all map[string]map[string]any
	json.Unmarshal([]byte(c.String("provider-config")), &all) //nolint:errcheck // zero value is fine
	return all
}

func (c *Config) ProviderConfig(providerID string) map[string]any {
	return c.allProviderConfigs()[providerID]
}

func (c *Config) SetProviderConfig(providerID string, cfg map[string]any) {
	all := c.allProviderConfigs()
	if all == nil {
		all = make(map[string]map[string]any)
	}
	all[providerID] = cfg
	data, err := json.Marshal(all)
	if err != nil {
		return
	}
	c.SetString("provider-config", string(data))
}
