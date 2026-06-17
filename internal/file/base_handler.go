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

	"golang.org/x/sys/unix"
)

// baseHandler is the implementation of the System using io, os, and bufio go modules.
type baseHandler struct {
	logger *slog.Logger
}

// ListDirs lists all the subdirectories in a given parent directory.
// The list contains the full path of the subdirectories found.
func (h *baseHandler) ListDirs(parent string) ([]string, error) {
	h.logger.Debug("listing all the directories", "source", parent)
	isDir, err := h.IsDir(parent)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, fmt.Errorf("listDirs %s: %w", parent, ErrNotDir)
	}
	files, err := os.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(files))
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
func (h *baseHandler) ListAllFiles(parent string) ([]string, error) {
	h.logger.Debug("listing all the files in the directory", "parent", parent)
	isDir, err := h.IsDir(parent)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, fmt.Errorf("listFiles %s: %w", parent, ErrNotDir)
	}
	var result []string
	err = filepath.WalkDir(parent, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
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

// IsDir checks whether the provided path is a directory.
func (h *baseHandler) IsDir(path string) (bool, error) {
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
func (h *baseHandler) IsEmptyDir(path string) (bool, error) {
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
func (h *baseHandler) Exists(path string) (bool, error) {
	h.logger.Debug("checking path exists", "path", path)
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
func (h *baseHandler) ReadLines(path string) ([]string, error) {
	h.logger.Debug("reading the file", "path", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var result []string
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
func (h *baseHandler) GetExistingFileType(src, dst string) (ExistingType, error) {
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
		if errors.Is(err, os.ErrNotExist) {
			return ExistingForeignSymlink, nil
		}
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

// Readable returns whether a provided path is readable
func (h *baseHandler) Readable(path string) error {
	if err := unix.Access(path, unix.R_OK); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return nil
}

// Writable returns whether a provided path is writable
func (h *baseHandler) Writable(path string) error {
	if err := unix.Access(path, unix.W_OK); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
