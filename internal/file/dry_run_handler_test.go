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
			name:    "create file",
			setup:   func(t *testing.T, dir string) {},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "create file - unwritable dir",
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
			stat, err := os.Stat(path)
			if err != nil {
				if !errors.Is(err, tc.wantErrIs) {
					return
				}
				t.Fatalf("got error %v, want %v", err, os.ErrNotExist)
			}
			t.Fatalf("expected error, got %v", stat)
		})
	}
}

func TestDryRunHandler_CreateDir(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		dir       func(dir string) string
		handler   *DryRunHandler
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "create dir",
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
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "non writable dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, permNonWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			dir: func(dir string) string {
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
			if tc.dir != nil {
				dir = tc.dir(dir)
			}
			err := tc.handler.CreateDir(dir)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
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
			name:    "link",
			setup:   func(t *testing.T, src, dest string) {},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "link - non writable dir",
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
			name: "move",
			setup: func(t *testing.T, src, dest string) {
				destDir := filepath.Dir(dest)
				if err := os.Mkdir(destDir, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewDryRunHandler(newTestLogger()),
		},
		{
			name: "move - unwritable dest dir",
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
		})
	}
}

func TestDryRunHandler_Remove(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, path string)
		handler   *DryRunHandler
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "file remove",
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
			name: "file remove - unwritable dir",
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
			name: "file remove - non existing file",
			setup: func(t *testing.T, path string) {
				dir := filepath.Dir(path)
				if err := os.Mkdir(dir, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewDryRunHandler(newTestLogger()),
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
		})
	}
}
