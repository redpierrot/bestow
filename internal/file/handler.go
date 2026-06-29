/*
All Rights Reversed (ɔ)
*/

package file

import (
	"log/slog"
	"os"
	"path/filepath"
)

// TODO: Make this configurable or make it passable as a parameter
const (
	permDirWrite  = 0o755
	permFileWrite = 0o644
)

// Handler is the implementation of the System using io, os, and bufio go modules.
type Handler struct {
	readHandler
	// Should use mutex here if migrate to goroutines
	createdDirs map[string]bool
}

// NewHandler returns a new Handler with the provided logger l.
func NewHandler(l *slog.Logger) *Handler {
	return &Handler{
		createdDirs: make(map[string]bool),
		readHandler: readHandler{logger: l},
	}
}

// CreateFile creates a file in the provided path and writes the provided content to the file.
func (h *Handler) CreateFile(path, content string) error {
	h.logger.Debug("writing to file", "file", path)
	if err := os.WriteFile(path, []byte(content), permFileWrite); err != nil {
		return err
	}
	h.logger.Debug("successfully written to file", "path", path)
	return nil
}

// CreateDir creates a directory on the provided path, including all the parent directories.
func (h *Handler) CreateDir(path string) error {
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
	if err := os.MkdirAll(path, permDirWrite); err != nil {
		return err
	}
	h.createdDirs[path] = true
	h.logger.Debug("created directory", "path", path)
	return nil
}

// Link creates a symlink of a provided src in the provided target.
// If the target directory does not exist, link will create all the parent directories.
func (h *Handler) Link(src, target string) error {
	h.logger.Debug("creating symlink", "source", src, "target", target)
	destParent := filepath.Dir(target)
	if err := h.CreateDir(destParent); err != nil {
		return err
	}
	if err := os.Symlink(src, target); err != nil {
		return err
	}
	h.logger.Debug("link created", "source", src, "target", target)
	return nil
}

// Move moves a file from src to target
// If the target directory does not exist, move will create all the parent directories.
func (h *Handler) Move(src, target string) error {
	h.logger.Debug("moving file", "source", src, "target", target)
	destParent := filepath.Dir(target)
	if err := h.CreateDir(destParent); err != nil {
		return err
	}
	if err := os.Rename(src, target); err != nil {
		return err
	}
	return nil
}

// Remove removes the file in the provided path.
func (h *Handler) Remove(path string) error {
	h.logger.Debug("removing the file", "path", path)
	exists, err := h.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		h.logger.Warn("file does not exist", "operation", "remove", "file", path)
		return nil
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	h.logger.Debug("successfully removed the file", "file_name", path)
	return nil
}
