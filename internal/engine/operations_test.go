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
		args      []string
		want      []operationCandidate
		wantErr   bool
		wantErrIs error
		wantErrAs func(t *testing.T, err error)
	}{
		{
			name: "stow link",
			setup: func() *Engine {
				mf := &mockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						return []string{"/Users/ru/dotfiles/bestow/src_file_1", "/Users/ru/dotfiles/bestow/src_file_2"}, nil
					},
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
				}
				e := newTestEngine(mf, nil)
				e.source = "/Users/ru/dotfiles"
				e.destination = "/Users/ru/"
				return e
			},
			args: []string{"bestow"},
			want: []operationCandidate{
				{source: "/Users/ru/dotfiles/bestow/src_file_1", destination: "/Users/ru/src_file_1"},
				{source: "/Users/ru/dotfiles/bestow/src_file_2", destination: "/Users/ru/src_file_2"},
			},
		},
		{
			name: "multiple errors",
			setup: func() *Engine {
				mf := &mockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						return nil, os.ErrPermission
					},
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			args:    []string{"bestow", "nvim", "stow"},
			wantErr: true,
			wantErrAs: func(t *testing.T, err error) {
				var expected *AggregatedError
				if !errors.As(err, &expected) {
					t.Fatalf("got %v, want AggregatedError", err)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			actions, err := e.findCandidates(tc.args)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				if tc.wantErrAs != nil {
					tc.wantErrAs(t, err)
				}
				return
			}
			if len(actions) != len(tc.want) {
				t.Fatalf("got candidates %d, want %d", len(actions), len(tc.want))
			}
			for i := range len(actions) {
				if actions[i] != tc.want[i] {
					t.Fatalf("got candidate %v, want %v", actions[i], tc.want[i])
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
				mf := &mockFileSystem{}
				return newTestEngine(mf, nil)
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
				mf := &mockFileSystem{}
				return newTestEngine(mf, nil)
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
				mf := &mockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						files := make([]string, 0)
						for i := range 5 {
							fileName := fmt.Sprintf("file_%d", i)
							files = append(files, filepath.Join(parent, fileName))
						}
						return files, nil
					},
				}
				return newTestEngine(mf, nil)
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
				mf := &mockFileSystem{
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
				return newTestEngine(mf, ignoreList)
			},
			want: []operationCandidate{candidate("file_1", "file_1"), candidate("file_2", "file_2"), candidate("file_3", "file_3"), candidate("file_4", "file_4")},
		},
		{
			name: "empty files list",
			setup: func() *Engine {
				mf := &mockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						return nil, nil
					},
				}
				ignoreList := newTestIgnoreList(mf, newTestLogger(), []string{"*0*"})
				return newTestEngine(mf, ignoreList)
			},
			want: make([]operationCandidate, 0),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			candidates, err := e.findPackageCandidates(tc.pkg)
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
		cmdAction  CommandKind
		want       []fileAction
		wantErr    bool
		wantErrIs  error
		wantErrAs  func(*testing.T, error)
	}{
		{
			name: "stow all",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
				}
				return newTestEngine(mf, nil)
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
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingManagedSymlink, nil
					},
				}
				return newTestEngine(mf, nil)
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
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, os.ErrPermission
					},
				}
				return newTestEngine(mf, nil)
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

func TestOperations_stowOperation(t *testing.T) {
	type strategyCase struct {
		name      string
		strategy  ResolveStrategy
		want      ActionKind
		wantErr   bool
		wantErrIs error
	}
	tests := []struct {
		name  string
		fs    func() *mockFileSystem
		cases []strategyCase
	}{
		{
			name: "dest not exist",
			fs: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
				}
			},
			cases: []strategyCase{
				{name: "skip", strategy: ResolveSkip, want: ActionLink},
			},
		},
		{
			name: "exiting file type function error",
			fs: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, os.ErrPermission
					},
				}
			},
			cases: []strategyCase{
				{name: "skip", strategy: ResolveSkip, wantErr: true, wantErrIs: os.ErrPermission},
			},
		},
		{
			name: "dest is dir",
			fs: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
			},
			cases: []strategyCase{
				{name: "skip", strategy: ResolveSkip, wantErr: true, wantErrIs: errDestIsDir},
				{name: "force", strategy: ResolveForce, wantErr: true, wantErrIs: errDestIsDir},
				{name: "adopt", strategy: ResolveAdopt, wantErr: true, wantErrIs: errDestIsDir},
				{name: "backup", strategy: ResolveBackup, wantErr: true, wantErrIs: errDestIsDir},
			},
		},
		{
			name: "existing managed symlink",
			fs: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingManagedSymlink, nil
					},
				}
			},
			cases: []strategyCase{
				{name: "skip", strategy: ResolveSkip, want: ActionUpToDate},
				{name: "force", strategy: ResolveForce, want: ActionUpToDate},
				{name: "adopt", strategy: ResolveAdopt, want: ActionUpToDate},
				{name: "backup", strategy: ResolveBackup, want: ActionUpToDate},
			},
		},
		{
			name: "existing foreign symlink",
			fs: func() *mockFileSystem {
				return &mockFileSystem{
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
			},
			cases: []strategyCase{
				{name: "skip", strategy: ResolveSkip, want: ActionSkip},
				{name: "force", strategy: ResolveForce, want: ActionReplace},
				{name: "adopt", strategy: ResolveAdopt, want: ActionSkip},
				{name: "backup", strategy: ResolveBackup, want: ActionBackup},
			},
		},
		{
			name: "existing regular file",
			fs: func() *mockFileSystem {
				return &mockFileSystem{
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
			},
			cases: []strategyCase{
				{name: "skip", strategy: ResolveSkip, want: ActionSkip},
				{name: "force", strategy: ResolveForce, want: ActionReplace},
				{name: "adopt", strategy: ResolveAdopt, want: ActionAdopt},
				{name: "backup", strategy: ResolveBackup, want: ActionBackup},
			},
		},
		{
			name: "unknown file type",
			fs: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src string, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, nil
					},
				}
			},
			cases: []strategyCase{
				{name: "skip", strategy: 100, wantErr: true, wantErrIs: errUnsupportedAction},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, st := range tc.cases {
				t.Run(st.name, func(t *testing.T) {
					mf := tc.fs()
					cand := candidate("", "") // Dummy candidate since we don't care about paths here.
					operation := newStowOperation(mf, newTestLogger(), st.strategy)
					fa, err := operation.FileAction(cand)
					if validateErrScenario(t, st.wantErr, err, st.wantErrIs) {
						return
					}
					if fa.kind() != st.want {
						t.Fatalf("got %v, want %v", fa.kind(), st.want)
					}
				})
			}
		})
	}
}

func TestOperations_unstowOperation(t *testing.T) {
	tests := []struct {
		name      string
		mf        func() *mockFileSystem
		candidate operationCandidate
		want      ActionKind
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "existing managed symlink",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingManagedSymlink, nil
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionRemove,
		},
		{
			name: "existing foreign symlink",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingForeignSymlink, nil
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionSkip,
		},
		{
			name: "existing regular file",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingRegularFile, nil
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionSkip,
		},
		{
			name: "existing dir",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			wantErr:   true,
			wantErrIs: errDestIsDir,
		},
		{
			name: "dest not exist",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionUpToDate,
		},
		{
			name: "unknown destination type",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, nil
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionSkip,
		},
		{
			name: "non existing destination",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			want:      ActionUpToDate,
		},
		{
			name: "existing function error",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, os.ErrPermission
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "exiting type function error",
			mf: func() *mockFileSystem {
				return &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, os.ErrPermission
					},
				}
			},
			candidate: candidate("src_file", "dest_file"),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf := tc.mf()
			operation := newUnstowOperation(mf, newTestLogger())
			fa, err := operation.FileAction(tc.candidate)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if fa.kind() != tc.want {
				t.Fatalf("got %v, want %v", fa.kind(), tc.want)
			}
		})
	}
}

func TestOperations_getOperation(t *testing.T) {
	tests := []struct {
		name      string
		kind      CommandKind
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "stow",
			kind: CommandStow,
		},
		{
			name: "unstow",
			kind: CommandUnstow,
		},
		{
			name:      "undefined",
			kind:      100,
			wantErr:   true,
			wantErrIs: errUnsupportedAction,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			operation, err := getOperation(tc.kind, &mockFileSystem{}, newTestLogger(), ResolveSkip)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if operation == nil {
				t.Fatalf("got nil, want Operation")
			}
		})
	}
}
