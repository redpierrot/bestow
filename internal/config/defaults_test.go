/*
All Rights Reversed (ɔ)
*/

package config

import (
	_ "embed"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults_FromTemplate(t *testing.T) {
	tests := []struct {
		name                string
		source, destination string
		want                func(t *testing.T, src, dest string) string
		wantErr             bool
		wantErrIs           error
	}{
		{
			name:        "no errors",
			source:      "/path/to/source",
			destination: "/path/to/dest",
			want: func(t *testing.T, src, dest string) string {
				return getConfigFileAsString(t, src, dest)
			},
		},
		{
			name:        "empty destination",
			source:      "/path/to/source",
			destination: "",
			want: func(t *testing.T, src, dest string) string {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatal(err)
				}
				return getConfigFileAsString(t, src, home)
			},
		},
		{
			name:        "expand paths",
			source:      "~/source",
			destination: "~/dest",
			want: func(t *testing.T, src, dest string) string {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatal(err)
				}
				fullSrc := filepath.Join(home, "source")
				fullDest := filepath.Join(home, "dest")
				return getConfigFileAsString(t, fullSrc, fullDest)
			},
		},
		{
			name:        "clean paths",
			source:      "~/source//",
			destination: "~//dest",
			want: func(t *testing.T, src, dest string) string {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatal(err)
				}
				fullSrc := filepath.Join(home, "source")
				fullDest := filepath.Join(home, "dest")
				return getConfigFileAsString(t, fullSrc, fullDest)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := FromTemplate(tc.source, tc.destination)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			want := tc.want(t, tc.source, tc.destination)
			if config != want {
				t.Fatalf("got %s, want %s", config, want)
			}
		})
	}
}

func TestDefaults_setDefaultDestination(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
		want    func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "no errors",
			profile: &Profile{
				Source: "/path/to/source",
			},
			want: func(t *testing.T) string {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatal(err)
				}
				return home
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Ignore error since there's no point of checking homedir error
			_ = setDefaultDestination(tc.profile, slog.New(slog.NewTextHandler(io.Discard, nil)))
			want := tc.want(t)
			if want != tc.profile.Destination {
				t.Fatalf("got %s, want %s", tc.profile.Destination, want)
			}
		})
	}
}
