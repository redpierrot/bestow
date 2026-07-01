/*
All Rights Reversed (t *testing.T)
*/

package engine

import (
	"errors"
	"os"
	"slices"
	"testing"
)

func TestOperations_validateDestinations(t *testing.T) {
	cand := func(src, dest string) operationCandidate {
		return operationCandidate{
			source:      src,
			destination: dest,
		}
	}
	tests := []struct {
		name       string
		setup      func() *Engine
		candidates []operationCandidate
		wantErrAs  func(*testing.T, error)
	}{
		{
			name: "validate destinations",
			setup: func() *Engine {
				mf := &MockFileSystem{}
				return newTestEngine("", "", mf, nil)
			},
			candidates: []operationCandidate{
				cand("dotfiles/nvim/.config/nvim/init.lua", "home/.config/nvim/init.lua"),
				cand("dotfiles/nvim/.config/nvim/plugins.lua", "home/.config/nvim/plugins.lua"),
				cand("dotfiles/bestow/.config/bestow/config.yaml", "home/.config/bestow/config.yaml"),
			},
		},
		{
			name: "validate destinations - conflicts",
			setup: func() *Engine {
				mf := &MockFileSystem{}
				return newTestEngine("", "", mf, nil)
			},
			candidates: []operationCandidate{
				cand("dotfiles/nvim/init.lua", "home/.config/init.lua"),
				cand("dotfiles/yazi/config.yaml", "home/.config/config.yaml"),
				cand("dotfiles/bestow/config.yaml", "home/.config/config.yaml"),
			},
			wantErrAs: func(t *testing.T, err error) {
				var conflictError *ConflictError
				if !errors.As(err, &conflictError) {
					t.Fatalf("got %v, want ConflictError", err)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			err := e.validateDestinations(tc.candidates)
			if tc.wantErrAs != nil {
				tc.wantErrAs(t, err)
			}
		})
	}
}

func TestOperations_buildOperationCandidates(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		args      []string
		want      []string
		wantErr   bool
		wantErrIs error
	}{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {})
	}
}

func TestOperations_buildFileActions(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		args      []string
		want      []string
		wantErr   bool
		wantErrIs error
	}{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {})
	}
}

func TestOperations_stowFileAction(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		args      []string
		want      []string
		wantErr   bool
		wantErrIs error
	}{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {})
	}
}

func TestOperations_unstowFileAction(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		args      []string
		want      []string
		wantErr   bool
		wantErrIs error
	}{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {})
	}
}

func TestOperations_calculateBackupPath(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(path string, existingFiles []string) *Engine
		existingFiles []string
		want          string
		wantErr       bool
		wantErrIs     error
		wantErrAs     func(*testing.T, error)
	}{
		{
			name: "calculate backup path",
			setup: func(path string, _ []string) *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			want: "dest_file.0.bestow.backup",
		},
		{
			name: "calculate backup path - existing backed up files",
			setup: func(path string, existingFiles []string) *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						if slices.Contains(existingFiles, path) {
							return true, nil
						}
						return false, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			existingFiles: []string{"dest_file.0.bestow.backup", "dest_file.1.bestow.backup", "dest_file.2.bestow.backup"},
			want:          "dest_file.3.bestow.backup",
		},
		{
			name: "calculate backup path -  existing return err",
			setup: func(path string, existingFiles []string) *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, os.ErrPermission
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			existingFiles: []string{"dest_file.0.bestow.backup", "dest_file.1.bestow.backup", "dest_file.2.bestow.backup"},
			wantErr:       true,
			wantErrIs:     os.ErrPermission,
		},
		{
			name: "calculate backup path - existing backed up files",
			setup: func(path string, existingFiles []string) *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						if slices.Contains(existingFiles, path) {
							return true, nil
						}
						return false, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			existingFiles: []string{
				"dest_file.0.bestow.backup",
				"dest_file.1.bestow.backup",
				"dest_file.2.bestow.backup",
				"dest_file.3.bestow.backup",
				"dest_file.4.bestow.backup",
				"dest_file.5.bestow.backup",
				"dest_file.6.bestow.backup",
			},
			wantErrAs: func(t *testing.T, err error) {
				var hintedErr *HintedError
				if !errors.As(err, &hintedErr) {
					t.Fatalf("got %v, want HintedError", err)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := "dest_file"
			eng := tc.setup(path, tc.existingFiles)
			backupPath, err := eng.calculateBackupPath(path)
			if tc.wantErrAs != nil {
				tc.wantErrAs(t, err)
				return
			}
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if backupPath != tc.want {
				t.Fatalf("got %v, want %v", backupPath, tc.want)
			}
		})
	}
}

func TestOperations_calculateStowActionByStrategy(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		args      []string
		want      []string
		wantErr   bool
		wantErrIs error
	}{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {})
	}
}
