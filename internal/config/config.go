/*
All Rights Reversed (ɔ)
*/

// Package config is the layer for handling configs in bestow.
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
	ProfileKey       = "profile"
	profilesKey      = "profiles"
	DefaultProfile   = "default"
)

// AppConfigHome returns the directory where the bestow configs are stored
func AppConfigHome() string {
	return filepath.Join(XDGConfigHome(), appName)
}

// XDGConfigHome returns the root directory of the configs.
// NOTE: on macOS, if the `XDG_CONFIG_HOME` env. is not set, it defaults to `/Library/Application Support/`.
// This bypasses that and returns the `~/.config` if the `XDG_CONFIG_HOME` is not set
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

// GetProfile returns a new Config for a given profile
// The passed viper instance is scoped, therefore, no need to filter from the profile
func GetProfile(profileName string, v *viper.Viper, l *slog.Logger) (*Profile, error) {
	var profile Profile
	if err := v.Unmarshal(&profile); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if err := setDefaultDestination(&profile, l); err != nil {
		return nil, err
	}
	l.Debug("config loaded successfully", "profile", profileName, "configs", profile)
	return &profile, nil
}

// ProfileConfig returns a viper instance scoped to the given profile name.
func ProfileConfig(v *viper.Viper, profileName string) (*viper.Viper, error) {
	sub := v.Sub(profilesKey + "." + profileName)
	if sub == nil {
		return nil, fmt.Errorf("profile %s: %w", profileName, ErrNotFound)
	}
	return sub, nil
}
