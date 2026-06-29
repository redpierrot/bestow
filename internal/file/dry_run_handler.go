/*
All Rights Reversed (ɔ)
*/

package file

import (
	"log/slog"
	"path/filepath"
)

// DryRunHandler is an implementation of engine.FileSystem, without any actual write operations.
type DryRunHandler struct {
	readHandler
	createdDirs map[string]bool
}

// NewDryRunHandler returns a new DryRunHandler with the provided logger l.
func NewDryRunHandler(l *slog.Logger) *DryRunHandler {
	return &DryRunHandler{
		createdDirs: make(map[string]bool),
		readHandler: readHandler{logger: l},
	}
}

// CreateFile creates a file in the provided path and writes the provided content to the file.
func (h *DryRunHandler) CreateFile(path, content string) error {
	h.logger.Debug("writing to file", "file", path)
	if err := isWritable(path); err != nil {
		return err
	}
	h.logger.Debug("successfully written to file", "path", path)
	return nil
}

// CreateDir creates a directory on the provided path, including all the parent directories.
func (h *DryRunHandler) CreateDir(path string) error {
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
	if err := isWritable(path); err != nil {
		return err
	}
	h.createdDirs[path] = true
	h.logger.Debug("created directory", "path", path)
	return nil
}

// Link creates a symlink of a provided src in the provided target.
// If the target directory does not exist, link will create all the parent directories.
func (h *DryRunHandler) Link(src, target string) error {
	h.logger.Debug("creating symlink", "source", src, "target", target)
	destParent := filepath.Dir(target)
	if err := isWritable(destParent); err != nil {
		return err
	}
	h.logger.Debug("link created", "source", src, "target", target)
	return nil
}

// Move moves a file from src to target
// If the target directory does not exist, move will create all the parent directories.
func (h *DryRunHandler) Move(src, target string) error {
	h.logger.Debug("moving file", "source", src, "target", target)
	destParent := filepath.Dir(target)
	if err := isWritable(destParent); err != nil {
		return err
	}
	h.logger.Debug("moved file", "from", src, "to", target)
	return nil
}

// Remove removes the file in the provided path.
func (h *DryRunHandler) Remove(path string) error {
	h.logger.Debug("removing the file", "path", path)
	exists, err := h.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		h.logger.Warn("file does not exist", "operation", "remove", "file", path)
		return nil
	}
	if err := isWritable(path); err != nil {
		return err
	}
	h.logger.Debug("successfully removed the file", "file_name", path)
	return nil
}
