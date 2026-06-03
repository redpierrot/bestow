/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/file"
)

type ResolveStrategy int

const (
	ResolveSkip ResolveStrategy = iota
	ResolveForce
	ResolveAdopt
	ResolveBackup
)

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

type OperationCandidate struct {
	source      string
	destination string
}

func (e *Engine) populateOperations(ctx *CommandContext) ([]FileAction, error) {
	e.logger.Debug("populating operations", "action", ctx.Action)
	packageList, err := e.populatePackageList(ctx.Args)
	if err != nil {
		return nil, err
	}
	candidates := make([]OperationCandidate, 0, len(packageList))
	for _, pkg := range packageList {
		packageCandidates, err := e.getFileOperations(pkg)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, packageCandidates...)
	}
	if err := validateDestinations(candidates); err != nil {
		return nil, err
	}

	switch ctx.Action {
	case ActionStow:
		return e.resolveStowOpts(candidates, ctx.ConflictStrategy)
	case ActionUnstow:
		// TODO: Remove empty parents should be config and unstow should handle it
		return e.resolveUnstowOpts(candidates)
	}
	return nil, fmt.Errorf("action %s: %w", ctx.Action, ErrUnsupportedAction)
}

func validateDestinations(candidates []OperationCandidate) error {
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
	return nil
}

func (e *Engine) getFileOperations(pkg string) ([]OperationCandidate, error) {
	pkgPath := filepath.Join(e.source, pkg)
	fileList, err := e.fileSystem.ListAllFiles(pkgPath)
	if err != nil {
		return nil, err
	}
	candidates := make([]OperationCandidate, 0, len(fileList))
	for _, filePath := range fileList {
		relPath, err := filepath.Rel(pkgPath, filePath)
		if err != nil {
			return nil, err
		}
		doIgnore, err := e.ignore.shouldIgnorePkgFile(relPath, pkg)
		if err != nil {
			return nil, err
		}
		if doIgnore {
			e.logger.Debug("ignoring the file due to ignore list", "file_name", filePath)
			continue
		}
		destinationFile := filepath.Join(e.destination, relPath)
		candidates = append(candidates, OperationCandidate{
			source:      filePath,
			destination: destinationFile,
		})
		e.logger.Debug("adding candidate file", "file_name", filePath)
	}
	return candidates, nil
}

func (e *Engine) resolveStowOpts(candidates []OperationCandidate, strategy ResolveStrategy) ([]FileAction, error) {
	actions := make([]FileAction, 0, len(candidates))
	for _, candidate := range candidates {
		action, err := e.getStowFileAction(candidate, strategy)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func (e *Engine) resolveUnstowOpts(candidates []OperationCandidate) ([]FileAction, error) {
	actions := make([]FileAction, 0, len(candidates))
	for _, candidate := range candidates {
		action, err := e.getUnstowFileAction(candidate)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func (e *Engine) getStowFileAction(candidate OperationCandidate, strategy ResolveStrategy) (FileAction, error) {
	destExists, err := e.fileSystem.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionLink(candidate.source, candidate.destination), nil
	}
	existing, err := e.fileSystem.GetExistingFileType(candidate.source, candidate.destination)
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
	// TODO: Managed symlink finding strategy has a flaw.
	// Managed symlink: Existing link lives inside the source
	// This can be either the same file or not. Should update it anyway
	if existing == file.ExistingManagedSymlink {
		return newFileActionUpToDate(candidate.source, candidate.destination, "file already stowed"), nil
	}
	switch strategy {
	case ResolveForce:
		e.logger.Debug("existing destination will be replaced", "destination", candidate.destination, "strategy", strategy)
		return newFileActionReplace(candidate.source, candidate.destination), nil
	case ResolveSkip:
		e.logger.Debug("skipping the existing file at the destination", "destination", candidate.destination, "strategy", strategy)
		return newFileActionSkip(candidate.source, candidate.destination, "strategy skip"), nil
	case ResolveBackup:
		e.logger.Debug("existing file at the destination will be backed up and replaced", "destination", candidate.destination, "strategy", strategy)
		backupFilePath := candidate.destination + backupExtension
		return newFileActionBackup(candidate.source, candidate.destination, backupFilePath), nil
	case ResolveAdopt:
		switch existing {
		case file.ExistingForeignSymlink:
			e.logger.Warn("cannot adopt the existing symlink at destination", "destination", candidate.destination)
			return newFileActionSkip(candidate.source, candidate.destination, "adopt foreign symlink"), nil
		case file.ExistingRegularFile:
			e.logger.Debug("existing destination will be adopted to source", "destination", candidate.destination, "strategy", strategy)
			return newFileActionAdopt(candidate.source, candidate.destination), nil
		default:
			return nil, fmt.Errorf("unsupported existing file type %s", existing)
		}
	default:
		e.logger.Warn("unsupported resolution strategy", "strategy", strategy, "destination", candidate.destination)
		return nil, fmt.Errorf("unsupported strategy %s: %w", strategy, ErrUnsupportedAction)
	}
}

func (e *Engine) getUnstowFileAction(candidate OperationCandidate) (FileAction, error) {
	exists, err := e.fileSystem.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !exists {
		return newFileActionUpToDate(candidate.source, candidate.destination, "destination does not exist"), nil
	}
	existing, err := e.fileSystem.GetExistingFileType(candidate.source, candidate.destination)
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
		return newFileActionSkip(candidate.source, candidate.destination, "regular file"), nil
	case file.ExistingManagedSymlink:
		return newFileActionRemove(candidate.source, candidate.destination), nil
	case file.ExistingForeignSymlink:
		return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink"), nil
	}
	e.logger.Warn("destination is not managed by bestow", "destination", candidate.destination, "file_type", existing)
	return newFileActionSkip(candidate.source, candidate.destination, "unmanaged symlink"), nil
}
