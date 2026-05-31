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

type Action string

const (
	ActionInit   Action = "init"
	ActionStow   Action = "stow"
	ActionUnstow Action = "unstow"
)

type CommandContext struct {
	Action           Action
	Args             []string
	DryRun           bool
	ConflictStrategy ResolveStrategy
	Force            bool
	IgnoreList       []string
}

type Engine struct {
	Source      string
	Destination string
	Ignore      *IgnoreList
	Logger      *slog.Logger
	FileSystem  file.System
}

func NewEngine(cfg *config.Config, dryrun bool, l *slog.Logger) (*Engine, error) {
	var handler file.System
	if dryrun {
		handler = file.NewNoWriteHandler(l)
	} else {
		handler = file.NewHandler(l) // TODO: Pass "remove empty parents" parameter
	}
	ignoreList, err := newIgnoreList(cfg.Source, handler, l)
	if err != nil {
		return nil, err
	}
	return &Engine{
		Source:      cfg.Source,
		Destination: cfg.Destination,
		Ignore:      ignoreList,
		Logger:      l.With("component", "engine"),
		FileSystem:  handler,
	}, nil
}

func (e *Engine) Execute(ctx *CommandContext) error {
	if ctx.Action == ActionInit {
		if err := e.init(ctx); err != nil {
			return err
		}
		return nil
	}

	actions, err := e.populateOperations(ctx)
	if err != nil {
		return err
	}
	if err := e.executeFileActions(actions, ctx.DryRun); err != nil {
		return err
	}
	return nil
}

// TODO: When skipping files;
// - in .bestowignore: debug log
// - skip because already stowed (due to state of the operation): include a summary
// - skip because conflict resolution strategy is set to skip: print as same as any other operation
func (e *Engine) executeFileActions(actions []FileAction, dryrun bool) error {
	for _, action := range actions {
		if err := action.Execute(e.FileSystem, dryrun); err != nil {
			return err
		}
	}
	return nil
}

type Summary struct {
	stowed   int
	skipped  int
	upToDate int
}

func (e *Engine) populatePackageList(args []string) ([]string, error) {
	e.Logger.Debug("populating package list", "source", e.Source)
	var pkgCandidates []string
	var err error
	if len(args) == 0 {
		e.Logger.Debug("no packages provided; processing all packages")
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
	e.Logger.Debug("package list populated", "package_list", packages)
	return packages, nil
}

func (e *Engine) getAllPackages() ([]string, error) {
	dirs, err := e.FileSystem.ListDirs(e.Source)
	if err != nil {
		return nil, err
	}
	candidates := []string{}
	for _, dir := range dirs {
		candidate, err := filepath.Rel(e.Source, dir)
		if err != nil {
			return nil, fmt.Errorf("rel %s %s: %w", e.Source, dir, err)
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (e *Engine) getPackagesFromArgs(candidates []string) ([]string, error) {
	result := []string{}
	for _, candidate := range candidates {
		if candidate == "." {
			return nil, &HintedError{
				Op:   fmt.Sprintf("read package %s", candidate),
				Hint: "move root files to suitable directory (`zsh/`, `bash/`, etc.)",
				Err:  ErrRootIsNotPkg,
			}
		}
		pkgPath := filepath.Clean(candidate)
		isDir, err := e.FileSystem.IsDir(filepath.Join(e.Source, pkgPath))
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
	e.Logger.Debug("filtering packages", "candidates", candidates, "filter", e.Ignore.items)
	result := []string{}
	for _, candidate := range candidates {
		shouldIgnore, err := e.Ignore.shouldIgnore(candidate, "")
		if err != nil {
			return nil, err
		}
		if shouldIgnore {
			e.Logger.Debug("ignoring package candidate", "candidate", candidate)
			continue
		}
		e.Logger.Debug("adding package to process", "package", candidate)
		result = append(result, candidate)
	}
	return result, nil
}
