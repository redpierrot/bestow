/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Init(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, cfgHome string) *Engine
		cfg        *InitConfig
		configHome string
		want       ExecuteResult
		wantErr    bool
		wantErrIs  error
		wantErrAs  func(*testing.T, error, string)
	}{
		{
			name: "no errors",
			setup: func(t *testing.T, cfgHome string) *Engine {
				e := newTestEngine(&mockFileSystem{}, nil)
				e.configHome = cfgHome
				return e
			},
			configHome: "config_home",
			cfg: &InitConfig{
				Force:      false,
				ConfigFile: "config_file",
			},
			want: ExecuteResult{
				Events: []ActionEvent{
					{
						Action:    fileOpCreated,
						Msg:       filepath.Join("config_home", "config_file"),
						EventType: EventSuccess,
					},
					{
						Action:    fileOpCreated,
						Msg:       filepath.Join("config_home", ignoreFileName),
						EventType: EventSuccess,
					},
				},
			},
		},
		{
			name: "directory creation fail",
			setup: func(t *testing.T, cfgHome string) *Engine {
				mf := &mockFileSystem{
					createDirFn: func(path string) error {
						return os.ErrPermission
					},
				}
				e := newTestEngine(mf, nil)
				e.configHome = cfgHome
				return e
			},
			configHome: "config_home",
			cfg: &InitConfig{
				Force:      false,
				ConfigFile: "config_file",
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "directory creation fail",
			setup: func(t *testing.T, cfgHome string) *Engine {
				mf := &mockFileSystem{
					createDirFn: func(path string) error {
						return os.ErrPermission
					},
				}
				e := newTestEngine(mf, nil)
				e.configHome = cfgHome
				return e
			},
			configHome: "config_home",
			cfg: &InitConfig{
				Force:      false,
				ConfigFile: "config_file",
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "config file creation fail",
			setup: func(t *testing.T, cfgHome string) *Engine {
				mf := &mockFileSystem{
					createFileFn: func(path, content string) error {
						if strings.Contains(path, "config_file") {
							return os.ErrPermission
						}
						return nil
					},
				}
				e := newTestEngine(mf, nil)
				e.configHome = cfgHome
				return e
			},
			configHome: "config_home",
			cfg: &InitConfig{
				Force:      false,
				ConfigFile: "config_file",
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "ignore file creation fail",
			setup: func(t *testing.T, cfgHome string) *Engine {
				mf := &mockFileSystem{
					createFileFn: func(path, content string) error {
						if strings.Contains(path, ".bestowignore") {
							return os.ErrPermission
						}
						return nil
					},
				}
				e := newTestEngine(mf, nil)
				e.configHome = cfgHome
				return e
			},
			configHome: "config_home",
			cfg: &InitConfig{
				Force:      false,
				ConfigFile: "config_file",
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "ignore file creation with ignore list",
			setup: func(t *testing.T, cfgHome string) *Engine {
				mf := &mockFileSystem{
					createFileFn: func(path, content string) error {
						if strings.Contains(path, ".bestowignore") {
							return os.ErrPermission
						}
						return nil
					},
				}
				e := newTestEngine(mf, nil)
				e.configHome = cfgHome
				return e
			},
			configHome: "config_home",
			cfg: &InitConfig{
				Force:      false,
				ConfigFile: "config_file",
				IgnoreList: []string{".git", "*.md"},
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup(t, tc.configHome)
			executeResult, err := e.Init(tc.cfg)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if executeResult.DryRun != tc.want.DryRun {
				t.Fatalf("got dry run %v, want %v", executeResult.DryRun, tc.want.DryRun)
			}
			if executeResult.Summary != tc.want.Summary {
				t.Fatalf("got summary %v, want %v", executeResult.Summary, tc.want.Summary)
			}
			if len(executeResult.Events) != len(tc.want.Events) {
				t.Fatalf("got %v, want %v", len(executeResult.Events), len(tc.want.Events))
			}
			for i := range len(executeResult.Events) {
				if executeResult.Events[i] != tc.want.Events[i] {
					t.Fatalf("got %v, want %v", executeResult.Events[i], tc.want.Events[i])
				}
			}
		})
	}
}

func Test_checkExistingFiles(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) Engine
		configFile string
		ignoreFile string
		force      bool
		op         string
		wantErr    bool
		wantErrIs  error
		wantErrAs  func(*testing.T, error, string)
	}{
		{
			name: "no errors",
			setup: func(t *testing.T) Engine {
				mf := &mockFileSystem{}
				return *newTestEngine(mf, nil)
			},
		},
		{
			name: "config check fail",
			setup: func(t *testing.T) Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						if strings.Contains(path, "config") {
							return false, os.ErrPermission
						}
						return true, nil
					},
				}
				return *newTestEngine(mf, nil)
			},
			configFile: "config_file",
			wantErr:    true,
			wantErrIs:  os.ErrPermission,
		},
		{
			name: "ignore check fail",
			setup: func(t *testing.T) Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						if strings.Contains(path, "ignore") {
							return false, os.ErrPermission
						}
						return true, nil
					},
				}
				return *newTestEngine(mf, nil)
			},
			ignoreFile: "ignore_file",
			wantErr:    true,
			wantErrIs:  os.ErrPermission,
		},
		{
			name: "config exist",
			setup: func(t *testing.T) Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						if strings.Contains(path, "config") {
							return true, nil
						}
						return false, nil
					},
				}
				return *newTestEngine(mf, nil)
			},
			configFile: "config_file",
			op:         "exists config_file",
			wantErr:    true,
			wantErrAs: func(t *testing.T, err error, op string) {
				var expected *HintedError
				if !errors.As(err, &expected) {
					t.Fatalf("got %v, want HintedError", err)
				}
				if expected.Op != op {
					t.Fatalf("got %s, want %s", expected, op)
				}
			},
		},
		{
			name: "ignore exist",
			setup: func(t *testing.T) Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						if strings.Contains(path, "ignore") {
							return true, nil
						}
						return false, nil
					},
				}
				return *newTestEngine(mf, nil)
			},
			ignoreFile: "ignore_file",
			op:         "exists ignore_file",
			wantErr:    true,
			wantErrAs: func(t *testing.T, err error, op string) {
				var expected *HintedError
				if !errors.As(err, &expected) {
					t.Fatalf("got %v, want HintedError", err)
				}
				if expected.Op != op {
					t.Fatalf("got %s, want %s", expected, op)
				}
			},
		},
		{
			name: "both exist",
			setup: func(t *testing.T) Engine {
				mf := &mockFileSystem{
					existsFn: func(path string) (bool, error) {
						return true, nil
					},
				}
				return *newTestEngine(mf, nil)
			},
			configFile: "config_file",
			ignoreFile: "ignore_file",
			op:         "exists config_file, ignore_file",
			wantErr:    true,
			wantErrAs: func(t *testing.T, err error, op string) {
				var expected *HintedError
				if !errors.As(err, &expected) {
					t.Fatalf("got %v, want HintedError", err)
				}
				if expected.Op != op {
					t.Fatalf("got %s, want %s", expected, op)
				}
			},
		},
		{
			name:  "force",
			force: true,
			setup: func(t *testing.T) Engine {
				return *newTestEngine(&mockFileSystem{}, nil)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.setup(t)
			err := e.checkExistingFiles(tc.configFile, tc.ignoreFile, tc.force)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				if tc.wantErrAs != nil {
					tc.wantErrAs(t, err, tc.op)
				}
			}
		})
	}
}
