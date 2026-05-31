/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/file"
	"github.com/bmatcuk/doublestar"
)

type IgnoreList struct {
	src        string
	items      []string
	fileSystem file.System
	logger     *slog.Logger
}

func newIgnoreList(src string, fs file.System, l *slog.Logger) (*IgnoreList, error) {
	list := &IgnoreList{src: src, fileSystem: fs, logger: l.With("section", "ignore_handler")}

	// Load global ignore list
	configHome := config.AppConfigHome()
	if err := readIgnoreFile(configHome, &list.items, fs); err != nil {
		return nil, err
	}

	//load source ignore list
	if err := readIgnoreFile(src, &list.items, fs); err != nil {
		return nil, err
	}
	return list, nil
}

func (i *IgnoreList) forPackage(pkg string) ([]string, error) {
	i.logger.Debug("getting ignore list for package", "package", pkg)
	if pkg == "" {
		return i.items, nil
	}
	result := append([]string(nil), i.items...)
	if err := readIgnoreFile(filepath.Join(i.src, pkg), &result, i.fileSystem); err != nil {
		return nil, err
	}
	return result, nil
}

func (i *IgnoreList) shouldIgnore(fileName, pkg string) (bool, error) {
	i.logger.Debug("checking ignorability", "package", pkg, "file_name", fileName)
	ignoreList, err := i.forPackage(pkg)
	if err != nil {
		return false, err
	}
	for _, ignoreItem := range ignoreList {
		match, err := doublestar.PathMatch(ignoreItem, fileName)
		if err != nil {
			return false, fmt.Errorf("parse %s %s: %w", ignoreItem, fileName, err)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}
