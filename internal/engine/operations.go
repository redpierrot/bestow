/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"

	"github.com/redpierrot/bestow/internal/file"
)

// ResolveStrategy defines the file action resolving strategy when the destination exist
type ResolveStrategy int

const (
	// ResolveSkip skips the operation when destination exists
	ResolveSkip ResolveStrategy = iota
	// ResolveForce forcefully replaces the destination file
	ResolveForce
	// ResolveAdopt copies the file from the destination to source before linking
	ResolveAdopt
	// ResolveBackup backs up the destination file before linking
	ResolveBackup
)

const maxBackupFileCount = 6

func (r ResolveStrategy) String() string {
	switch r {
	case ResolveSkip:
		return "skip"
	case ResolveForce:
		return "force"
	case ResolveAdopt:
		return "adopt"
	case ResolveBackup:
		return "backup"
	default:
		return fmt.Sprintf("unknown %d", r)
	}
}

type operationCandidate struct {
	source      string
	destination string
}

func (e *Engine) buildOperations(cfg *CommandConfig) ([]fileAction, error) {
	e.logger.Debug("populating operations", "action", cfg.Action)
	packageList, err := e.buildPackageList(cfg.Args)
	if err != nil {
		return nil, err
	}
	candidates := make([]operationCandidate, 0, len(packageList))
	errs := make([]error, 0, len(packageList))
	for _, pkg := range packageList {
		packageCandidates, err := e.fileOperations(pkg)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		candidates = append(candidates, packageCandidates...)
	}
	if len(errs) > 0 {
		return nil, &AggregatedError{
			Msg:   "failed to calculate the operations",
			Items: errs,
		}
	}
	if err := e.validateDestinations(candidates); err != nil {
		return nil, err
	}

	switch cfg.Action {
	case CommandStow:
		return e.buildStowOperations(candidates, cfg.ConflictStrategy)
	case CommandUnstow:
		// TODO: Remove empty parents should be configurable and unstow should handle it
		return e.buildUnstowOperations(candidates)
	}
	return nil, fmt.Errorf("action %s: %w", cfg.Action, ErrUnsupportedAction)
}

func (e *Engine) validateDestinations(candidates []operationCandidate) error {
	destinations := make(map[string][]string)
	for _, candidate := range candidates {
		if candidate.destination != "" {
			destinations[candidate.destination] = append(destinations[candidate.destination], candidate.source)
		}
	}
	conflicts := make([]DestinationConflict, 0, len(destinations))
	for destination, sources := range destinations {
		if len(sources) > 1 {
			conflicts = append(conflicts, DestinationConflict{
				Destination: destination,
				Sources:     sources,
			})
		}
	}
	if len(conflicts) > 0 {
		return &ConflictError{
			Op:        "validate",
			Conflicts: conflicts,
			Err:       ErrMultiFile,
		}
	}
	e.logger.Debug("all destinations are valid")
	return nil
}

func (e *Engine) fileOperations(pkg string) ([]operationCandidate, error) {
	pkgPath := filepath.Join(e.source, pkg)
	fileList, err := e.fileSystem.ListAllFiles(pkgPath)
	if err != nil {
		return nil, err
	}
	candidates := make([]operationCandidate, 0, len(fileList))
	for _, path := range fileList {
		relPath, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return nil, err
		}
		shouldIgnore, err := e.ignore.isIgnoredFile(relPath, pkg)
		if err != nil {
			return nil, err
		}
		if shouldIgnore {
			e.logger.Debug("ignoring the file due to ignore list", "file_name", path)
			continue
		}
		destinationFile := filepath.Join(e.destination, relPath)
		candidates = append(candidates, operationCandidate{
			source:      path,
			destination: destinationFile,
		})
		e.logger.Debug("adding candidate file", "file_name", path)
	}
	return candidates, nil
}

func (e *Engine) buildStowOperations(candidates []operationCandidate, strategy ResolveStrategy) ([]fileAction, error) {
	actions := make([]fileAction, 0, len(candidates))
	errs := make([]error, 0, len(candidates))
	for _, candidate := range candidates {
		action, err := e.stowFileAction(candidate, strategy)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		actions = append(actions, action)
	}
	if len(errs) > 0 {
		return nil, &AggregatedError{
			Msg:   "failed to resolve operations",
			Items: errs,
		}
	}
	return actions, nil
}

func (e *Engine) buildUnstowOperations(candidates []operationCandidate) ([]fileAction, error) {
	actions := make([]fileAction, 0, len(candidates))
	errs := make([]error, 0, len(candidates))
	for _, candidate := range candidates {
		action, err := e.unstowFileAction(candidate)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		actions = append(actions, action)
	}
	if len(errs) > 0 {
		return nil, &AggregatedError{
			Msg:   "failed to resolve operations",
			Items: errs,
		}
	}
	return actions, nil
}

func (e *Engine) stowFileAction(candidate operationCandidate, strategy ResolveStrategy) (fileAction, error) {
	destExists, err := e.fileSystem.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionLink(candidate.source, candidate.destination, e.logger), nil
	}
	existing, err := e.fileSystem.ExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return nil, err
	}
	if existing == file.ExistingDir {
		return nil, &HintedError{
			Op:   fmt.Sprintf("stow %s %s", candidate.source, candidate.destination),
			Hint: fmt.Sprintf("remove the directory %s", candidate.destination),
			Err:  ErrDestIsDir,
		}
	}
	if existing == file.ExistingManagedSymlink {
		return newFileActionUpToDate(candidate.source, candidate.destination, "file already stowed", e.logger), nil
	}
	switch strategy {
	case ResolveForce:
		e.logger.Debug("existing destination will be replaced", "destination", candidate.destination, "strategy", strategy)
		return newFileActionReplace(candidate.source, candidate.destination, e.logger), nil
	case ResolveSkip:
		e.logger.Debug("skipping the existing file at the destination", "destination", candidate.destination, "strategy", strategy)
		return newFileActionSkip(candidate.source, candidate.destination, fmt.Sprintf("%s: %s", existing.String(), strategy), e.logger), nil
	case ResolveBackup:
		e.logger.Debug("existing file at the destination will be backed up and replaced", "destination", candidate.destination, "strategy", strategy)
		backupPath, err := e.calculateBackupPath(candidate.destination)
		if err != nil {
			return nil, err
		}
		return newFileActionBackup(candidate.source, candidate.destination, backupPath, e.logger), nil
	case ResolveAdopt:
		switch existing {
		case file.ExistingForeignSymlink:
			e.logger.Warn("cannot adopt the existing symlink at destination", "destination", candidate.destination)
			return newFileActionSkip(candidate.source, candidate.destination, "adopt foreign symlink", e.logger), nil
		case file.ExistingRegularFile:
			e.logger.Debug("existing destination will be adopted to source", "destination", candidate.destination, "strategy", strategy)
			return newFileActionAdopt(candidate.source, candidate.destination, e.logger), nil
		default:
			return nil, fmt.Errorf("unsupported existing file type %s", existing)
		}
	default:
		e.logger.Warn("unsupported resolution strategy", "strategy", strategy, "destination", candidate.destination)
		return nil, fmt.Errorf("unsupported strategy %s: %w", strategy, ErrUnsupportedAction)
	}
}

func (e *Engine) unstowFileAction(candidate operationCandidate) (fileAction, error) {
	destExists, err := e.fileSystem.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionUpToDate(candidate.source, candidate.destination, "destination does not exist", e.logger), nil
	}
	existing, err := e.fileSystem.ExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return nil, err
	}
	switch existing {
	case file.ExistingDir:
		return nil, &HintedError{
			Op:   fmt.Sprintf("unstow %s %s", candidate.source, candidate.destination),
			Hint: fmt.Sprintf("remove the directory %s", candidate.destination),
			Err:  ErrDestIsDir,
		}
	case file.ExistingRegularFile:
		return newFileActionSkip(candidate.source, candidate.destination, "regular file", e.logger), nil
	case file.ExistingManagedSymlink:
		return newFileActionRemove(candidate.source, candidate.destination, e.logger), nil
	case file.ExistingForeignSymlink:
		return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", e.logger), nil
	}
	e.logger.Warn("destination is not managed by bestow", "destination", candidate.destination, "file_type", existing)
	return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink", e.logger), nil
}

func (e *Engine) calculateBackupPath(dest string) (string, error) {
	i := 0
	for i < maxBackupFileCount {
		backupPath := fmt.Sprintf("%s.%d.%s", dest, i, backupExtension)
		exists, err := e.fileSystem.Exists(backupPath)
		if err != nil {
			return "", fmt.Errorf("backup %s: %w", backupPath, err)
		}
		if !exists {
			return backupPath, nil
		}
		i++
	}
	return "", &HintedError{
		Op:   fmt.Sprintf("backup %s", dest),
		Err:  fmt.Errorf("exceeds maximum backup count %d", maxBackupFileCount),
		Hint: fmt.Sprintf("remove existing backups %s.*.bestow.backup", dest),
	}
}
