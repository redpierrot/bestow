/*
All Rights Reversed (ɔ)
*/

package file

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDryRunHandler_CreateFile(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		handler   *DryRunHandler
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "non existing",
			setup:   func(t *testing.T, dir string) {},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "unwritable dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, permNonWritableDir); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, permWritableDir) })
			},
			handler:   NewDryRunHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			dir := filepath.Join(testRoot, "dir")
			tc.setup(t, dir)
			path := filepath.Join(dir, "file")
			err := tc.handler.CreateFile(path, "sample file content")
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			_, err = os.Stat(path)
			if err == nil {
				t.Fatalf("file actually created %s", path)
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("got error %v, want %v", err, os.ErrNotExist)
			}
		})
	}
}

func TestDryRunHandler_CreateDir(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(t *testing.T, dir string)
		dirFn              func(dir string) string
		skipPostValidation bool
		handler            *DryRunHandler
		wantErr            bool
		wantErrIs          error
	}{
		{
			name:    "non existing",
			setup:   func(t *testing.T, dir string) {},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "existing dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			skipPostValidation: true,
			handler:            NewDryRunHandler(newTestLogger()),
		},
		{
			name: "non writable dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, permNonWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			dirFn: func(dir string) string {
				return filepath.Join(dir, "subdir")
			},
			handler:   NewDryRunHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			dir := filepath.Join(testRoot, "dir")
			tc.setup(t, dir)
			if tc.dirFn != nil {
				dir = tc.dirFn(dir)
			}
			err := tc.handler.CreateDir(dir)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if tc.skipPostValidation {
				return
			}
			_, err = os.Stat(dir)
			if err == nil {
				t.Fatalf("directory created %s", dir)
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("got %v, want %v", err, os.ErrNotExist)
			}
		})
	}
}

func TestDryRunHandler_Link(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, src, dest string)
		handler   *DryRunHandler
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "non existing dest",
			setup:   func(t *testing.T, src, dest string) {},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "non writable dir",
			setup: func(t *testing.T, src, dest string) {
				destDir := filepath.Dir(dest)
				if err := os.Mkdir(destDir, permNonWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler:   NewDryRunHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			src := filepath.Join(testRoot, "src", "src_file")
			dest := filepath.Join(testRoot, "dest", "dest_file")
			tc.setup(t, src, dest)
			err := tc.handler.Link(src, dest)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			_, err = os.Stat(dest)
			if err == nil {
				t.Fatalf("link exists %s", dest)
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("got %v, want %v", err, os.ErrNotExist)
			}
		})
	}
}

func TestDryRunHandler_Move(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, src, dest string)
		handler   *DryRunHandler
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "existing",
			setup: func(t *testing.T, src, dest string) {
				destDir := filepath.Dir(dest)
				if err := os.Mkdir(destDir, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "unwritable dest dir",
			setup: func(t *testing.T, src, dest string) {
				destDir := filepath.Dir(dest)
				if err := os.Mkdir(destDir, permNonWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler:   NewDryRunHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			src := filepath.Join(testRoot, "src", "src_file")
			dest := filepath.Join(testRoot, "dest", "dest_file")
			tc.setup(t, src, dest)
			err := tc.handler.Move(src, dest)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			_, err = os.Stat(dest)
			if err == nil {
				t.Fatalf("got nil, want %v", os.ErrNotExist)
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("got %v, want %v", err, os.ErrNotExist)
			}
		})
	}
}

func TestDryRunHandler_Remove(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(t *testing.T, path string)
		handler            *DryRunHandler
		skipPostValidation bool
		wantErr            bool
		wantErrIs          error
	}{
		{
			name: "existing file",
			setup: func(t *testing.T, path string) {
				dir := filepath.Dir(path)
				if err := os.Mkdir(dir, permWritableDir); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "no perm file",
			setup: func(t *testing.T, path string) {
				dir := filepath.Dir(path)
				if err := os.Mkdir(dir, permWritableDir); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("sample file content"), permNone); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(path, permFileWrite) })
			},
			handler:   NewDryRunHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "non existing file",
			setup: func(t *testing.T, path string) {
				dir := filepath.Dir(path)
				if err := os.Mkdir(dir, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler:            NewDryRunHandler(newTestLogger()),
			skipPostValidation: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			path := filepath.Join(testRoot, "dir", "file")
			tc.setup(t, path)
			err := tc.handler.Remove(path)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if tc.skipPostValidation {
				return
			}
			if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
				t.Fatalf("file %s is removed by dry run handler", path)
			}
		})
	}
}
