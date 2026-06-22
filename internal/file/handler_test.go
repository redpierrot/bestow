package file

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const noWritePerm = 0o555
const writePerm = 0o755

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
				if err := os.WriteFile(dest, []byte("destination file content"), filePermissions); err != nil {
					t.Fatalf("dest creation: %v", err)
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
				if err := os.Chmod(filepath.Dir(dest), noWritePerm); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(filepath.Dir(dest), writePerm) })
			},
			destFn:    func(dir string) string { return filepath.Join(dir, "dest_file") },
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "src_file")
			if err := os.WriteFile(src, []byte("Sample Config"), 0o644); err != nil {
				t.Fatalf("source creation failed: %v", err)
			}

			dest := filepath.Join(dir, "dest_file")
			if tc.destFn != nil {
				dest = tc.destFn(dir)
			}
			tc.setup(t, dir, src, dest)
			err := tc.handler.Link(src, dest)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("errors.Is(%v), want %v", err, tc.wantErrIs)
				}
				return
			}
			got, readErr := os.Readlink(dest)
			if readErr != nil {
				t.Fatalf("symlink missing: %v", readErr)
			}
			if got != src {
				t.Fatalf("symlink target %q, want %q", got, src)
			}
		})
	}
}

func TestHandler_CreateFile(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		setup     func(t *testing.T, parent string)
		wantErr   bool
		wantErrIs error
		handler   *Handler
	}{
		{
			name:    "create file",
			setup:   func(t *testing.T, parent string) {},
			lines:   []string{"this is sample file content"},
			handler: NewHandler(newTestLogger()),
		},
		{
			name: "no perm parent",
			setup: func(t *testing.T, parent string) {
				if os.Getuid() == 0 {
					t.Skip("root bypasses the permission checks")
				}
				if err := os.Chmod(parent, noWritePerm); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(parent, writePerm) })
			},
			lines:     []string{"this is sample file content"},
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "file_path")
			tc.setup(t, dir)
			err := tc.handler.CreateFile(path, strings.Join(tc.lines, "\n"))
			if tc.wantErr {
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("errors.Is(%v), want %v", err, tc.wantErrIs)
				}
				return
			}
			result, err := tc.handler.ReadLines(path)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(tc.lines, result) {
				t.Fatalf("content mismatch read: %v, want: %v", result, tc.lines)
			}

		})
	}
}

func TestHandler_IsEmptyDir(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		handler   Handler
		want      bool
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "check empty dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, writePerm); err != nil {
					t.Fatal(err)
				}
			},
			handler: *NewHandler(newTestLogger()),
			want:    true,
		},
		{
			name: "check non-empty dir",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(dir, writePerm); err != nil {
					t.Fatal(err)
				}
				tmpFile := filepath.Join(dir, "source_file")
				if err := os.WriteFile(tmpFile, []byte("Sample file content"), writePerm); err != nil {
					t.Fatal(err)
				}
			},
			handler: *NewHandler(newTestLogger()),
			want:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "source")
			tc.setup(t, src)
			isEmpty, err := tc.handler.IsEmptyDir(src)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, want = %v", err, tc.wantErr)
			}
			if err != nil {
				if errors.Is(err, tc.wantErrIs) {
					return
				}
				t.Fatalf("errors.Is(%v), want %v", err, tc.wantErrIs)
			}
			if isEmpty != tc.want {
				t.Fatalf("isEmpty: %v, want: %v", isEmpty, tc.want)
			}
		})
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
