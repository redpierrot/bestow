/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"os"
	"testing"

	"github.com/redpierrot/bestow/internal/file"
)

func TestStatus_Status(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Engine
		args        []string
		wantCountOf State
		want        int
		wantErr     bool
		wantErrIs   error
	}{
		{
			name: "no error",
			setup: func() *Engine {
				mf := &mockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						return []string{"file_1", "file_2", "file_3"}, nil
					},
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingManagedSymlink, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			args:        []string{"nvim"},
			wantCountOf: Stowed,
			want:        3,
		},
		{
			name: "find candidate error",
			setup: func() *Engine {
				mf := &mockFileSystem{
					isDirFn: func(path string) (bool, error) {
						return false, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			args:      []string{"nvim"},
			wantErr:   true,
			wantErrIs: errPkgIsNotDir,
		},
		{
			name: "get state error",
			setup: func() *Engine {
				mf := &mockFileSystem{
					listAllFilesFn: func(parent string) ([]string, error) {
						return []string{"file_1", "file_2", "file_3"}, nil
					},
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, os.ErrPermission
					},
				}
				return newTestEngine(mf, nil)
			},
			args:      []string{"nvim"},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			status, err := e.Status(tc.args)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if status.Count(tc.wantCountOf) != tc.want {
				t.Fatalf("got count %d want %d", status.Count(tc.wantCountOf), tc.want)
			}
		})
	}
}

func TestStatus_getState(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		candidate operationCandidate
		want      State
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "stowed",
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
			want: Stowed,
		},
		{
			name: "unstowed",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			want: Unstowed,
		},
		{
			name: "foreign symlink",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingForeignSymlink, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			want: Conflict,
		},
		{
			name: "regular file",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingRegularFile, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			want: Conflict,
		},
		{
			name: "directory",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingDir, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			want: Conflict,
		},
		{
			name: "unknown",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			want: Unknown,
		},
		{
			name: "existing error",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return false, os.ErrPermission
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, nil
					},
				}
				return newTestEngine(mf, nil)
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "existing file type error",
			setup: func() *Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
					existingFileTypeFn: func(src, dest string) (file.ExistingType, error) {
						return file.ExistingUnknown, os.ErrPermission
					},
				}
				return newTestEngine(mf, nil)
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup()
			state, err := e.getState(tc.candidate)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if state != tc.want {
				t.Fatalf("got %v, want %v", state, tc.want)
			}
		})
	}
}
