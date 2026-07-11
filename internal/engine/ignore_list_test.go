/*
All Rights Reversed (t *testing.T)
*/

package engine

import (
	"os"
	"slices"
	"strings"
	"testing"
)

type mockIgnoreReader struct {
	existsFn    func(path string) (bool, error)
	readLinesFn func(path string) ([]string, error)
}

func (m *mockIgnoreReader) Exists(path string) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(path)
	}
	return false, nil
}

func (m *mockIgnoreReader) ReadLines(path string) ([]string, error) {
	if m.readLinesFn != nil {
		return m.readLinesFn(path)
	}
	return nil, nil
}

func Test_newIgnoreList(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		configHome string
		reader     *mockIgnoreReader
		wantErr    bool
		wantErrIs  error
	}{
		{
			name:   "no errors",
			reader: &mockIgnoreReader{},
		},
		{
			name:       "global file error",
			configHome: "config_home",
			reader: &mockIgnoreReader{
				existsFn: func(path string) (bool, error) {
					if strings.Contains(path, "config_home") {
						return false, os.ErrPermission
					}
					return true, nil
				},
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "source file error",
			src:  "src",
			reader: &mockIgnoreReader{
				existsFn: func(path string) (bool, error) {
					if strings.Contains(path, "src") {
						return false, os.ErrPermission
					}
					return true, nil
				},
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := newTestLogger()
			_, err := newIgnoreList(tc.src, tc.configHome, tc.reader, l)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
		})
	}
}

func Test_forPackage(t *testing.T) {
}

func Test_isIgnoredFile(t *testing.T) {
}

func Test_isIgnored(t *testing.T) {
}

func TestReadIgnoreFile(t *testing.T) {
	tests := []struct {
		name      string
		reader    *mockIgnoreReader
		want      []string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "file not exist",
			reader: &mockIgnoreReader{
				existsFn: func(path string) (bool, error) {
					return false, nil
				},
			},
		},
		{
			name: "file read error",
			reader: &mockIgnoreReader{
				existsFn: func(path string) (bool, error) {
					return false, os.ErrPermission
				},
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
		{
			name: "file with no comments",
			reader: &mockIgnoreReader{
				existsFn: func(path string) (bool, error) {
					return true, nil
				},
				readLinesFn: func(path string) ([]string, error) {
					return []string{".git"}, nil
				},
			},
			want: []string{".git"},
		},
		{
			name: "file with comments",
			reader: &mockIgnoreReader{
				existsFn: func(path string) (bool, error) {
					return true, nil
				},
				readLinesFn: func(path string) ([]string, error) {
					return []string{".git", "  # This is a comment", "     nvim", "# Another Comment", "this line is not # Comment"}, nil
				},
			},
			want: []string{".git", "nvim", "this line is not # Comment"},
		},
		{
			name: "read lines error",
			reader: &mockIgnoreReader{
				existsFn: func(path string) (bool, error) {
					return true, nil
				},
				readLinesFn: func(path string) ([]string, error) {
					return nil, os.ErrPermission
				},
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := readIgnoreFile("file_path", tc.reader)
			if validateErrScenario(t, tc.wantErr, err, tc.wantErrIs) {
				return
			}
			if !slices.Equal(lines, tc.want) {
				t.Fatalf("got %v, want %v", lines, tc.want)
			}
		})
	}
}
