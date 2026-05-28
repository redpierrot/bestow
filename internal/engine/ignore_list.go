/*
All Rights Reversed (ɔ)
*/

package engine

import (
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

func newIgnoreList(src string, l *slog.Logger) (*IgnoreList, error) {
	handler := file.NewHandler(l)
	list := &IgnoreList{src: src, fileSystem: handler, logger: l.With("section", "ignore_handler")}

	// Load global ignore list
	configHome := config.AppConfigHome()
	if err := readIgnoreFile(configHome, &list.items, handler); err != nil {
		return nil, err
	}

	//load source ignore list
	if err := readIgnoreFile(src, &list.items, handler); err != nil {
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
		return nil, &EngineError{
			Message: "failed to read the ignore list for the package",
			Cause:   err,
		}
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
			return false, &EngineError{
				Message: "ignore pattern validation error",
				Cause:   err,
			}
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}
