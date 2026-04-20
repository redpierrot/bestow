package engine

import (
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/bmatcuk/doublestar"
)

type IgnoreList struct {
	src   string
	items []string
}

func newIgnoreList(src string) (*IgnoreList, error) {
	list := &IgnoreList{src: src}

	//TODO: Should return custom error?
	// Load global ignore list
	configHome := config.AppConfigHome()
	if err := readIgnoreFile(configHome, &list.items); err != nil {
		return nil, err
	}

	//load package ignore list
	if err := readIgnoreFile(src, &list.items); err != nil {
		return nil, err
	}
	return list, nil
}

func (i *IgnoreList) forPackage(pkg string) ([]string, error) {
	result := append([]string(nil), i.items...)
	if err := readIgnoreFile(filepath.Join(i.src, pkg), &result); err != nil {
		return nil, &EngineError{
			Message: "failed to read the ignore list for the package",
			Cause:   err,
		}
	}
	return result, nil
}

func (i *IgnoreList) shouldIgnore(fileName, pkg string) (bool, error) {
	var ignoreList []string
	if pkg == "" {
		ignoreList = i.items
	} else {
		var err error
		ignoreList, err = i.forPackage(pkg)
		if err != nil {
			return false, err
		}
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
