/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/redpierrot/bestow/internal/file"
)

func TestOperations_buildOperations(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		cfg       *CommandConfig
		want      []fileAction
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "stow link",
			setup: func() *Engine {
				mf := &MockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						return []string{"src_file_1", "src_file_2"}, nil
					},
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			cfg: &CommandConfig{
				Action: CommandStow,
				Args:   []string{"bestow"},
			},
			want: []fileAction{
				newFileActionLink("src_file_1", "src_file_1", newTestLogger()),
				newFileActionLink("src_file_2", "src_file_2", newTestLogger()),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			actions, err := e.buildOperations(tc.cfg)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if len(actions) != len(tc.want) {
				t.Fatalf("got actions %d, want %d", len(actions), len(tc.want))
			}
			for i := range len(actions) {
				if actions[i].kind() != tc.want[i].kind() {
					t.Fatalf("got action %v, want %v", actions[i], tc.want[i])
				}
			}
		})
	}
}

func TestOperations_validateDestinations(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Engine
		candidates []operationCandidate
		wantErr    bool
		wantErrAs  func(*testing.T, error)
	}{
		{
			name: "no conflict",
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
			name: "conflict",
			setup: func() *Engine {
				mf := &MockFileSystem{}
				return newTestEngine("", "", mf, nil)
			},
			candidates: []operationCandidate{
				candidate("dotfiles/nvim/init.lua", "home/.config/init.lua"),
				candidate("dotfiles/yazi/config.yaml", "home/.config/config.yaml"),
				candidate("dotfiles/bestow/config.yaml", "home/.config/config.yaml"),
			},
			wantErr: true,
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
			if validateErrScenario(t, tc.wantErr, err, nil) {
				if tc.wantErrAs != nil {
					tc.wantErrAs(t, err)
				}
				return
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
			name: "no ignore list",
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
			want: []operationCandidate{
				candidate("file_0", "file_0"),
				candidate("file_1", "file_1"),
				candidate("file_2", "file_2"),
				candidate("file_3", "file_3"),
				candidate("file_4", "file_4"),
			},
		},
		{
			name: "with ignore list",
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
			name: "empty files list",
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
			name: "stow all",
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
			name: "unstow all",
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
			name: "collect errors",
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
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				if tc.wantErrAs != nil {
					tc.wantErrAs(t, err)
				}
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
		strategy  ResolveStrategy
		want      ActionKind
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "non existing destination",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveSkip,
			want:     ActionLink,
		},
		{
			name: "existing dir: skip",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy:  ResolveSkip,
			wantErr:   true,
			wantErrIs: errDestIsDir,
		},
		{
			name: "existing dir: force",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy:  ResolveForce,
			wantErr:   true,
			wantErrIs: errDestIsDir,
		},
		{
			name: "existing dir: adopt",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy:  ResolveAdopt,
			wantErr:   true,
			wantErrIs: errDestIsDir,
		},
		{
			name: "existing dir: backup",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy:  ResolveBackup,
			wantErr:   true,
			wantErrIs: errDestIsDir,
		},
		{
			name: "existing managed symlink: skip",
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
			strategy: ResolveSkip,
			want:     ActionUpToDate,
		},
		{
			name: "existing managed symlink: force",
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
			strategy: ResolveForce,
			want:     ActionUpToDate,
		},
		{
			name: "existing managed symlink: adopt",
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
			strategy: ResolveAdopt,
			want:     ActionUpToDate,
		},
		{
			name: "existing managed symlink: backup",
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
			strategy: ResolveBackup,
			want:     ActionUpToDate,
		},
		{
			name: "existing foreign symlink: skip",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingForeignSymlink, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveSkip,
			want:     ActionSkip,
		},
		{
			name: "existing foreign symlink: force",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingForeignSymlink, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveForce,
			want:     ActionReplace,
		},
		{
			name: "existing foreign symlink: adopt",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingForeignSymlink, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveAdopt,
			want:     ActionSkip,
		},
		{
			name: "existing foreign symlink: backup",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						if strings.Contains(path, "backup") {
							return false, nil
						}
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingForeignSymlink, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveBackup,
			want:     ActionBackup,
		},
		{
			name: "existing regular file: skip",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingRegularFile, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveSkip,
			want:     ActionSkip,
		},
		{
			name: "existing regular file: force",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingRegularFile, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveForce,
			want:     ActionReplace,
		},
		{
			name: "existing regular file: adopt",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingRegularFile, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveAdopt,
			want:     ActionAdopt,
		},
		{
			name: "existing regular file: backup",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						if strings.Contains(path, "backup") {
							return false, nil
						}
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingRegularFile, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			strategy: ResolveBackup,
			want:     ActionBackup,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			cand := candidate("", "") // Dummy candidate since we don't care about paths here.
			fa, err := e.stowFileAction(cand, tc.strategy)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if fa.kind() != tc.want {
				t.Fatalf("got %v, want %v", fa.kind(), tc.want)
			}
		})
	}
}

func TestOperations_unstowFileAction(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		candidate operationCandidate
		want      ActionKind
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "existing managed symlink",
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
			candidate: candidate("src_file", "dest_file"),
			want:      ActionRemove,
		},
		{
			name: "existing foreign symlink",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingForeignSymlink, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionSkip,
		},
		{
			name: "existing regular file",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingRegularFile, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionSkip,
		},
		{
			name: "existing dir",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			candidate: candidate("src_file", "dest_file"),
			wantErr:   true,
			wantErrIs: errDestIsDir,
		},
		{
			name: "dest not exist",
			setup: func() *Engine {
				mf := &MockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
				return newTestEngine("", "", mf, nil)
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionUpToDate,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			fa, err := e.unstowFileAction(tc.candidate)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if fa.kind() != tc.want {
				t.Fatalf("got %v, want %v", fa.kind(), tc.want)
			}
		})
	}
}

func TestOperations_calculateBackupPath(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(existingFiles []string) *Engine
		existingFiles []string
		want          string
		wantErr       bool
		wantErrIs     error
		wantErrAs     func(*testing.T, error)
	}{
		{
			name: "no existing backup",
			setup: func(_ []string) *Engine {
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
			name: "existing backed up files",
			setup: func(existingFiles []string) *Engine {
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
			name: "existing return err",
			setup: func(existingFiles []string) *Engine {
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
			name: "exceeds maximum backups",
			setup: func(existingFiles []string) *Engine {
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
			wantErr: true,
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
			eng := tc.setup(tc.existingFiles)
			backupPath, err := eng.calculateBackupPath(path)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				if tc.wantErrAs != nil {
					tc.wantErrAs(t, err)
				}
				return
			}
			if backupPath != tc.want {
				t.Fatalf("got %v, want %v", backupPath, tc.want)
			}
		})
	}
}
