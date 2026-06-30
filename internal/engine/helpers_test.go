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

func newTestEngine(src, dest string, fs *MockFileSystem, ignoreList *IgnoreList) *Engine {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if ignoreList == nil {
		ignoreList = newTestIgnoreList(fs, logger, nil)
	}
	return &Engine{
		logger:      logger,
		source:      src,
		destination: dest,
		fileSystem:  fs,
		ignore:      ignoreList,
	}
}

func newTestIgnoreList(fs *MockFileSystem, logger *slog.Logger, items []string) *IgnoreList {
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
	if wantErr && !errors.Is(err, wantErrIs) {
		t.Fatalf("error got %v, want %v", err, wantErrIs)
	}
	return wantErr
}

func getSamplePackageList(parent string) []string {
	return []string{filepath.Join(parent, "pkg1"), filepath.Join(parent, "pkg2"), filepath.Join(parent, "pkg3")}
}

type MockFileSystem struct {
	listDirFn          func(parent string) ([]string, error)
	listAllFilesFn     func(parent string) ([]string, error)
	createFileFn       func(path, content string) error
	createDirFn        func(path string) error
	linkFn             func(src, target string) error
	moveFn             func(src, target string) error
	removeFn           func(path string) error
	isDirFn            func(path string) (bool, error)
	existsFn           func(path string) (bool, error)
	readLinesFn        func(path string) ([]string, error)
	existingFileTypeFn func(src, dest string) (file.ExistingType, error)
}

func (mf *MockFileSystem) ListDirs(parent string) ([]string, error) {
	if mf.listDirFn != nil {
		return mf.listDirFn(parent)
	}
	return nil, nil
}

func (mf *MockFileSystem) ListAllFiles(parent string) ([]string, error) {
	if mf.listAllFilesFn != nil {
		return mf.listAllFilesFn(parent)
	}
	return nil, nil
}

func (mf *MockFileSystem) CreateFile(path, content string) error {
	if mf.createFileFn != nil {
		return mf.createFileFn(path, content)
	}
	return nil
}

func (mf *MockFileSystem) CreateDir(path string) error {
	if mf.createDirFn != nil {
		return mf.createDirFn(path)
	}
	return nil
}

func (mf *MockFileSystem) Link(src, target string) error {
	if mf.linkFn != nil {
		return mf.linkFn(src, target)
	}
	return nil
}

func (mf *MockFileSystem) Move(src, target string) error {
	if mf.moveFn != nil {
		return mf.moveFn(src, target)
	}
	return nil
}

func (mf *MockFileSystem) Remove(path string) error {
	if mf.removeFn != nil {
		return mf.removeFn(path)
	}
	return nil
}

func (mf *MockFileSystem) IsDir(path string) (bool, error) {
	if mf.isDirFn != nil {
		return mf.isDirFn(path)
	}
	return false, nil
}

func (mf *MockFileSystem) IsEmptyDir(path string) (bool, error) {
	return true, nil
}

func (mf *MockFileSystem) Exists(path string) (bool, error) {
	if mf.existsFn != nil {
		return mf.existsFn(path)
	}
	return false, nil
}

func (mf *MockFileSystem) ReadLines(path string) ([]string, error) {
	if mf.readLinesFn != nil {
		return mf.readLinesFn(path)
	}
	return make([]string, 0), nil
}

func (mf *MockFileSystem) ExistingFileType(src, dest string) (file.ExistingType, error) {
	if mf.existingFileTypeFn != nil {
		return mf.existingFileTypeFn(src, dest)
	}
	return file.ExistingUnknown, nil
}
