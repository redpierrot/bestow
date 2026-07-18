/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/redpierrot/bestow/internal/file"
)

func newTestEngine(fs *mockFileSystem, ignoreList *IgnoreList) *Engine {
	l := newTestLogger()
	if ignoreList == nil {
		ignoreList = newTestIgnoreList(fs, l, nil)
	}
	return &Engine{
		logger:     l,
		fileSystem: fs,
		ignore:     ignoreList,
	}
}

func newTestIgnoreList(fs IgnoreReader, logger *slog.Logger, items []string) *IgnoreList {
	if items == nil {
		items = make([]string, 0)
	}
	return &IgnoreList{
		src:          "",
		items:        items,
		reader:       fs,
		logger:       logger,
		packageLists: make(map[string][]string),
	}
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

func testPackageList(parent string) []string {
	return []string{filepath.Join(parent, "pkg1"), filepath.Join(parent, "pkg2"), filepath.Join(parent, "pkg3")}
}

func candidate(src, dest string) operationCandidate {
	return operationCandidate{
		source:      src,
		destination: dest,
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func isSameAction(f1, f2 fileAction) bool {
	return f1.kind() == f2.kind()
}

type mockFileSystem struct {
	listDirFn          func(parent string) ([]string, error)
	listAllFilesFn     func(parent string) ([]string, error)
	createFileFn       func(path, content string) error
	createDirFn        func(path string) error
	linkFn             func(src, target string) error
	moveFn             func(src, target string) error
	removeFn           func(path string) error
	isDirFn            func(path string) (bool, error)
	isEmptyDirFn       func(path string) (bool, error)
	existsFn           func(path string) (bool, error)
	readLinesFn        func(path string) ([]string, error)
	existingFileTypeFn func(src, dest string) (file.ExistingType, error)
}

func (mf *mockFileSystem) ListDirs(parent string) ([]string, error) {
	if mf.listDirFn != nil {
		return mf.listDirFn(parent)
	}
	return nil, nil
}

func (mf *mockFileSystem) ListAllFiles(parent string) ([]string, error) {
	if mf.listAllFilesFn != nil {
		return mf.listAllFilesFn(parent)
	}
	return nil, nil
}

func (mf *mockFileSystem) CreateFile(path, content string) error {
	if mf.createFileFn != nil {
		return mf.createFileFn(path, content)
	}
	return nil
}

func (mf *mockFileSystem) CreateDir(path string) error {
	if mf.createDirFn != nil {
		return mf.createDirFn(path)
	}
	return nil
}

func (mf *mockFileSystem) Link(src, target string) error {
	if mf.linkFn != nil {
		return mf.linkFn(src, target)
	}
	return nil
}

func (mf *mockFileSystem) Move(src, target string) error {
	if mf.moveFn != nil {
		return mf.moveFn(src, target)
	}
	return nil
}

func (mf *mockFileSystem) Remove(path string) error {
	if mf.removeFn != nil {
		return mf.removeFn(path)
	}
	return nil
}

func (mf *mockFileSystem) IsDir(path string) (bool, error) {
	if mf.isDirFn != nil {
		return mf.isDirFn(path)
	}
	return false, nil
}

func (mf *mockFileSystem) IsEmptyDir(path string) (bool, error) {
	if mf.isEmptyDirFn != nil {
		return mf.isEmptyDirFn(path)
	}
	return true, nil
}

func (mf *mockFileSystem) Exists(path string) (bool, error) {
	if mf.existsFn != nil {
		return mf.existsFn(path)
	}
	return false, nil
}

func (mf *mockFileSystem) ReadLines(path string) ([]string, error) {
	if mf.readLinesFn != nil {
		return mf.readLinesFn(path)
	}
	return make([]string, 0), nil
}

func (mf *mockFileSystem) ExistingFileType(src, dest string) (file.ExistingType, error) {
	if mf.existingFileTypeFn != nil {
		return mf.existingFileTypeFn(src, dest)
	}
	return file.ExistingUnknown, nil
}

func TestSummary(t *testing.T) {
	tests := []struct {
		name       string
		summary    *Summary
		action     ActionKind
		wantCount  int
		wantRevert int
	}{
		{
			name: "stow",
			summary: &Summary{
				counts:   [numActionKinds]int{ActionLink: 1},
				reverted: 5,
			},
			action:     ActionLink,
			wantCount:  1,
			wantRevert: 5,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count := tc.summary.Count(tc.action)
			if count != tc.wantCount {
				t.Fatalf("got %d, want %d", count, tc.wantCount)
			}
			reverted := tc.summary.Reverted()
			if reverted != tc.wantRevert {
				t.Fatalf("got reverted %d, want %d", reverted, tc.wantRevert)
			}
		})
	}
}
