/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"
)

func (e *Engine) buildPackageList(args []string) ([]string, error) {
	e.logger.Debug("populating package list", "source", e.source)
	var pkgCandidates []string
	var err error
	if len(args) == 0 {
		e.logger.Debug("no packages provided; processing all packages")
		pkgCandidates, err = e.retrieveAllPackages()
		if err != nil {
			return nil, err
		}
	} else {
		pkgCandidates, err = e.retrievePackagesFromArgs(args)
		if err != nil {
			return nil, err
		}
	}
	packages := e.filterPackages(pkgCandidates)
	e.logger.Debug("package list populated", "package_list", packages)
	return packages, nil
}

func (e *Engine) retrieveAllPackages() ([]string, error) {
	dirs, err := e.fileSystem.ListDirs(e.source)
	if err != nil {
		return nil, err
	}
	candidates := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		candidate, err := filepath.Rel(e.source, dir)
		if err != nil {
			return nil, fmt.Errorf("rel %s %s: %w", e.source, dir, err)
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (e *Engine) retrievePackagesFromArgs(candidates []string) ([]string, error) {
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "." {
			return nil, &HintedError{
				Op:   fmt.Sprintf("read package %s", candidate),
				Hint: "move root files to suitable directory (`zsh/`, `bash/`, etc.)",
				Err:  errRootIsNotPkg,
			}
		}
		pkgPath := filepath.Clean(candidate)
		isDir, err := e.fileSystem.IsDir(filepath.Join(e.source, pkgPath))
		if err != nil {
			return nil, fmt.Errorf("read package %s: %w", candidate, err)
		}
		if !isDir {
			return nil, &HintedError{
				Op:   fmt.Sprintf("read package %s", candidate),
				Hint: fmt.Sprintf("make sure the %s is a directory", candidate),
				Err:  errPkgIsNotDir,
			}
		}
		result = append(result, pkgPath)
	}
	return result, nil
}

func (e *Engine) filterPackages(candidates []string) []string {
	e.logger.Debug("filtering packages", "candidates", candidates, "filter", e.ignore.items)
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		shouldIgnore := e.ignore.isIgnored(candidate)
		if shouldIgnore {
			e.logger.Debug("ignoring package candidate", "candidate", candidate)
			continue
		}
		e.logger.Debug("adding package to process", "package", candidate)
		result = append(result, candidate)
	}
	return result
}
