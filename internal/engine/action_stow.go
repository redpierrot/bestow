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

const timestampFormat = "20060102150405"

type StowCommand struct {
	Strategy ResolveStrategy
}

func (c *StowCommand) resolve(candidate operationCandidate, fs FileSystem, l *slog.Logger) (fileAction, error) {
	destExists, err := fs.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionLink(candidate.source, candidate.destination, l), nil
	}
	existing, err := fs.ExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return nil, err
	}
	if existing == file.ExistingDir {
		return nil, fmt.Errorf("stow %s: %w", candidate.destination, errDestIsDir)
	}
	if existing == file.ExistingManagedSymlink {
		return newFileActionUpToDate(candidate.source, candidate.destination, "file already stowed", l), nil
	}

	switch c.Strategy {
	case ResolveForce:
		l.Debug("existing destination will be replaced", "destination", candidate.destination, "strategy", c.Strategy)
		return newFileActionReplace(candidate.source, candidate.destination, l), nil
	case ResolveSkip:
		l.Debug("skipping the existing file at the destination", "destination", candidate.destination, "strategy", c.Strategy)
		return newFileActionSkip(candidate.source, candidate.destination, fmt.Sprintf("%s: %s", existing, "skip"), l), nil
	case ResolveBackup:
		l.Debug("existing file at the destination will be backed up and replaced", "destination", candidate.destination, "strategy", c.Strategy)
		backupID := time.Now().Format(timestampFormat)
		backupPath := fmt.Sprintf("%s.%s.%s", candidate.destination, backupID, backupExtension)
		return newFileActionBackup(candidate.source, candidate.destination, backupPath, l), nil
	case ResolveAdopt:
		if existing != file.ExistingRegularFile {
			return newFileActionSkip(candidate.source, candidate.destination, fmt.Sprintf("adopt %s: %s", candidate.destination, existing), l), nil
		}
		return newFileActionAdopt(candidate.source, candidate.destination, l), nil
	default:
		l.Warn("unsupported resolution strategy", "strategy", c.Strategy, "destination", candidate.destination)
		return nil, fmt.Errorf("unsupported strategy %v: %w", c.Strategy, errUnsupportedAction)
	}

}
