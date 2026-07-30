/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"

	"github.com/redpierrot/bestow/internal/file"
)

type UnstowOperation struct {
	fs FileSystem
	l  *slog.Logger
}

func newUnstowOperation(fs FileSystem, l *slog.Logger) *UnstowOperation {
	return &UnstowOperation{
		fs: fs,
		l:  l,
	}
}

func (uo *UnstowOperation) FileAction(candidate operationCandidate) (fileAction, error) {
	destExists, err := uo.fs.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionUpToDate(candidate.source, candidate.destination, "destination does not exist", uo.l), nil
	}
	existing, err := uo.fs.ExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return nil, err
	}
	switch existing {
	case file.ExistingDir:
		return nil, fmt.Errorf("unstow %s: %w", candidate.destination, errDestIsDir)
	case file.ExistingRegularFile:
		return newFileActionSkip(candidate.source, candidate.destination, "regular file", uo.l), nil
	case file.ExistingManagedSymlink:
		return newFileActionRemove(candidate.source, candidate.destination, uo.l), nil
	case file.ExistingForeignSymlink:
		return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", uo.l), nil
	}
	uo.l.Warn("destination is not managed by bestow", "destination", candidate.destination, "file_type", existing)
	return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", uo.l), nil
}
