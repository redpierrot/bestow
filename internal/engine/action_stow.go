/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/redpierrot/bestow/internal/file"
)

type StowOperation struct {
	fs       FileSystem
	l        *slog.Logger
	strategy ResolveStrategy
}

func newStowOperation(fs FileSystem, l *slog.Logger, strategy ResolveStrategy) *StowOperation {
	return &StowOperation{
		fs:       fs,
		l:        l,
		strategy: strategy,
	}
}

func (so *StowOperation) FileAction(candidate operationCandidate) (fileAction, error) {
	destExists, err := so.fs.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionLink(candidate.source, candidate.destination, so.l), nil
	}
	existing, err := so.fs.ExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return nil, err
	}
	if existing == file.ExistingDir {
		return nil, fmt.Errorf("stow %s: %w", candidate.destination, errDestIsDir)
	}
	if existing == file.ExistingManagedSymlink {
		return newFileActionUpToDate(candidate.source, candidate.destination, "file already stowed", so.l), nil
	}

	if so.strategy == ResolveAdopt {
		if existing != file.ExistingRegularFile {
			return newFileActionSkip(candidate.source, candidate.destination, fmt.Sprintf("adopt %s: %s", candidate.destination, existing), so.l), nil
		}
		return newFileActionAdopt(candidate.source, candidate.destination, so.l), nil
	}
	switch so.strategy {
	case ResolveForce:
		so.l.Debug("existing destination will be replaced", "destination", candidate.destination, "strategy", so.strategy)
		return newFileActionReplace(candidate.source, candidate.destination, so.l), nil
	case ResolveSkip:
		so.l.Debug("skipping the existing file at the destination", "destination", candidate.destination, "strategy", so.strategy)
		return newFileActionSkip(candidate.source, candidate.destination, fmt.Sprintf("%s: %s", existing, "skip"), so.l), nil
	case ResolveBackup:
		so.l.Debug("existing file at the destination will be backed up and replaced", "destination", candidate.destination, "strategy", so.strategy)
		backupID := time.Now().Format("yyyymmddhhmmss")
		backupPath := fmt.Sprintf("%s.%s.%s", candidate.destination, backupID, backupExtension)
		return newFileActionBackup(candidate.source, candidate.destination, backupPath, so.l), nil
	default:
		so.l.Warn("unsupported resolution strategy", "strategy", so.strategy, "destination", candidate.destination)
		return nil, fmt.Errorf("unsupported strategy %v: %w", so.strategy, errUnsupportedAction)
	}
}
