/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

type IgnoreList struct {
	src          string
	items        []string
	fileSystem   IgnoreReader
	logger       *slog.Logger
	packageLists map[string][]string
}

type IgnoreReader interface {
	Exists(path string) (bool, error)
	ReadLines(path string) ([]string, error)
}

func newIgnoreList(src, configHome string, reader IgnoreReader, l *slog.Logger) (*IgnoreList, error) {
	list := &IgnoreList{src: src, fileSystem: reader, logger: l.With("section", "ignore_handler"), packageLists: make(map[string][]string)}

	// Load global ignore list
	list.logger.Debug("reading global ignore file", "path", configHome)
	ignoreItems, err := readIgnoreFile(configHome, reader)
	if err != nil {
		return nil, err
	}

	//load source ignore list
	list.logger.Debug("readong source ignore file", "path", src)
	items, err := readIgnoreFile(src, reader)
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
	pkgPath := filepath.Join(i.src, pkg)
	i.logger.Debug("read package ignore file", "path", pkgPath)
	result, err := readIgnoreFile(pkgPath, i.fileSystem)
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
