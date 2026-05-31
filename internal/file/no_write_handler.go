/*
All Rights Reversed (ɔ)
*/

package file

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

// NoWriteHandler is the implementation of the System using io, os, and bufio go modules.
type NoWriteHandler struct {
	// Should use mutex here if migrate to go routines
	createdDirs map[string]bool
	logger      *slog.Logger
}

// NewNoWriteHandler returns a new NoWriteHandler with the provided logger l.
func NewNoWriteHandler(l *slog.Logger) *NoWriteHandler {
	return &NoWriteHandler{
		createdDirs: make(map[string]bool),
		logger:      l.With("component", "file"),
	}
}

// ListFiles returns a list of all the files in a given parent directory, excluding the directories.
// The file list includes the full paths of the files found.
func (h *NoWriteHandler) ListFiles(parent string) ([]string, error) {
	isDir, err := h.IsDir(parent)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, fmt.Errorf("stat %s: %w", parent, ErrNotDir)
	}
	files, err := os.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		result = append(result, filepath.Join(parent, file.Name()))
	}
	h.logger.Debug("found files in the directory", "directory", parent, "files", result)
	return result, nil
}

// ListDirs lists all the subdirectories in a given parent directory.
// The list contains the full path of the subdirectories found.
func (h *NoWriteHandler) ListDirs(parent string) ([]string, error) {
	h.logger.Debug("listing all the directories", "source", parent)
	isDIr, err := h.IsDir(parent)
	if err != nil {
		return nil, err
	}
	if !isDIr {
		return nil, fmt.Errorf("listDirs %s: %w", parent, ErrNotDir)
	}
	files, err := os.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			result = append(result, filepath.Join(parent, file.Name()))
		}
	}
	h.logger.Debug("finished searching directories", "dirs", result)
	return result, nil
}

// ListAllFiles lists all the files in a given parent directory, including the files in the subdirectories.
// The returned list contains the full path of the files found.
func (h *NoWriteHandler) ListAllFiles(parent string) ([]string, error) {
	h.logger.Debug("listing all the files in the directory", "parent", parent)
	isDir, err := h.IsDir(parent)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, fmt.Errorf("listFiles %s: %w", parent, ErrNotDir)
	}
	result := []string{}

	err = filepath.WalkDir(parent, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateFile creates a file in the provided path and writes the provided content to the file.
func (h *NoWriteHandler) CreateFile(path, content string) error {
	h.logger.Debug("writing to file", "file", path)
	h.logger.Debug("successfully written to file", "path", path)
	return nil
}

// CreateDir creates a directory on the provided path, including all the parent directories.
func (h *NoWriteHandler) CreateDir(path string) error {
	h.logger.Debug("creating directory", "path", path)
	if h.createdDirs[path] {
		h.logger.Debug("directory already created", "path", path)
		return nil
	}
	exists, err := h.IsDir(path)
	if err != nil {
		return err
	}
	if exists {
		h.logger.Debug("directory already exists", "path", path)
		h.createdDirs[path] = true
		return nil
	}
	h.createdDirs[path] = true
	h.logger.Debug("created directory", "path", path)
	return nil
}

// Link creates a symlink of a provided src in the provided target.
// If the target directory does not exist, link will create all the parent directories.
func (h *NoWriteHandler) Link(src, target string) error {
	h.logger.Debug("creating symlink", "source", src, "target", target)
	destParent := filepath.Dir(target)
	if err := h.CreateDir(destParent); err != nil {
		return err
	}
	h.logger.Debug("link created", "source", src, "target", target)
	return nil
}

// Move moves a file from src to target
// If the target directory does not exist, move will create all the parent directories.
func (h *NoWriteHandler) Move(src, target string) error {
	h.logger.Debug("moving file", "source", src, "target", target)
	destParent := filepath.Dir(target)
	if err := h.CreateDir(destParent); err != nil {
		return err
	}
	h.logger.Debug("moved file", "from", src, "to", target)
	return nil
}

// Remove removes the file in the provided path.
func (h *NoWriteHandler) Remove(path string) error {
	h.logger.Debug("removing the file", "path", path)
	exists, err := h.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		h.logger.Warn("file does not exist", "operation", "remove", "file", path)
		return nil
	}
	h.logger.Debug("successfully removed the file", "file_name", path)
	return nil
}

// IsDir checks whether the provided path is a directory.
func (h *NoWriteHandler) IsDir(path string) (bool, error) {
	h.logger.Debug("checking whether the path is a directory", "path", path)
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return stat.IsDir(), nil
}

// IsEmptyDir returns true if the provided path is empty. It will return an error is the provided path is not a
// directory.
func (h *NoWriteHandler) IsEmptyDir(path string) (bool, error) {
	h.logger.Debug("checking if the provided path is an empty directory", "path", path)
	isDir, err := h.IsDir(path)
	if err != nil {
		return false, err
	}
	if !isDir {
		return false, fmt.Errorf("read %s: %w", path, ErrNotDir)
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.ReadDir(1)
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	return false, nil
}

// Exists returns true if the provided path exists.
func (h *NoWriteHandler) Exists(path string) (bool, error) {
	h.logger.Debug("checking whether the provided path exists", "path", path)
	_, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ReadLines reads the content of a file in the given path and returns the lines of text as a list of strings.
func (h *NoWriteHandler) ReadLines(path string) ([]string, error) {
	h.logger.Debug("reading the file", "path", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return result, nil
}

// GetExistingFileType returns the type of the existing dst compared to the provided src. Possible values are:
//   - ExistingRegularFile: dst is a regular file
//   - ExistingManagedSymlink: dst is a symlink that is managed by bestow
//   - ExistingForeignSymlink: dst is a symlink that is not managed by bestow
//   - ExistingDir: dst is a directory
func (h *NoWriteHandler) GetExistingFileType(src, dst string) (ExistingType, error) {
	h.logger.Debug("checking existing file type", "source", src, "destination", dst)
	lstat, err := os.Lstat(dst)
	if err != nil {
		return ExistingUnknown, err
	}
	if lstat.Mode().IsRegular() {
		h.logger.Debug("found regular file")
		return ExistingRegularFile, nil
	}
	if lstat.IsDir() {
		h.logger.Debug("found directory", "path", dst)
		return ExistingDir, nil
	}

	h.logger.Debug("found symlink", "path", dst)
	srcInfo, err := os.Stat(src)
	if err != nil {
		return ExistingUnknown, err
	}
	destInfo, err := os.Stat(dst)
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			return ExistingForeignSymlink, nil
		}
		return ExistingUnknown, err
	}
	if os.SameFile(srcInfo, destInfo) {
		h.logger.Debug("found managed symlink", "source", src, "destination", dst)
		return ExistingManagedSymlink, nil
	}
	return ExistingForeignSymlink, nil
}
