/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/redpierrot/bestow/internal/file"
)

type UnstowCommand struct {
	KeepEmptyParents bool
}

func (c *UnstowCommand) resolve(candidate operationCandidate, fs FileSystem, l *slog.Logger) (fileAction, error) {
	destExists, err := fs.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionUpToDate(candidate.source, candidate.destination, "destination does not exist", l), nil
	}
	existing, err := fs.ExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return nil, err
	}
	switch existing {
	case file.ExistingDir:
		return nil, fmt.Errorf("unstow %s: %w", candidate.destination, errDestIsDir)
	case file.ExistingRegularFile:
		return newFileActionSkip(candidate.source, candidate.destination, "regular file", l), nil
	case file.ExistingManagedSymlink:
		return newFileActionRemove(candidate.source, candidate.destination, l), nil
	case file.ExistingForeignSymlink:
		return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", l), nil
	}
	l.Warn("destination is not managed by bestow", "destination", candidate.destination, "file_type", existing)
	return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", l), nil
}

func (c *UnstowCommand) cleanup(candidates []operationCandidate, fs FileSystem, l *slog.Logger, root string) []fileAction {
	if c.KeepEmptyParents {
		return nil
	}
	dirs := getDirs(candidates, root)
	actions := make([]fileAction, 0)
	for _, dir := range dirs {
		actions = append(actions, newFileActionRemoveDir("", dir, l))
		l.Debug("remove parent dirs", "directory", dir)
	}
	return actions
}

func getDirs(candidates []operationCandidate, root string) []string {
	dirMap := make(map[string]bool)
	root = filepath.Clean(root)
	for _, candidate := range candidates {
		dir := filepath.Clean(filepath.Dir(candidate.destination))
		for dir != root {
			rel, err := filepath.Rel(root, dir)
			if err != nil {
				fmt.Println(err)
				break
			}
			if strings.HasPrefix(rel, "..") || rel == "." {
				break
			}
			dirMap[dir] = true
			dir = filepath.Dir(dir)
		}
	}
	dirs := slices.Collect(maps.Keys(dirMap))
	slices.Sort(dirs)
	slices.Reverse(dirs) // Reverse the array so the child directories removed first
	return dirs
}
