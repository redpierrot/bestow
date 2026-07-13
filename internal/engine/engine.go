/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/redpierrot/bestow/internal/file"
)

// CommandAction defines the different actions in Bestow
type CommandAction int

const (
	CommandStow CommandAction = iota
	CommandUnstow
)

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
		events = append(events, operationEvents...)
		if executeErr != nil {
			undoResult, undoErr := e.undoFileActions(completedActions, summary, events)
			if undoErr != nil {
				return undoResult, fmt.Errorf("undo failed: %w; execution failed: %w", undoErr, executeErr)
			}
			return undoResult, executeErr
		}
		e.updateSummary(action, summary, false)
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
		e.updateSummary(action, summary, true)
		events = append(events, operationEvents...)
	}
	return &ExecuteResult{
		Events:  events,
		Summary: summary,
		DryRun:  e.dryRun,
	}, nil
}

func (e *Engine) updateSummary(action fileAction, summary *Summary, isUndo bool) {
	if !isUndo {
		summary.counts[action.kind()]++
		return
	}
	if action.kind() != ActionSkip && action.kind() != ActionUpToDate {
		summary.reverted++
	}
}
