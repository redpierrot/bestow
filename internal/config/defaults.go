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
	"text/template"
)

//go:embed defaults/default-config.yaml
var defaultTemplate string

var DefaultIgnoreList = []string{".git", ".gitignore", "README.md", "LICENSE", "**/.bestowignore", "**/.stow-local-ignore"}

func DefaultTemplate(source, destination string) (string, error) {
	tmpl, err := template.New("config").Parse(defaultTemplate)
	if err != nil {
		return "", err
	}
	if destination == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("parse home dir: %w", err)
		}
		destination = home
	}
	data := struct {
		Source      string
		Destination string
	}{
		Source:      source,
		Destination: destination,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func setDefaultDestination(cfg *Config, l *slog.Logger) error {
	l.Debug("checking destination config")
	if cfg.Destination != "" {
		l.Debug("destination is set by configs", "destination", cfg.Destination)
		return nil
	}
	l.Debug("no destination provided, setting default destination")
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	cfg.Destination = home
	l.Debug("default value is set for destination", "destination", cfg.Destination)
	return nil
}
