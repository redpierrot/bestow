/*
All Rights Reversed (ɔ)
*/

package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed defaults/default-config.yaml
var defaultTemplate string

const tildePrefix = "~/"

// DefaultIgnoreList stores the most commonly used ignore patterns
var DefaultIgnoreList = []string{".git", ".gitignore", "README.md", "LICENSE", "**/.bestowignore", "**/.stow-local-ignore"}

type Config struct {
	Version  string             `mapstructure:"version"`
	Profiles map[string]Profile `mapstructure:"profiles"`
}

type Profile struct {
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
}

// FromTemplate populates and returns the default template with the provided source and destination
func FromTemplate(src, dest string) (string, error) {
	tmpl, err := template.New("config").Parse(defaultTemplate)
	if err != nil {
		return "", err
	}
	profile, err := getProfile(src, dest)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, profile); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func setDefaultDestination(profile *Profile, l *slog.Logger) error {
	l.Debug("checking destination config")
	if profile.Destination != "" {
		l.Debug("destination is set by configs", "destination", profile.Destination)
		return nil
	}
	l.Debug("no destination provided, setting default destination")
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	profile.Destination = home
	l.Debug("default value is set for destination", "destination", profile.Destination)
	return nil
}

func getProfile(src, dest string) (*Profile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("parse home dir: %w", err)
	}
	if dest == "" {
		dest = home
	}
	src = getAbsPath(src, home)
	dest = getAbsPath(dest, home)
	return &Profile{
		Source:      src,
		Destination: dest,
	}, nil
}

func getAbsPath(path string, home string) string {
	path = os.ExpandEnv(path)
	if !strings.HasPrefix(path, tildePrefix) {
		return path
	}
	path = strings.Replace(path, tildePrefix, "", 1)
	path = filepath.Join(home, path)
	return filepath.Clean(path)
}
