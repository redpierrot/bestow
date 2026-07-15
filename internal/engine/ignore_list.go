/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type IgnoreList struct {
	src          string
	items        []string
	reader       IgnoreReader
	logger       *slog.Logger
	packageLists map[string][]string
}

type IgnoreReader interface {
	Exists(path string) (bool, error)
	ReadLines(path string) ([]string, error)
}

func newIgnoreList(src, configHome string, reader IgnoreReader, l *slog.Logger) (*IgnoreList, error) {
	list := &IgnoreList{src: src, reader: reader, logger: l.With("section", "ignore_handler"), packageLists: make(map[string][]string)}

	// Load global ignore list
	list.logger.Debug("reading global ignore file", "path", configHome)
	ignoreItems, err := readIgnoreFile(configHome, reader)
	if err != nil {
		return nil, err
	}

	//load source ignore list
	list.logger.Debug("reading source ignore file", "path", src)
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
	if list, ok := i.packageLists[pkg]; ok {
		return list, nil
	}
	pkgPath := filepath.Join(i.src, pkg)
	i.logger.Debug("read package ignore file", "path", pkgPath)
	pkgItems, err := readIgnoreFile(pkgPath, i.reader)
	if err != nil {
		return nil, err
	}
	// Avoid mutating i.items
	pkgIgnoreList := make([]string, 0, len(i.items)+len(pkgItems))
	pkgIgnoreList = append(pkgIgnoreList, i.items...)
	pkgIgnoreList = append(pkgIgnoreList, pkgItems...)
	i.packageLists[pkg] = pkgIgnoreList
	return pkgIgnoreList, nil
}

func (i *IgnoreList) isIgnoredFile(path, pkg string) (bool, error) {
	i.logger.Debug("checking ignorability", "package", pkg, "file_path", path)
	ignoreList, err := i.forPackage(pkg)
	if err != nil {
		return false, err
	}
	for _, ignoreItem := range ignoreList {
		match := doublestar.MatchUnvalidated(ignoreItem, path)
		if match {
			return true, nil
		}
		if name := filepath.Base(path); name != path {
			match = doublestar.MatchUnvalidated(ignoreItem, name)
			if match {
				return true, nil
			}
		}
	}
	return false, nil
}

func (i *IgnoreList) isIgnored(pkg string) bool {
	for _, ignoreItem := range i.items {
		match := doublestar.MatchUnvalidated(ignoreItem, pkg)
		if match {
			return true
		}
	}
	return false
}

func readIgnoreFile(source string, reader IgnoreReader) ([]string, error) {
	ignoreFile := filepath.Join(source, ignoreFileName)
	exists, err := reader.Exists(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ignoreFile, err)
	}
	if !exists {
		return nil, nil
	}
	lines, err := reader.ReadLines(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ignoreFile, err)
	}
	patterns := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !doublestar.ValidatePattern(trimmed) {
			return nil, fmt.Errorf("parse %s: %w", trimmed, errInvalidPattern)
		}
		patterns = append(patterns, trimmed)
	}
	return patterns, nil
}
