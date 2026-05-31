/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/file"
)

type ResolveStrategy string

const (
	ResolveSkip        ResolveStrategy = "skip"
	ResolveForce       ResolveStrategy = "force"
	ResolveAdopt       ResolveStrategy = "adopt"
	ResolveBackup      ResolveStrategy = "backup"
	ResolveInteractive ResolveStrategy = "interactive"
)

type Operation struct {
	Source      string
	Destination string
	BackupPath  string
	Action      FileAction
}

type OperationCandidate struct {
	source      string
	destination string
}

// TODO: Need to verify if two operations have the same destination.
// Which should be an error; We should catch it here before processing to
// execute the operations
func (e *Engine) populateOperations(ctx *CommandContext) ([]FileAction, error) {
	e.Logger.Debug("populating operations", "action", ctx.Action)
	packageList, err := e.populatePackageList(ctx.Args)
	if err != nil {
		return nil, err
	}
	candidates := []OperationCandidate{}
	for _, pkg := range packageList {
		packageCandidates, err := e.getFileOperations(pkg)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, packageCandidates...)
	}
	switch ctx.Action {
	case ActionStow:
		return e.resolveStowOpts(candidates, ctx.ConflictStrategy)
	case ActionUnstow:
		return e.resolveUnstowOpts(candidates)
	}
	return nil, fmt.Errorf("action %s: %w", ctx.Action, ErrUnsupportedAction)
}

func (e *Engine) getFileOperations(pkg string) ([]OperationCandidate, error) {
	pkgPath := filepath.Join(e.Source, pkg)
	fileList, err := e.FileSystem.ListAllFiles(pkgPath)
	if err != nil {
		return nil, err
	}
	candidates := make([]OperationCandidate, 0, len(fileList))
	for _, filePath := range fileList {
		relPath, err := filepath.Rel(pkgPath, filePath)
		if err != nil {
			return nil, err
		}
		doIgnore, err := e.Ignore.shouldIgnore(relPath, pkg)
		if err != nil {
			return nil, err
		}
		if doIgnore {
			e.Logger.Debug("ignoring the file due to ignore list", "file_name", filePath)
			continue
		}
		e.Logger.Debug("retrieving file path", "file_name", filePath, "source_file", filePath)
		relativePath, err := filepath.Rel(pkgPath, filePath)
		if err != nil {
			return nil, err
		}
		e.Logger.Debug("relative path of the file", "file_path", relativePath)
		destinationFile := filepath.Join(e.Destination, relativePath)
		e.Logger.Debug("retrieving file path", "file_name", filePath, "destination_file", destinationFile)
		candidates = append(candidates, OperationCandidate{
			source:      filePath,
			destination: destinationFile,
		})
		e.Logger.Debug("adding candidate file", "file_name", filePath)
	}
	return candidates, nil
}

func (e *Engine) resolveStowOpts(candidates []OperationCandidate, strategy ResolveStrategy) ([]FileAction, error) {
	destinations := make(map[string][]string)
	actions := make([]FileAction, 0, len(candidates))
	for _, candidate := range candidates {
		action, err := e.getStowFileAction(candidate, strategy)
		if err != nil {
			return nil, err
		}
		if affectsDestination(action) {
			destinations[action.Destination()] = append(destinations[action.Destination()], action.Source())
		}
		actions = append(actions, action)
	}
	conflicts := []DestinationConflict{}
	for destination, sources := range destinations {
		if len(sources) > 1 {
			conflicts = append(conflicts, DestinationConflict{
				Destination: destination,
				Sources:     sources,
			})
		}
	}
	if len(conflicts) > 0 {
		return nil, &ConflictError{
			Op:        "stow",
			Conflicts: conflicts,
			Err:       ErrMultiFile,
		}
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

// TODO: When skipping files;
// - in .bestowignore: debug log: Done
// - skip because already stowed (due to state of the operation): include a summary: FileActionNoOp
// - skip because conflict resolution strategy is set to skip: print as same as any other operation: FileActionSkip
func (e *Engine) getStowFileAction(candidate OperationCandidate, strategy ResolveStrategy) (FileAction, error) {
	destExists, err := e.FileSystem.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !destExists {
		return newFileActionLink(candidate.source, candidate.destination), nil
	}
	existing, err := e.FileSystem.GetExistingFileType(candidate.source, candidate.destination)
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
		return newFileActionNoOp(candidate.source, candidate.destination, "file already stowed"), nil
	}
	switch strategy {
	case ResolveForce:
		e.Logger.Debug("existing destination will be replaced", "destination", candidate.destination, "strategy", strategy)
		return newFileActionReplace(candidate.source, candidate.destination), nil
	case ResolveSkip:
		e.Logger.Warn("skipping the existing file at the destination", "destination", candidate.destination, "strategy", strategy)
		return newFileActionSkip(candidate.source, candidate.destination), nil
	case ResolveBackup:
		e.Logger.Debug("existing file at the destination will be backed up and replaced", "destination", candidate.destination, "strategy", strategy)
		backupFilePath := candidate.destination + backupExtension
		return newFileActionBackup(candidate.source, candidate.destination, backupFilePath), nil
	case ResolveAdopt:
		switch existing {
		case file.ExistingForeignSymlink:
			e.Logger.Warn("cannot adopt the existing symlink at destination", "destination", candidate.destination)
			return newFileActionNoOp(candidate.source, candidate.destination, "foreign symlinks cannot be adopted"), nil
		case file.ExistingRegularFile:
			e.Logger.Debug("existing destination will be adopted to source", "destination", candidate.destination, "strategy", strategy)
			return newFileActionAdopt(candidate.source, candidate.destination), nil
		default:
			return nil, fmt.Errorf("unsupported existing file type %s", existing)
		}
	default:
		e.Logger.Warn("unsupported resolution strategy", "strategy", strategy, "destination", candidate.destination)
		return nil, fmt.Errorf("unsupported strategy %s: %w", strategy, ErrUnsupportedAction)
	}
}

func (e *Engine) getUnstowFileAction(candidate OperationCandidate) (FileAction, error) {
	exists, err := e.FileSystem.Exists(candidate.destination)
	if err != nil {
		return nil, err
	}
	if !exists {
		return newFileActionSkip(candidate.source, candidate.destination), nil
	}
	existing, err := e.FileSystem.GetExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return nil, err
	}
	if existing == file.ExistingDir {
		return nil, &HintedError{
			Op:   fmt.Sprintf("unstow %s %s", candidate.source, candidate.destination),
			Hint: fmt.Sprintf("remove the directory %s", candidate.destination),
			Err:  ErrDestIsDir,
		}
	}
	if existing == file.ExistingManagedSymlink {
		return newFileActionRemove(candidate.source, candidate.destination), nil
	}
	e.Logger.Warn("destination is not managed by bestow", "destination", candidate.destination, "file_type", existing)
	return newFileActionSkip(candidate.source, candidate.destination), nil
}
