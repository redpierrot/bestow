/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/file"
)

type Action int

const (
	ActionStow Action = iota
	ActionUnstow
)

func (a Action) String() string {
	switch a {
	case ActionStow:
		return "stow"
	case ActionUnstow:
		return "unstow"
	default:
		return "unknown action"
	}
}

type CommandContext struct {
	Action           Action
	Args             []string
	ConflictStrategy ResolveStrategy
}

type Engine struct {
	source      string
	destination string
	ignore      *IgnoreList
	logger      *slog.Logger
	fileSystem  FileSystem
	dryRun      bool
}

func NewEngine(cfg *config.Config, dryRun bool, l *slog.Logger) (*Engine, error) {
	var handler FileSystem
	if dryRun {
		handler = file.NewNoWriteHandler(l)
	} else {
		handler = file.NewHandler(l)
	}
	ignoreList, err := newIgnoreList(cfg.Source, handler, l)
	if err != nil {
		return nil, err
	}
	return &Engine{
		source:      cfg.Source,
		destination: cfg.Destination,
		ignore:      ignoreList,
		logger:      l.With("component", "engine"),
		fileSystem:  handler,
		dryRun:      dryRun,
	}, nil
}

// Execute will execute the operation (stow, unstow) with the provided context.
func (e *Engine) Execute(ctx *CommandContext) (*ExecuteSummary, error) {
	actions, err := e.populateOperations(ctx)
	if err != nil {
		return nil, err
	}
	if err := e.checkPreflight(actions); err != nil {
		return nil, err
	}
	summary, err := e.executeFileActions(actions)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (e *Engine) checkPreflight(actions []fileAction) error {
	errorList := make([]error, 0, len(actions))
	for _, action := range actions {
		if err := action.preflight(e.fileSystem); err != nil {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) > 0 {
		return &AggregatedError{
			Msg:   "preflight checks failed",
			Items: errorList,
		}
	}
	return nil
}

func (e *Engine) executeFileActions(actions []fileAction) (*ExecuteSummary, error) {
	summary := &Summary{}
	actionList := make([]ActionEvent, 0, len(actions))
	for _, action := range actions {
		fileActions, err := action.execute(e.fileSystem)
		if err != nil {
			return nil, err
		}
		actionType := action.kind()
		switch actionType {
		case UpToDate:
			summary.UpToDate += 1
		case Skip:
			summary.Skipped += 1
		case Link:
			summary.Stowed += 1
		case Replace:
			summary.Replaced += 1
		case Backup:
			summary.BackedUp += 1
		case Adopt:
			summary.Adopted += 1
		case Remove:
			summary.Unstowed += 1
		default:
			return nil, fmt.Errorf("undefined action %d", actionType)
		}
		actionList = append(actionList, fileActions...)
	}
	return &ExecuteSummary{
		Actions:          actionList,
		OperationSummary: summary,
		DryRun:           e.dryRun,
	}, nil
}

func (e *Engine) populatePackageList(args []string) ([]string, error) {
	e.logger.Debug("populating package list", "source", e.source)
	var pkgCandidates []string
	var err error
	if len(args) == 0 {
		e.logger.Debug("no packages provided; processing all packages")
		pkgCandidates, err = e.getAllPackages()
		if err != nil {
			return nil, err
		}
	} else {
		pkgCandidates, err = e.getPackagesFromArgs(args)
		if err != nil {
			return nil, err
		}
	}
	packages, err := e.filterPackages(pkgCandidates)
	if err != nil {
		return nil, err
	}
	e.logger.Debug("package list populated", "package_list", packages)
	return packages, nil
}

func (e *Engine) getAllPackages() ([]string, error) {
	dirs, err := e.fileSystem.ListDirs(e.source)
	if err != nil {
		return nil, err
	}
	candidates := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		candidate, err := filepath.Rel(e.source, dir)
		if err != nil {
			return nil, fmt.Errorf("rel %s %s: %w", e.source, dir, err)
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (e *Engine) getPackagesFromArgs(candidates []string) ([]string, error) {
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "." {
			return nil, &HintedError{
				Op:   fmt.Sprintf("read package %s", candidate),
				Hint: "move root files to suitable directory (`zsh/`, `bash/`, etc.)",
				Err:  ErrRootIsNotPkg,
			}
		}
		pkgPath := filepath.Clean(candidate)
		isDir, err := e.fileSystem.IsDir(filepath.Join(e.source, pkgPath))
		if err != nil {
			return nil, fmt.Errorf("read package %s: %w", candidate, err)
		}
		if !isDir {
			return nil, &HintedError{
				Op:   fmt.Sprintf("read package %s", candidate),
				Hint: fmt.Sprintf("make sure the %s is a directory", candidate),
				Err:  ErrPkgIsNotDir,
			}
		}
		result = append(result, pkgPath)
	}
	return result, nil
}

func (e *Engine) filterPackages(candidates []string) ([]string, error) {
	e.logger.Debug("filtering packages", "candidates", candidates, "filter", e.ignore.items)
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		shouldIgnore, err := e.ignore.shouldIgnorePkg(candidate)
		if err != nil {
			return nil, err
		}
		if shouldIgnore {
			e.logger.Debug("ignoring package candidate", "candidate", candidate)
			continue
		}
		e.logger.Debug("adding package to process", "package", candidate)
		result = append(result, candidate)
	}
	return result, nil
}
