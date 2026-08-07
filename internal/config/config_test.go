/*
All Rights Reversed (ɔ)
*/

package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

func TestConfig_XDGConfigHome(t *testing.T) {
	tests := []struct {
		name     string
		xdgHome  string
		userHome string
		want     string
	}{
		{"xdg home set", "/Users/ru/.config", "/Users/ru", "/Users/ru/.config"},
		{"xdg home not set", "", "/Users/ru", "/Users/ru/.config"},
		{"user home error", "", "", xdg.ConfigHome},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tc.xdgHome)
			t.Setenv("HOME", tc.userHome)
			got := XDGConfigHome()
			if got != tc.want {
				t.Fatalf("got XDG Home %s, want %s", got, tc.want)
			}
		})
	}
}

func TestConfig_NewConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantCfg func(t *testing.T) *Config
		wantErr error
	}{
		{
			name: "profile with source and destination",
			yaml: `
profiles:
  default:
    source: /home/ru/dotfiles/
    destination: /home/ru/
`,
			wantCfg: func(t *testing.T) *Config {
				return &Config{
					Source:      "/home/ru/dotfiles/",
					Destination: "/home/ru/",
				}
			},
		},
		{
			name: "profile without setting destination",
			yaml: `
profiles:
  default:
    source: /home/ru/dotfiles/
`,
			wantCfg: func(t *testing.T) *Config {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatal(err)
				}
				return &Config{
					Source:      "/home/ru/dotfiles/",
					Destination: home,
				}
			},
		},
		{
			name: "pick from multiple profiles",
			yaml: `
profile: sandbox
profiles:
  default:
    source: /home/ru/dotfiles/
    destination: /home/ru/
  sandbox:
    source: /home/thisaru/sandbox/
    destination: /home/thisaru/fakehome/
`,
			wantCfg: func(t *testing.T) *Config {
				return &Config{
					Source:      "/home/thisaru/sandbox/",
					Destination: "/home/thisaru/fakehome/",
				}
			},
		},
		{
			name: "profile not found",
			yaml: `
profile: noProfile
profiles:
  default:
    source: /home/ru/dotfiles/
    destination: /home/ru/
  sandbox:
    source: /home/thisaru/sandbox/
    destination: /home/thisaru/fakehome/
`,
			wantErr: fmt.Errorf("profile noProfile: not found"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := slog.New(slog.NewTextHandler(io.Discard, nil))
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(strings.NewReader(tc.yaml)); err != nil {
				t.Fatal(err)
			}
			got, err := NewConfig(v, l)
			if err != nil {
				if tc.wantErr == nil {
					t.Fatalf("got %v, want %v", err, tc.wantErr)
				}
				if err.Error() != tc.wantErr.Error() {
					t.Fatalf("got %v, want %v", err, tc.wantErr)
				}
				return
			}
			if tc.wantErr != nil {
				t.Fatalf("got %v, want %v", got, tc.wantErr)
			}
			want := tc.wantCfg(t)
			if *got != *want {
				t.Fatalf("got %v, want %v", got, want)
			}
		})
	}
}
