/*
All Rights Reversed (ɔ)
*/

package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	ConfigDir        string = ".config"
	EnvXdgConfigHome string = "XDG_CONFIG_HOME"
)

type Profile struct {
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
}

type rawConfig struct {
	Profiles map[string]Profile `mapstructure:"profiles"`
}

type Config struct {
	Source      string
	Destination string
}

type ConfigError struct {
	Message    string
	ConfigName string
	Value      string
	Cause      error
}

func AppConfigHome() string {
	return filepath.Join(XdgConfigHome(), constant.AppName)
}

// XdgConfigHome returns the root directory of the configs.
// NOTE: on macOS, if the `XDG_CONFIG_HOME` env. is not set,
// it defaults to `/Library/Application Support/`.
// This bypasses that and return the `~/.config` if the `XDG_CONFIG_HOME`
// is not set
func XdgConfigHome() string {
	if dir := os.Getenv(EnvXdgConfigHome); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return xdg.ConfigHome
	}
	return filepath.Join(home, ConfigDir)
}

// TODO: Support directory-level config file for different source and destination options
func GetConfig(viper *viper.Viper, l *slog.Logger) (*Config, error) {
	config, err := loadConfig(viper, l)
	if err != nil {
		return nil, &ConfigError{
			Message: "failed to read the configs",
			Cause:   err,
		}
	}
	return config, nil
}

func (e *ConfigError) Error() string {
	msg := e.Message
	if e.ConfigName != "" {
		msg += fmt.Sprintf(" [%s]", e.ConfigName)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	return msg
}

func (e *ConfigError) Unwrap() error { return e.Cause }

func loadConfig(viper *viper.Viper, l *slog.Logger) (*Config, error) {
	l.Debug("loading configs")

	var raw rawConfig
	if err := viper.Unmarshal(&raw); err != nil {
		return nil, &ConfigError{
			Message: "failed to load the configs",
			Cause:   err,
		}
	}
	l.Debug("unmarshaled the configs", "raw", raw)

	profileName := viper.GetString(constant.ProfileKey)
	if profileName == "" {
		profileName = constant.DefaultProfile
	}
	profile, ok := raw.Profiles[profileName]
	if !ok {
		return nil, &ConfigError{
			Message:    "profile not found",
			ConfigName: "profile",
			Value:      profileName,
		}
	}

	cfg := Config{
		Source:      profile.Source,
		Destination: profile.Destination,
	}

	if err := setDefaultSource(&cfg, l); err != nil {
		return nil, err
	}
	if err := setDefaultDestination(&cfg, l); err != nil {
		return nil, err
	}
	l.Debug("config loaded successfully", "cfg", cfg)

	return &cfg, nil
}
