/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/bmatcuk/doublestar/v4"
)

type IgnoreList struct {
	src          string
	items        []string
	fileSystem   FileSystem
	logger       *slog.Logger
	packageLists map[string][]string
}

func newIgnoreList(src string, fs FileSystem, l *slog.Logger) (*IgnoreList, error) {
	list := &IgnoreList{src: src, fileSystem: fs, logger: l.With("section", "ignore_handler"), packageLists: make(map[string][]string)}

	// Load global ignore list
	configHome := config.AppConfigHome()
	ignoreItems, err := readIgnoreFile(configHome, fs)
	if err != nil {
		return nil, err
	}

	//load source ignore list
	items, err := readIgnoreFile(src, fs)
	if err != nil {
		return nil, err
	}
	ignoreItems = append(ignoreItems, items...)
	list.items = ignoreItems
	return list, nil
}

func (i *IgnoreList) forPackage(pkg string) ([]string, error) {
	i.logger.Debug("getting ignore list for package", "package", pkg)
	if pkg == "" {
		return i.items, nil
	}
	if i.packageLists[pkg] != nil {
		return i.packageLists[pkg], nil
	}
	result, err := readIgnoreFile(filepath.Join(i.src, pkg), i.fileSystem)
	if err != nil {
		return nil, err
	}
	// Avoid mutating i.items
	packageList := append(append([]string(nil), i.items...), result...)
	i.packageLists[pkg] = packageList
	return packageList, nil
}

func (i *IgnoreList) shouldIgnorePkgFile(name, pkg string) (bool, error) {
	i.logger.Debug("checking ignorability", "package", pkg, "file_name", name)
	ignoreList, err := i.forPackage(pkg)
	if err != nil {
		return false, err
	}
	for _, ignoreItem := range ignoreList {
		match, err := doublestar.Match(ignoreItem, name)
		if err != nil {
			return false, fmt.Errorf("parse %s %s: %w", ignoreItem, name, err)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (i *IgnoreList) shouldIgnorePkg(pkg string) (bool, error) {
	for _, ignoreItem := range i.items {
		match, err := doublestar.Match(ignoreItem, pkg)
		if err != nil {
			return false, fmt.Errorf("parse %s %s: %w", ignoreItem, pkg, err)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}
