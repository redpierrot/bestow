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

func TestConfig_GetProfile(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		profileName string
		wantProfile func(t *testing.T) *Profile
		wantErr     error
	}{
		{
			name:        "default profile",
			profileName: "default",
			yaml: `
profiles:
  default:
    source: /home/ru/dotfiles/
    destination: /home/ru/
`,
			wantProfile: func(t *testing.T) *Profile {
				return &Profile{
					Source:      "/home/ru/dotfiles/",
					Destination: "/home/ru/",
				}
			},
		},
		{
			name:        "profile without setting destination",
			profileName: "default",
			yaml: `
profiles:
  default:
    source: /home/ru/dotfiles/
`,
			wantProfile: func(t *testing.T) *Profile {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatal(err)
				}
				return &Profile{
					Source:      "/home/ru/dotfiles/",
					Destination: home,
				}
			},
		},
		{
			name:        "pick named profile from multiple profiles",
			profileName: "sandbox",
			yaml: `
profiles:
  default:
    source: /home/ru/dotfiles/
    destination: /home/ru/
  sandbox:
    source: /home/thisaru/sandbox/
    destination: /home/thisaru/fakehome/
`,
			wantProfile: func(t *testing.T) *Profile {
				return &Profile{
					Source:      "/home/thisaru/sandbox/",
					Destination: "/home/thisaru/fakehome/",
				}
			},
		},
		{
			name:        "non-existing profile",
			profileName: "noProfile",
			yaml: `
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
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(strings.NewReader(tc.yaml)); err != nil {
				t.Fatal(err)
			}
			profileConfig, err := ProfileConfig(v, tc.profileName)
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
				t.Fatalf("got %v, want %v", profileConfig, tc.wantErr)
			}
			want := tc.wantProfile(t)
			gotProfile, err := GetProfile(tc.profileName, profileConfig, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if err != nil {
				t.Fatal(err)
			}
			if *gotProfile != *want {
				t.Fatalf("got %v, want %v", gotProfile, want)
			}
		})
	}
}
