/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"

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
		return newFileActionRemove(candidate.source, candidate.destination, c.KeepEmptyParents, l), nil
	case file.ExistingForeignSymlink:
		return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", l), nil
	}
	l.Warn("destination is not managed by bestow", "destination", candidate.destination, "file_type", existing)
	return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", l), nil
}
