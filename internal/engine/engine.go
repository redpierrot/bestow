/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/redpierrot/bestow/internal/file"
)

// CommandAction defines the different actions in Bestow
type CommandAction int

const (
	CommandStow CommandAction = iota
	CommandUnstow
)

func (a CommandAction) String() string {
	switch a {
	case CommandStow:
		return "stow"
	case CommandUnstow:
		return "unstow"
	default:
		return "unknown action"
	}
}

// CommandConfig stores the configurations for a given command execution
type CommandConfig struct {
	Action           CommandAction
	Args             []string
	ConflictStrategy ResolveStrategy
}

// Engine is the brain of Bestow. It keeps the state of a given execution and handles all the file system calls
type Engine struct {
	source      string
	destination string
	ignore      *IgnoreList
	logger      *slog.Logger
	fileSystem  FileSystem
	configHome  string
	dryRun      bool
}

// EngineConfig is used to pass the configurations of the engine for a given execution
type EngineConfig struct {
	ConfigHome  string
	Source      string
	Destination string
}

// NewEngine returns an Engine value with the provided configs
func NewEngine(cfg *EngineConfig, dryRun bool, l *slog.Logger) (*Engine, error) {
	var handler FileSystem
	if dryRun {
		handler = file.NewDryRunHandler(l)
	} else {
		handler = file.NewHandler(l)
	}
	ignoreList, err := newIgnoreList(cfg.Source, cfg.ConfigHome, handler, l)
	if err != nil {
		return nil, err
	}
	return &Engine{
		source:      cfg.Source,
		destination: cfg.Destination,
		ignore:      ignoreList,
		logger:      l.With("component", "engine"),
		fileSystem:  handler,
		configHome:  cfg.ConfigHome,
		dryRun:      dryRun,
	}, nil
}

// Execute executes the given operation with the provided configs
func (e *Engine) Execute(ctx context.Context, cfg *CommandConfig) (*ExecuteResult, error) {
	actions, err := e.buildOperations(cfg)
	if err != nil {
		return nil, err
	}
	return e.executeFileActions(ctx, actions)
}

func (e *Engine) executeFileActions(ctx context.Context, actions []fileAction) (*ExecuteResult, error) {
	summary := &Summary{}
	events := make([]ActionEvent, 0, len(actions))
	completedActions := make([]fileAction, 0, len(actions))
	for _, action := range actions {
		// Handle cancellations mid operation
		if err := ctx.Err(); err != nil {
			undoResult, undoErr := e.undoFileActions(completedActions, summary, events)
			if undoErr != nil {
				return undoResult, fmt.Errorf("undo failed: %w; execution failed: %w", undoErr, err)
			}
			return undoResult, fmt.Errorf("operation interrupted; reverted changes: %w", err)
		}
		operationEvents, executeErr := action.execute(e.fileSystem)
		if executeErr != nil {
			undoResult, undoErr := e.undoFileActions(completedActions, summary, events)
			if undoErr != nil {
				return undoResult, fmt.Errorf("undo failed: %w; execution failed: %w", undoErr, executeErr)
			}
			return undoResult, executeErr
		}
		if err := e.updateSummary(action, summary, false); err != nil {
			return &ExecuteResult{
				Events:  events,
				Summary: summary,
				DryRun:  e.dryRun,
			}, err
		}
		events = append(events, operationEvents...)
		if kind := action.kind(); kind != ActionSkip && kind != ActionUpToDate {
			completedActions = append(completedActions, action)
		}
		e.logger.Debug("executed action", "action", action, "summary", summary)
	}
	return &ExecuteResult{
		Events:  events,
		Summary: summary,
		DryRun:  e.dryRun,
	}, nil
}

func (e *Engine) undoFileActions(actions []fileAction, summary *Summary, events []ActionEvent) (*ExecuteResult, error) {
	// Undo the completed actions from the last action to the top
	for _, action := range slices.Backward(actions) {
		operationEvents, err := action.undo(e.fileSystem)
		if err != nil {
			return &ExecuteResult{
				Events:  events,
				Summary: summary,
				DryRun:  e.dryRun,
			}, err
		}
		if err := e.updateSummary(action, summary, true); err != nil {
			return &ExecuteResult{
				Events:  events,
				Summary: summary,
				DryRun:  e.dryRun,
			}, err
		}
		events = append(events, operationEvents...)
	}
	return &ExecuteResult{
		Events:  events,
		Summary: summary,
		DryRun:  e.dryRun,
	}, nil
}

func (e *Engine) buildPackageList(args []string) ([]string, error) {
	e.logger.Debug("populating package list", "source", e.source)
	var pkgCandidates []string
	var err error
	if len(args) == 0 {
		e.logger.Debug("no packages provided; processing all packages")
		pkgCandidates, err = e.allPackages()
		if err != nil {
			return nil, err
		}
	} else {
		pkgCandidates, err = e.packagesFromArgs(args)
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

func (e *Engine) allPackages() ([]string, error) {
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

func (e *Engine) packagesFromArgs(candidates []string) ([]string, error) {
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
		shouldIgnore, err := e.ignore.isIgnored(candidate)
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

func (e *Engine) updateSummary(action fileAction, summary *Summary, isUndo bool) error {
	if isUndo {
		summary.Reverted += 1
		return nil
	}
	actionType := action.kind()
	switch actionType {
	case ActionUpToDate:
		summary.UpToDate += 1
	case ActionSkip:
		summary.Skipped += 1
	case ActionLink:
		summary.Stowed += 1
	case ActionReplace:
		summary.Replaced += 1
	case ActionBackup:
		summary.BackedUp += 1
	case ActionAdopt:
		summary.Adopted += 1
	case ActionRemove:
		summary.Unstowed += 1
	default:
		return fmt.Errorf("undefined action %d", actionType)
	}
	return nil
}
