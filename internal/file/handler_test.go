/*
All Rights Reversed (ɔ)
*/

package file

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

const (
	permNone           = 0o000
	permNonWritableDir = 0o500
	permWritableDir    = 0o744
)

func TestHandler_CreateFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		setup     func(t *testing.T, parent string)
		wantErr   bool
		wantErrIs error
		handler   *Handler
	}{
		{
			name:    "no error",
			setup:   func(t *testing.T, parent string) {},
			content: "this is sample file content",
			handler: NewHandler(newTestLogger()),
		},
		{
			name: "no perm parent",
			setup: func(t *testing.T, parent string) {
				if os.Getuid() == 0 {
					t.Skip("root bypasses the permission checks")
				}
				if err := os.Chmod(parent, permNone); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(parent, permWritableDir) })
			},
			content:   "this is sample file content",
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "existing path",
			setup: func(t *testing.T, parent string) {
				if err := os.Chmod(parent, permWritableDir); err != nil {
					t.Fatal(err)
				}
				_, err := os.Create(filepath.Join(parent, "file_path"))
				if err != nil {
					t.Fatal(err)
				}
			},
			content: "this is sample file content",
			handler: NewHandler(newTestLogger()),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			path := filepath.Join(testRoot, "file_path")
			tc.setup(t, testRoot)
			err := tc.handler.CreateFile(path, tc.content)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(content) != tc.content {
				t.Fatalf("got %v, want %v", string(content), tc.content)
			}
		})
	}
}

func TestHandler_CreateDir(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, path string)
		pathsFn   func(parent string) []string
		handler   *Handler
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "non existing path",
			setup:   func(t *testing.T, path string) {},
			handler: NewHandler(newTestLogger()),
		},
		{
			name: "existing dir",
			setup: func(t *testing.T, path string) {
				if err := os.Mkdir(path, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
		},
		{
			name:    "create sub dirs on same path",
			setup:   func(t *testing.T, path string) {},
			handler: NewHandler(newTestLogger()),
			pathsFn: func(parent string) []string {
				return []string{filepath.Join(parent, "l1", "d1"), filepath.Join(parent, "l1", "d2"), filepath.Join(parent, "l1", "d1")}
			},
		},
		{
			name: "no perm",
			setup: func(t *testing.T, path string) {
				if os.Getuid() == 0 {
					t.Skip("root bypasses permission checks")
				}
				if err := os.Chmod(filepath.Dir(path), permNone); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(filepath.Dir(path), permWritableDir) })
			},
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			var paths []string
			if tc.pathsFn != nil {
				paths = tc.pathsFn(testRoot)
			} else {
				paths = []string{filepath.Join(testRoot, "dest")}
			}
			for _, path := range paths {
				tc.setup(t, path)
				t.Logf("creating dir: %s", path)
				err := tc.handler.CreateDir(path)
				if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
					return
				}
			}
		})
	}
}

func TestHandler_Link(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir, src, dest string)
		destFn    func(dir string) string
		handler   *Handler
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "existing parent",
			setup:   func(t *testing.T, dir, src, dest string) {},
			handler: NewHandler(newTestLogger()),
		},
		{
			name:    "non-existing parent",
			setup:   func(t *testing.T, dir, src, dest string) {},
			destFn:  func(dir string) string { return filepath.Join(dir, "sub_directory", "dest_file") },
			handler: NewHandler(newTestLogger()),
		},
		{
			name: "existing dest file",
			setup: func(t *testing.T, dir, src, dest string) {
				if err := os.WriteFile(dest, []byte("destination file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
			},
			destFn:    func(dir string) string { return filepath.Join(dir, "dest_file") },
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrExist,
		},
		{
			name: "parent not writable",
			setup: func(t *testing.T, dir, src, dest string) {
				if os.Getuid() == 0 {
					t.Skip("root bypasses permission checks")
				}
				if err := os.Chmod(filepath.Dir(dest), permNone); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(filepath.Dir(dest), permWritableDir) })
			},
			destFn:    func(dir string) string { return filepath.Join(dir, "dest_file") },
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			src := filepath.Join(testRoot, "src_file")
			if err := os.WriteFile(src, []byte("Sample Config"), permFileWrite); err != nil {
				t.Fatal(err)
			}

			dest := filepath.Join(testRoot, "dest_file")
			if tc.destFn != nil {
				dest = tc.destFn(testRoot)
			}
			tc.setup(t, testRoot, src, dest)
			err := tc.handler.Link(src, dest)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			got, readErr := os.Readlink(dest)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if got != src {
				t.Fatalf("got symlink target %q, want %q", got, src)
			}
		})
	}
}

func TestHandler_IsEmptyDir(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		handler   *Handler
		want      bool
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "empty dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
			want:    true,
		},
		{
			name: "non-empty dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, permWritableDir); err != nil {
					t.Fatal(err)
				}
				tmpFile := filepath.Join(dir, "source_file")
				if err := os.WriteFile(tmpFile, []byte("Sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
			want:    false,
		},
		{
			name:      "non-existent dir",
			setup:     func(t *testing.T, dir string) {},
			handler:   NewHandler(newTestLogger()),
			want:      false,
			wantErr:   true,
			wantErrIs: ErrNotDir,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			src := filepath.Join(testRoot, "source")
			tc.setup(t, src)
			isEmpty, err := tc.handler.IsEmptyDir(src)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if isEmpty != tc.want {
				t.Fatalf("got isEmpty %v, want %v", isEmpty, tc.want)
			}
		})
	}
}

func TestHandler_Remove(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		setup     func(t *testing.T, dir string)
		wantErr   bool
		wantErrIs error
		handler   *Handler
	}{
		{
			name: "file",
			path: "src_file",
			setup: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("test file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
		},
		{
			name: "dir",
			path: "src_dir",
			setup: func(t *testing.T, path string) {
				if err := os.Mkdir(path, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
		},
		{
			name:    "non-existing",
			path:    "src_dir",
			setup:   func(t *testing.T, path string) {},
			handler: NewHandler(newTestLogger()),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			path := filepath.Join(testRoot, tc.path)
			tc.setup(t, path)
			err := tc.handler.Remove(path)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			_, err = os.Stat(path)
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("got err %v, want %v", err, os.ErrNotExist)
			}
		})
	}
}

func TestHandler_Move(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, src, dest string)
		wantErr   bool
		wantErrIs error
		handler   *Handler
	}{
		{
			name: "existing file",
			setup: func(t *testing.T, src, _ string) {
				if err := os.WriteFile(src, []byte("test file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
		},
		{
			name: "no perm",
			setup: func(t *testing.T, src, dest string) {
				if err := os.WriteFile(src, []byte("test file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(filepath.Dir(dest), permNone); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(filepath.Dir(dest), permWritableDir) })
			},
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			src := filepath.Join(testRoot, "src")
			dest := filepath.Join(testRoot, "dest")
			tc.setup(t, src, dest)
			err := tc.handler.Move(src, dest)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			_, err = os.Stat(dest)
			if err != nil {
				t.Fatal(err)
			}
			_, err = os.Stat(src)
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("got error %v, want %v", err, os.ErrNotExist)
			}
		})
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func validateErrScenario(t *testing.T, wantErr bool, err, wantErrIs error) bool {
	t.Helper()
	if (err != nil) != wantErr {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if wantErr && wantErrIs != nil && !errors.Is(err, wantErrIs) {
		t.Fatalf("error got %v, want %v", err, wantErrIs)
	}
	return wantErr
}
