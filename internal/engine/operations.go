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
// Which should be an error; We should catch it here before proceesing to
// execute the operations
func (e *Engine) populateOperations(ctx *CommandContext) ([]Operation, error) {
	e.Logger.Debug("populating operations", "action", ctx.Action)
	result := []Operation{}
	packageList, err := e.populatePackageList(ctx.Args)
	if err != nil {
		return nil, err
	}
	for _, pkg := range packageList {
		pacakgeOperations, err := e.getPackageOperation(pkg, ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, pacakgeOperations...)
	}
	return result, nil
}

func (e *Engine) getPackageOperation(pkg string, ctx *CommandContext) ([]Operation, error) {
	candidates, err := e.getFileOperations(pkg)
	if err != nil {
		return nil, err
	}
	operations := []Operation{}

	switch ctx.Action {
	case ActionStow:
		return e.resolveStowOpts(candidates, ctx.ConflictStrategy)
	case ActionUnstow:
		return e.resolveUnstowOpts(candidates, ctx.ConflictStrategy)
	}
	return operations, nil
}

func (e *Engine) getFileOperations(pkg string) ([]OperationCandidate, error) {
	fileList, err := e.FileSystem.ListAllFilesInDir(e.Source, pkg)
	if err != nil {
		return nil, &EngineError{
			Message: "failed to read the files in the package",
			Cause:   err,
		}
	}
	candidates := []OperationCandidate{}
	for _, fileName := range fileList {
		doIgnore, err := e.Ignore.shouldIgnore(fileName, pkg)
		if err != nil {
			return nil, err
		}
		if doIgnore {
			e.Logger.Debug("ignoring the file due to ignore list", "file_name", fileName)
			continue
		}
		sourceFile := filepath.Join(e.Source, fileName)
		e.Logger.Debug("file name", "file_name", fileName)
		relativePath := filepath.Join(e.FileSystem.GetPathSegments(fileName)[1:]...)
		e.Logger.Debug("relative path of the file", "file_path", relativePath)
		destinationFile := filepath.Join(e.Destination, relativePath)
		// TODO: Remove relative path; calculate destination by removing the package name; remove unncessary calculation of source and dest files
		candidates = append(candidates, OperationCandidate{
			source:      sourceFile,
			destination: destinationFile,
		})
		e.Logger.Debug("adding candidate file", "file_name", fileName)
	}
	return candidates, nil
}

func (e *Engine) resolveStowOpts(candidates []OperationCandidate, strategy ResolveStrategy) ([]Operation, error) {
	destinations := make(map[string]bool)
	operations := []Operation{}
	for _, candidate := range candidates {
		operation, err := e.resolveStowOpt(candidate, strategy)
		if err != nil {
			return nil, err
		}
		if destinations[operation.Destination] {
			return nil, &EngineError{
				Message: fmt.Sprintf("multiple sources competing for same detination: %s", operation.Destination),
			}
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (e *Engine) resolveUnstowOpts(candidates []OperationCandidate, strategy ResolveStrategy) ([]Operation, error) {
	operations := []Operation{}
	for _, candidate := range candidates {
		operation, err := e.resolveUnstowOpt(candidate, strategy)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (e *Engine) resolveStowOpt(candidate OperationCandidate, strategy ResolveStrategy) (Operation, error) {
	e.Logger.Debug("resolving stow operation", "source", candidate.source, "destination", candidate.destination)
	destExists, err := e.FileSystem.Exists(candidate.destination)
	if err != nil {
		return Operation{}, &EngineError{
			Message: "failed to check the destination file",
			Cause:   err,
			Hint:    "check permissions on destination",
		}
	}
	operation := &Operation{
		Source:      candidate.source,
		Destination: candidate.destination,
	}
	if !destExists {
		operation.Action = FileActionLink
		return *operation, nil
	}
	action, err := e.getStowFileAction(candidate, strategy)
	if err != nil {
		return *operation, err
	}
	operation.Action = action
	return *operation, nil
}

func (e *Engine) resolveUnstowOpt(candidate OperationCandidate, strategy ResolveStrategy) (Operation, error) {
	destExists, err := e.FileSystem.Exists(candidate.destination)
	if err != nil {
		return Operation{}, &EngineError{
			Message: "failed to check the destination file",
			Cause:   err,
			Hint:    "check permissions on destination",
		}
	}
	operation := &Operation{
		Source:      candidate.source,
		Destination: candidate.destination,
	}
	if !destExists {
		e.Logger.Debug("file is not stowed", "file_name", candidate.source)
		operation.Action = FileActionNoOp
		return *operation, nil
	}
	action, err := e.getUnstowFileAction(candidate, strategy)
	if err != nil {
		return *operation, err
	}
	operation.Action = action
	return *operation, nil
}

// TODO: Should optimize the source file and destinationFile calculation

// TODO: When skipping files;
// - in .bestowignore: debug log: Done
// - skip because already stowed (due to state of the operation): include a summary: FileActionNoOp
// - skip because conflict resolution strategy is set to skip: print as same as any other operation: FileActionSkip
func (e *Engine) getStowFileAction(candidate OperationCandidate, strategy ResolveStrategy) (FileAction, error) {
	existing, err := e.FileSystem.GetExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return FileActionSkip, &EngineError{
			Message: "failed to read the existing file",
			Cause:   err,
		}
	}
	if existing == file.ExistingDir {
		return FileActionSkip, &EngineError{
			Message: "destination is a directory",
			Hint:    fmt.Sprintf("check the directory %s", candidate.destination),
		}
	}
	if existing == file.ExistingManagedSymlink {
		return FileActionNoOp, nil
	}
	if existing == file.ExistingForeignSymlink {
		if strategy == ResolveForce {
			e.Logger.Debug("existing foreign symlink will be replaced", "file", candidate.destination, "strategy", strategy)
			return FileActionReplaceLink, nil
		}
		if strategy == ResolveSkip {
			e.Logger.Warn("skipping the existing symlink", "file", candidate.destination)
			return FileActionSkip, nil
		}
		if strategy == ResolveAdopt {
			e.Logger.Warn("cannot adopt the existing symlink", "file", candidate.destination)
			return FileActionSkip, nil
		}
		if strategy == ResolveBackup {
			e.Logger.Debug("existing foreign symlink will be backed up and replaced", "file", candidate.destination, "strategy", strategy)
			return FileActionBackupLink, nil
		}
		// TODO: Interactive resolver?
	}
	if existing == file.ExistingRegularFile {
		if strategy == ResolveForce {
			e.Logger.Debug("existing file will be replaced", "file", candidate.destination)
			return FileActionReplaceLink, nil
		}
		if strategy == ResolveSkip {
			return FileActionSkip, nil
		}
		if strategy == ResolveAdopt {
			return FileActionAdoptLink, nil
		}
		if strategy == ResolveBackup {
			return FileActionBackupLink, nil
		}
	}

	return FileActionLink, nil
}

func (e *Engine) getUnstowFileAction(candidate OperationCandidate, strategy ResolveStrategy) (FileAction, error) {
	existing, err := e.FileSystem.GetExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return FileActionSkip, &EngineError{
			Message: "failed to read the existing file",
			Cause:   err,
		}
	}
	if existing == file.ExistingDir {
		return FileActionSkip, &EngineError{
			Message: "destination is a directory",
			Hint:    fmt.Sprintf("check the directory %s", candidate.destination),
		}
	}
	if existing == file.ExistingManagedSymlink {
		return FileActionUnlink, nil
	}

	return FileActionNoOp, nil
}
