/*
All Rights Reversed (ɔ)
*/

package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	configDir        = ".config"
	envXDGConfigHome = "XDG_CONFIG_HOME"
	appName          = "bestow"
	profileKey       = "profile"
	defaultProfile   = "default"
)

type Profile struct {
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
}

type configFile struct {
	Profiles map[string]Profile `mapstructure:"profiles"`
}

type Config struct {
	Source      string
	Destination string
}

func AppConfigHome() string {
	return filepath.Join(XDGConfigHome(), appName)
}

// XDGConfigHome returns the root directory of the configs.
// NOTE: on macOS, if the `XDG_CONFIG_HOME` env. is not set,
// it defaults to `/Library/Application Support/`.
// This bypasses that and return the `~/.config` if the `XDG_CONFIG_HOME`
// is not set
func XDGConfigHome() string {
	if dir := os.Getenv(envXDGConfigHome); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return xdg.ConfigHome
	}
	return filepath.Join(home, configDir)
}

func NewConfig(v *viper.Viper, l *slog.Logger) (*Config, error) {
	l.Debug("loading configs")

	var raw configFile
	if err := v.Unmarshal(&raw); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	l.Debug("unmarshaled the configs", "raw", raw)

	profileName := v.GetString(profileKey)
	if profileName == "" {
		profileName = defaultProfile
	}
	profile, ok := raw.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile %s: %w", profileName, ErrNotFound)
	}

	cfg := Config(profile)

	if err := setDefaultDestination(&cfg, l); err != nil {
		return nil, err
	}
	l.Debug("config loaded successfully", "cfg", cfg)

	return &cfg, nil
}
