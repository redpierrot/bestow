/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestPackageOperations_buildPackageList(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		args      []string
		want      []string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "build package list",
			setup: func() *Engine {
				src := filepath.Join("home", "user", "dotfiles")
				dest := filepath.Join("home", "user")
				packageList := getSamplePackageList(src)
				fs := &MockFileSystem{
					listDirFn: func(parent string) ([]string, error) {
						return packageList, nil
					},
				}
				return newTestEngine(src, dest, fs, nil)
			},
			want: []string{"pkg1", "pkg2", "pkg3"},
		},
		{
			name: "build package list - with args",
			setup: func() *Engine {
				src := filepath.Join("home", "user", "dotfiles")
				dest := filepath.Join("home", "user")
				packageList := getSamplePackageList(src)
				fs := &MockFileSystem{
					listDirFn: func(parent string) ([]string, error) {
						return packageList, nil
					},
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
				}
				return newTestEngine(src, dest, fs, nil)
			},
			args: []string{"pkg2"},
			want: []string{"pkg2"},
		},
		{
			name: "build package list - list dirs error",
			setup: func() *Engine {
				src := filepath.Join("home", "user", "dotfiles")
				dest := filepath.Join("home", "user")
				fs := &MockFileSystem{
					listDirFn: func(parent string) ([]string, error) {
						return nil, os.ErrNotExist
					},
					isDirFn: func(path string) (bool, error) {
						return false, os.ErrNotExist
					},
				}
				return newTestEngine(src, dest, fs, nil)
			},
			wantErr:   true,
			wantErrIs: os.ErrNotExist,
		},
		{
			name: "build package list - list dirs error with args",
			setup: func() *Engine {
				src := filepath.Join("home", "user", "dotfiles")
				dest := filepath.Join("home", "user")
				fs := &MockFileSystem{
					listDirFn: func(parent string) ([]string, error) {
						return nil, os.ErrNotExist
					},
					isDirFn: func(path string) (bool, error) {
						return false, os.ErrNotExist
					},
				}
				return newTestEngine(src, dest, fs, nil)
			},
			args:      []string{"pkg2"},
			wantErr:   true,
			wantErrIs: os.ErrNotExist,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			engine := tc.setup()
			packageList, err := engine.buildPackageList(tc.args)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if !slices.Equal(packageList, tc.want) {
				t.Fatalf("got %v, want %v", packageList, tc.want)
			}
		})
	}
}

func TestPackageOperations_retrieveAllPackages(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		want      []string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "retrieve all packages",
			setup: func() *Engine {
				src := filepath.Join("home", "user", "dotfiles")
				dest := filepath.Join("home", "user")
				dirs := []string{filepath.Join(src, "pkg1"), filepath.Join(src, "pkg2"), filepath.Join(src, "pkg3")}
				fs := &MockFileSystem{
					listDirFn: func(parent string) ([]string, error) {
						return dirs, nil
					},
				}
				return newTestEngine(src, dest, fs, nil)
			},
			want: []string{"pkg1", "pkg2", "pkg3"},
		},
		{
			name: "retrieve all packages - list dir error",
			setup: func() *Engine {
				fs := &MockFileSystem{
					listDirFn: func(parent string) ([]string, error) {
						return nil, os.ErrNotExist
					},
				}
				return newTestEngine("", "", fs, nil)
			},
			wantErr:   true,
			wantErrIs: os.ErrNotExist,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			engine := tc.setup()
			pkgCandidates, err := engine.retrieveAllPackages()
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if !slices.Equal(pkgCandidates, tc.want) {
				t.Fatalf("got %v, want %v", pkgCandidates, tc.want)
			}
		})
	}
}

func TestPackageOperations_retrievePackagesFromArgs(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Engine
		args      []string
		want      []string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "retrieve packages from args",
			setup: func() *Engine {
				fs := &MockFileSystem{
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
				}
				return newTestEngine("", "", fs, nil)
			},
			want: []string{"pkg1", "pkg2", "pkg3"},
			args: []string{"pkg1", "pkg2", "pkg3"},
		},
		{
			name: "retrieve packages from args - root package",
			setup: func() *Engine {
				fs := &MockFileSystem{
					isDirFn: func(path string) (bool, error) {
						return true, nil
					},
				}
				return newTestEngine("", "", fs, nil)
			},
			args:      []string{"pkg1", ".", "pkg3"},
			wantErr:   true,
			wantErrIs: ErrRootIsNotPkg,
		},
		{
			name: "retrieve packages from args - not dir",
			setup: func() *Engine {
				fs := &MockFileSystem{
					isDirFn: func(path string) (bool, error) {
						return false, nil
					},
				}
				return newTestEngine("", "", fs, nil)
			},
			args:      []string{"pkg1", "pkg3"},
			wantErr:   true,
			wantErrIs: ErrPkgIsNotDir,
		},
		{
			name: "retrieve packages from args - dir checking fail",
			setup: func() *Engine {
				fs := &MockFileSystem{
					isDirFn: func(path string) (bool, error) {
						return false, os.ErrPermission
					},
				}
				return newTestEngine("", "", fs, nil)
			},
			args:      []string{"pkg1", "pkg3"},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			engine := tc.setup()
			pkgCandidates, err := engine.retrievePackagesFromArgs(tc.args)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if !slices.Equal(pkgCandidates, tc.want) {
				t.Fatalf("got %v, want %v", pkgCandidates, tc.want)
			}
		})
	}
}

func TestPackageOperations_filterPackages(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Engine
		want       []string
		candidates []string
		wantErr    bool
		wantErrIs  error
	}{
		{
			name: "filter pacakges",
			setup: func() *Engine {
				fs := &MockFileSystem{}
				ignoreList := newTestIgnoreList(fs, slog.New(slog.NewTextHandler(io.Discard, nil)), []string{"docs"})
				return newTestEngine("", "", fs, ignoreList)
			},
			candidates: []string{"nvim", "zsh", "docs", "mydocs"},
			want:       []string{"nvim", "zsh", "mydocs"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			engine := tc.setup()
			filtered, err := engine.filterPackages(tc.candidates)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if !slices.Equal(filtered, tc.want) {
				t.Fatalf("got %v, want %v", filtered, tc.want)
			}
		})
	}
}
