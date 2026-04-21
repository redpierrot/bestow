package engine

import (
	"github.com/ThisaruGuruge/bestow/internal/file"
	"github.com/ThisaruGuruge/bestow/internal/log"
)

type FileAction string

const (
	FileActionLink       FileAction = "Link"
	FileActionSkip       FileAction = "Skip"
	FileActionRemoveLink FileAction = "Remove and Link"
	FileActionBackupLink FileAction = "Backup and Link"
	FileActionAdoptLink  FileAction = "Adopt and Link"
)

func (e *Engine) resolveFileAction(operation *Operation, strategy ResolveStrategy, existing file.ExistingType) error {
	log.Debug("Resolving file actions", "source", operation.Source, "destination", operation.Destination, "strategy", strategy, "existing_type", existing)
	switch existing {
	case file.ExistingManagedSymlink:
		log.Debug("symlink already exists, skipping", "destination", operation.Destination, "strategy", strategy, "existing_type", existing)
		operation.Action = FileActionSkip
	case file.ExistingDir:
		return resolveExistingDir(operation, strategy)
	case file.ExistingRegularFile:
		return resolveRegularFile(operation, strategy)
	case file.ExistingForeignSymlink:
		return resolveForeignSymlink(operation, strategy)
	}
	return nil
}

func resolveExistingDir(operation *Operation, strategy ResolveStrategy) error {
	if strategy == ResolveSkip {
		log.Warn("destination is a directory; skipping the file", "destination", operation.Destination, "destination_type", "DIRECTORY", "strategy", strategy)
		operation.Action = FileActionSkip
		return nil
	}
	return &EngineError{
		Message: "cannot perform operation; destination is a directory",
	}
}

func resolveRegularFile(operation *Operation, strategy ResolveStrategy) error {
	switch strategy {
	case ResolveSkip:
		log.Warn("destination exists; skipping the file", "destination", operation.Destination, "destination_type", "FILE", "strategy", strategy)
		operation.Action = FileActionSkip
	case ResolveForce:
		operation.Action = FileActionRemoveLink
	case ResolveAdopt:
		operation.Action = FileActionAdoptLink
	case ResolveBackup:
		operation.Action = FileActionBackupLink
	}
	return nil
}

func resolveForeignSymlink(operation *Operation, strategy ResolveStrategy) error {
	switch strategy {
	case ResolveSkip:
		operation.Action = FileActionSkip
		log.Warn("destination exists, skipping the file", "destination", operation.Destination, "destination_type", "FOREIGN SYMLINK", "strategy", strategy)
	case ResolveForce:
		operation.Action = FileActionRemoveLink
	case ResolveAdopt:
		operation.Action = FileActionAdoptLink
	case ResolveBackup:
		operation.Action = FileActionBackupLink
	}
	return nil
}

func resolveManagedSymlink(operation *Operation, strategy ResolveStrategy) error {

	return nil
}
