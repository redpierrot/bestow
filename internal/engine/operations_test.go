/*
All Rights Reversed (t *testing.T)
*/

package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/redpierrot/bestow/internal/file"
)

func TestOperations_validateDestinations(t *testing.T) {
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
				candidate("dotfiles/nvim/.config/nvim/init.lua", "home/.config/nvim/init.lua"),
				candidate("dotfiles/nvim/.config/nvim/plugins.lua", "home/.config/nvim/plugins.lua"),
				candidate("dotfiles/bestow/.config/bestow/config.yaml", "home/.config/bestow/config.yaml"),
			},
		},
		{
			name: "validate destinations - conflicts",
			setup: func() *Engine {
				mf := &MockFileSystem{}
				return newTestEngine("", "", mf, nil)
			},
			candidates: []operationCandidate{
				candidate("dotfiles/nvim/init.lua", "home/.config/init.lua"),
				candidate("dotfiles/yazi/config.yaml", "home/.config/config.yaml"),
				candidate("dotfiles/bestow/config.yaml", "home/.config/config.yaml"),
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
		pkg       string
		want      []operationCandidate
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "build operation candidates",
			setup: func() *Engine {
				mf := &MockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						files := make([]string, 0)
						for i := range 5 {
							fileName := fmt.Sprintf("file_%d", i)
							files = append(files, filepath.Join(parent, fileName))
						}
						return files, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			want: []operationCandidate{candidate("file_0", "file_0"), candidate("file_1", "file_1"), candidate("file_2", "file_2"), candidate("file_3", "file_3"), candidate("file_4", "file_4")},
		},
		{
			name: "build operation candidates - with ignore files",
			setup: func() *Engine {
				mf := &MockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						files := make([]string, 0)
						for i := range 5 {
							fileName := fmt.Sprintf("file_%d", i)
							files = append(files, filepath.Join(parent, fileName))
						}
						return files, nil
					},
				}
				ignoreList := newTestIgnoreList(mf, newTestLogger(), []string{"*0*"})
				return newTestEngine("", "", mf, ignoreList)
			},
			want: []operationCandidate{candidate("file_1", "file_1"), candidate("file_2", "file_2"), candidate("file_3", "file_3"), candidate("file_4", "file_4")},
		},
		{
			name: "build operation candidates - no files list",
			setup: func() *Engine {
				mf := &MockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						return nil, nil
					},
				}
				ignoreList := newTestIgnoreList(mf, newTestLogger(), []string{"*0*"})
				return newTestEngine("", "", mf, ignoreList)
			},
			want: make([]operationCandidate, 0),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			candidates, err := e.buildOperationCandidates(tc.pkg)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if !slices.Equal(candidates, tc.want) {
				t.Fatalf("got candidates %v, want %v", candidates, tc.want)
			}
		})
	}
}

func TestOperations_buildFileActions(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Engine
		candidates []operationCandidate
		strategy   ResolveStrategy
		cmdAction  CommandAction
		want       []fileAction
		wantErr    bool
		wantErrIs  error
		wantErrAs  func(*testing.T, error)
	}{
		{
			name: "build operations - stow all",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			candidates: []operationCandidate{candidate("file1", "file1"), candidate("file2", "file2"), candidate("file3", "file3")},
			strategy:   ResolveSkip,
			cmdAction:  CommandStow,
			want: []fileAction{
				newFileActionLink("file1", "file1", newTestLogger()),
				newFileActionLink("file2", "file2", newTestLogger()),
				newFileActionLink("file3", "file3", newTestLogger()),
			},
		},
		{
			name: "build operations - unstow all",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingManagedSymlink, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			candidates: []operationCandidate{candidate("file1", "file1"), candidate("file2", "file2"), candidate("file3", "file3")},
			strategy:   ResolveSkip,
			cmdAction:  CommandUnstow,
			want: []fileAction{
				newFileActionRemove("file1", "file1", newTestLogger()),
				newFileActionRemove("file2", "file2", newTestLogger()),
				newFileActionRemove("file3", "file3", newTestLogger()),
			},
		},
		{
			name: "build operations - collect errors",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, os.ErrPermission
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			candidates: []operationCandidate{candidate("file1", "file1"), candidate("file2", "file2"), candidate("file3", "file3")},
			strategy:   ResolveSkip,
			cmdAction:  CommandStow,
			wantErr:    true,
			wantErrAs: func(t *testing.T, err error) {
				var aggregatedErr *AggregatedError
				if !errors.As(err, &aggregatedErr) {
					t.Fatalf("got %v, want aggregatedErr", err)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			fileActions, err := e.buildFileActions(tc.candidates, tc.strategy, tc.cmdAction)
			if tc.wantErrAs != nil {
				tc.wantErrAs(t, err)
				return
			}
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if len(fileActions) != len(tc.want) {
				t.Fatalf("got file actions %d, want %d", len(fileActions), len(tc.want))
			}
			for i := range len(fileActions) {
				if !isSameAction(fileActions[i], tc.want[i]) {
					t.Fatalf("got %v, want %v", fileActions[i], tc.want[i])
				}
			}
		})
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
