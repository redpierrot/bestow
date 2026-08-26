/*
All Rights Reversed (ɔ)
*/

// Package engine is the main operation engine of the bestow.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/redpierrot/bestow/internal/file"
)

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
	DryRun      bool
}

// NewEngine returns an Engine value with the provided configs
func NewEngine(cfg *EngineConfig, l *slog.Logger) (*Engine, error) {
	var handler FileSystem
	if cfg.DryRun {
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
		dryRun:      cfg.DryRun,
	}, nil
}

// Stow executes the stow operation
func (e *Engine) Stow(ctx context.Context, args []string, strategy ResolveStrategy) (*ExecuteResult, error) {
	candidates, err := e.findCandidates(args)
	if err != nil {
		return nil, err
	}
	stowCmd := &StowCommand{
		Strategy: strategy,
	}
	actions, err := e.buildFileActions(candidates, stowCmd.resolve)
	if err != nil {
		return nil, err
	}
	return e.executeFileActions(ctx, actions)
}

// Unstow executes the unstow operation
func (e *Engine) Unstow(ctx context.Context, args []string, keepEmptyParents bool) (*ExecuteResult, error) {
	candidates, err := e.findCandidates(args)
	if err != nil {
		return nil, err
	}
	unstowCmd := &UnstowCommand{
		KeepEmptyParents: keepEmptyParents,
	}
	actions, err := e.buildFileActions(candidates, unstowCmd.resolve)
	if err != nil {
		return nil, err
	}
	cleanupActions := unstowCmd.cleanup(candidates, e.logger, e.destination)
	actions = append(actions, cleanupActions...)
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
		if _, ok := action.(undoableAction); ok {
			completedActions = append(completedActions, action)
		}
		e.logger.Debug("executed action", "action", action, "summary", summary)
	}
	return &ExecuteResult{events, summary, e.dryRun}, nil
}

func (e *Engine) undoFileActions(actions []fileAction, summary *Summary, events []ActionEvent) (*ExecuteResult, error) {
	// Undo the completed actions from the last action to the top
	for _, action := range slices.Backward(actions) {
		undoAction, ok := action.(undoableAction)
		if !ok {
			continue
		}
		operationEvents, err := undoAction.undo(e.fileSystem)
		if err != nil {
			return &ExecuteResult{events, summary, e.dryRun}, err
		}
		e.updateSummary(action, summary, true)
		events = append(events, operationEvents...)
	}
	return &ExecuteResult{events, summary, e.dryRun}, nil
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
