/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"log/slog"
	"path/filepath"
)

type Operation interface {
	FileAction(candidate operationCandidate) (fileAction, error)
}

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

type operationCandidate struct {
	source      string
	destination string
}

func getOperation(kind CommandKind, fs FileSystem, l *slog.Logger, strategy ResolveStrategy) (Operation, error) {
	switch kind {
	case CommandStow:
		return newStowOperation(fs, l, strategy), nil
	case CommandUnstow:
		return newUnstowOperation(fs, l), nil
	default:
		return nil, errUnsupportedAction
	}
}

func (e *Engine) findCandidates(args []string) ([]operationCandidate, error) {
	packageList, err := e.buildPackageList(args)
	if err != nil {
		return nil, err
	}
	candidates := make([]operationCandidate, 0, len(packageList))
	errs := make([]error, 0, len(packageList))
	for _, pkg := range packageList {
		packageCandidates, err := e.findPackageCandidates(pkg)
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
	return candidates, nil
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
			Err:       errMultiFile,
		}
	}
	e.logger.Debug("all destinations are valid")
	return nil
}

func (e *Engine) findPackageCandidates(pkg string) ([]operationCandidate, error) {
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

func (e *Engine) buildFileActions(candidates []operationCandidate, strategy ResolveStrategy, cmdKind CommandKind) ([]fileAction, error) {
	actions := make([]fileAction, 0, len(candidates))
	errs := make([]error, 0, len(candidates))
	operation, err := getOperation(cmdKind, e.fileSystem, e.logger, strategy)
	if err != nil {
		return nil, err
	}
	for _, candidate := range candidates {
		action, err := operation.FileAction(candidate)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		actions = append(actions, action)
	}
	if len(errs) > 0 {
		return nil, &AggregatedError{
			Msg:   "resolve operations",
			Items: errs,
		}
	}
	return actions, nil
}
