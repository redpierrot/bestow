/*
All Rights Reversed (ɔ)
*/

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestReadHandler_ListAllFiles(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		dirFn     func(t *testing.T, dir string) string
		handler   *Handler
		want      []string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "list files",
			setup: func(t *testing.T, dir string) {
				for i := range 5 {
					filePath := filepath.Join(dir, fmt.Sprintf("file_%d", i))
					if err := os.WriteFile(filePath, []byte("file content"), permFileWrite); err != nil {
						t.Fatal(err)
					}
				}
			},
			handler: NewHandler(newTestLogger()),
			want:    []string{"file_0", "file_1", "file_2", "file_3", "file_4"},
		},
		{
			name:    "list files - empty dir",
			setup:   func(t *testing.T, dir string) {},
			handler: NewHandler(newTestLogger()),
		},
		{
			name: "list files - with sub dirs",
			setup: func(t *testing.T, dir string) {
				paths := make([]string, 0, 3)
				for i := range 3 {
					path := filepath.Join(dir, fmt.Sprintf("subdir_%d", i))
					if err := os.Mkdir(path, permWritableDir); err != nil {
						t.Fatal(err)
					}
					paths = append(paths, path)
				}
				for _, path := range paths {
					if err := os.WriteFile(filepath.Join(path, "file"), []byte("file content"), permFileWrite); err != nil {
						t.Fatal(err)
					}
				}
			},
			handler: NewHandler(newTestLogger()),
			want:    []string{filepath.Join("subdir_0", "file"), filepath.Join("subdir_1", "file"), filepath.Join("subdir_2", "file")},
		},
		{
			name: "list files - with sub dirs as symlinks",
			setup: func(t *testing.T, dir string) {
				paths := make([]string, 0, 3)
				for i := range 3 {
					path := filepath.Join(dir, fmt.Sprintf("subdir_%d", i))
					if err := os.Mkdir(path, permWritableDir); err != nil {
						t.Fatal(err)
					}
					paths = append(paths, path)
				}
				for _, path := range paths {
					if err := os.WriteFile(filepath.Join(path, "file"), []byte("file content"), permFileWrite); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(filepath.Join(dir, "subdir_0"), filepath.Join(dir, "subdir_10")); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
			want:    []string{filepath.Join("subdir_0", "file"), filepath.Join("subdir_1", "file"), filepath.Join("subdir_2", "file")},
		},
		{
			name:  "list files - non existent path",
			setup: func(t *testing.T, dir string) {},
			dirFn: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "subdir")
			},
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: ErrNotDir,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			if tc.dirFn != nil {
				testRoot = tc.dirFn(t, testRoot)
			}
			tc.setup(t, testRoot)
			files, err := tc.handler.ListAllFiles(testRoot)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if len(files) != len(tc.want) {
				t.Fatalf("got fileCount %d, want %d", len(files), len(tc.want))
			}
			for _, wantFile := range tc.want {
				wantPath := filepath.Join(testRoot, wantFile)
				if !slices.Contains(files, wantPath) {
					t.Fatalf("missing %s", wantPath)
				}
			}
		})
	}
}

func TestReadHandler_ExistingFileType(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, src, dest string) (string, string)
		handler   *Handler
		want      ExistingType
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "existing file type dir",
			setup: func(t *testing.T, src, dest string) (string, string) {
				if err := os.Mkdir(dest, permNone); err != nil {
					t.Fatal(err)
				}
				return src, dest
			},
			handler: NewHandler(newTestLogger()),
			want:    ExistingDir,
		},
		{
			name: "existing file type regular file",
			setup: func(t *testing.T, src, dest string) (string, string) {
				if err := os.WriteFile(dest, []byte("file content"), permNone); err != nil {
					t.Fatal(err)
				}
				return src, dest
			},
			handler: NewHandler(newTestLogger()),
			want:    ExistingRegularFile,
		},
		{
			name: "existing file type non existing dest",
			setup: func(t *testing.T, src, dest string) (string, string) {
				return src, dest
			},
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrNotExist,
		},
		{
			name: "existing file type non existing src",
			setup: func(t *testing.T, src, dest string) (string, string) {
				if err := os.Mkdir(src, permWritableDir); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(dest, permWritableDir); err != nil {
					t.Fatal(err)
				}
				srcPath := filepath.Join(src, "source_file")
				tempPath := filepath.Join(src, "temp_file")
				if err := os.WriteFile(tempPath, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
				destPath := filepath.Join(dest, "dest_file")
				if err := os.Symlink(tempPath, destPath); err != nil {
					t.Fatal(err)
				}
				return srcPath, destPath
			},
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: os.ErrNotExist,
		},
		{
			name: "existing file type dest link missing",
			setup: func(t *testing.T, src, dest string) (string, string) {
				if err := os.Mkdir(src, permWritableDir); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(dest, permWritableDir); err != nil {
					t.Fatal(err)
				}
				srcPath := filepath.Join(src, "source_file")
				if err := os.WriteFile(srcPath, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
				tempPath := filepath.Join(src, "temp_file")
				if err := os.WriteFile(tempPath, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
				destPath := filepath.Join(dest, "dest_file")
				if err := os.Symlink(tempPath, destPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(tempPath); err != nil {
					t.Fatal(err)
				}
				return srcPath, destPath
			},
			handler: NewHandler(newTestLogger()),
			want:    ExistingForeignSymlink,
		},
		{
			name: "existing file type managed symlink",
			setup: func(t *testing.T, src, dest string) (string, string) {
				if err := os.Mkdir(src, permWritableDir); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(dest, permWritableDir); err != nil {
					t.Fatal(err)
				}
				srcPath := filepath.Join(src, "source_file")
				if err := os.WriteFile(srcPath, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
				destPath := filepath.Join(dest, "dest_file")
				if err := os.Symlink(srcPath, destPath); err != nil {
					t.Fatal(err)
				}
				return srcPath, destPath
			},
			handler: NewHandler(newTestLogger()),
			want:    ExistingManagedSymlink,
		},
		{
			name: "existing file type foreign symlink - different target",
			setup: func(t *testing.T, src, dest string) (string, string) {
				if err := os.Mkdir(src, permWritableDir); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(dest, permWritableDir); err != nil {
					t.Fatal(err)
				}
				srcPath := filepath.Join(src, "source_file")
				if err := os.WriteFile(srcPath, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
				tempPath := filepath.Join(src, "temp_file")
				if err := os.WriteFile(tempPath, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
				destPath := filepath.Join(dest, "dest_file")
				if err := os.Symlink(tempPath, destPath); err != nil {
					t.Fatal(err)
				}
				return srcPath, destPath
			},
			handler: NewHandler(newTestLogger()),
			want:    ExistingForeignSymlink,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			src := filepath.Join(testRoot, "src")
			dest := filepath.Join(testRoot, "dest")
			srcFile, destFile := tc.setup(t, src, dest)
			existingType, err := tc.handler.ExistingFileType(srcFile, destFile)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if existingType != tc.want {
				t.Fatalf("got existingType %v, want %v", existingType, tc.want)
			}
		})
	}
}

func TestReadHandler_ListDirs(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, parent string)
		handler   *Handler
		want      []string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "list dirs",
			setup: func(t *testing.T, parent string) {
				if err := os.Mkdir(parent, permWritableDir); err != nil {
					t.Fatal(err)
				}
				for i := range 5 {
					subDir := filepath.Join(parent, fmt.Sprintf("subdir_%d", i))
					if err := os.Mkdir(subDir, permWritableDir); err != nil {
						t.Fatal(err)
					}
					t.Logf("created: %s", subDir)
				}
			},
			handler: NewHandler(newTestLogger()),
			want:    []string{"subdir_0", "subdir_1", "subdir_2", "subdir_3", "subdir_4"},
		},
		{
			name: "list dirs - empty dir",
			setup: func(t *testing.T, parent string) {
				if err := os.Mkdir(parent, permWritableDir); err != nil {
					t.Fatal(err)
				}
			},
			handler: NewHandler(newTestLogger()),
			want:    make([]string, 0),
		},
		{
			name: "list dirs - non dir path",
			setup: func(t *testing.T, parent string) {
				if err := os.WriteFile(parent, []byte("sample file content"), permFileWrite); err != nil {
					t.Fatal(err)
				}
			},
			handler:   NewHandler(newTestLogger()),
			wantErr:   true,
			wantErrIs: ErrNotDir,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			parent := filepath.Join(testRoot, "dest")
			tc.setup(t, parent)
			got, err := tc.handler.ListDirs(parent)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if len(tc.want) != len(got) {
				t.Fatalf("got len(dirs) %d, want %d", len(got), len(tc.want))
			}
			for _, wantDir := range tc.want {
				wantPath := filepath.Join(parent, wantDir)
				if !slices.Contains(got, wantPath) {
					t.Fatalf("missing %s", wantPath)
				}
			}
		})
	}
}

func TestReadHandler_ReadLines(t *testing.T) {
	tests := []struct {
		name      string
		handler   *Handler
		setup     func(t *testing.T, path, content string)
		want      []string
		content   string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "read lines",
			handler: NewHandler(newTestLogger()),
			setup: func(t *testing.T, path, content string) {
				if err := os.WriteFile(path, []byte(content), permFileWrite); err != nil {
					t.Fatal(err)
				}
			},
			content: "sample file content\nwith lines",
			want:    []string{"sample file content", "with lines"},
		},
		{
			name:      "read lines - non existing file",
			handler:   NewHandler(newTestLogger()),
			setup:     func(t *testing.T, path, content string) {},
			content:   "",
			wantErr:   true,
			wantErrIs: os.ErrNotExist,
		},
		{
			name:    "read lines - no perm",
			handler: NewHandler(newTestLogger()),
			setup: func(t *testing.T, path, content string) {
				if err := os.WriteFile(path, []byte(content), permNone); err != nil {
					t.Fatal(err)
				}
			},
			content:   "",
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRoot := t.TempDir()
			path := filepath.Join(testRoot, "file")
			tc.setup(t, path, tc.content)
			content, err := tc.handler.ReadLines(path)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if !slices.Equal(content, tc.want) {
				t.Fatalf("got %v, want %v", content, tc.want)
			}
		})
	}
}
