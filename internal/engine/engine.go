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
	Ignore      IgnoreList
	Logger      *slog.Logger
	FileSystem  file.FileSystem
}

func NewEngine(cfg *config.Config, dryrun bool, l *slog.Logger) (*Engine, error) {
	ignoreList, err := newIgnoreList(cfg.Source, l)
	if err != nil {
		return nil, &EngineError{
			Message: "failed to initialize the ignore list",
			Cause:   err,
		}
	}
	handler := file.NewFileHandler(l) // TODO: Pass "remove empty parents" parameter
	return &Engine{
		Source:      cfg.Source,
		Destination: cfg.Destination,
		Ignore:      *ignoreList,
		Logger:      l.With("component", "engine"),
		FileSystem:  &handler,
	}, nil
}

func (e *Engine) Execute(ctx *CommandContext, args *[]string) error {
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
	packages, err := e.filterPackages(pkgCandidates, e.Ignore)
	if err != nil {
		return nil, err
	}
	e.Logger.Debug("package list populated", "package_list", packages)
	return packages, nil
}

func (e *Engine) getAllPackages() ([]string, error) {
	candidates, err := e.FileSystem.ListDirs(e.Source)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}

func (e *Engine) getPackagesFromArgs(candidates []string) ([]string, error) {
	result := []string{}
	for _, candidate := range candidates {
		if candidate == "." {
			return nil, &EngineError{
				Message: "invalid package; root is not considered a package",
				Hint:    "move root files to suitable directory (`zsh/`, `bash/`, etc.)",
			}
		}
		pkgPath := filepath.Clean(candidate)
		isDir, err := e.FileSystem.IsDir(filepath.Join(e.Source, pkgPath))
		if err != nil {
			return nil, &EngineError{
				Message: "failed to read the package",
				Cause:   err,
			}
		}
		if !isDir {
			return nil, &EngineError{
				Message: "failed to read the package; path is not a directory",
				Hint:    fmt.Sprintf("path: %s", pkgPath),
			}
		}
		result = append(result, pkgPath)
	}
	return result, nil
}

func (e *Engine) filterPackages(candidates []string, ignoreList IgnoreList) ([]string, error) {
	e.Logger.Debug("filtering packages", "candidates", candidates, "filter", ignoreList.items)
	result := []string{}
	for _, candidate := range candidates {
		shouldIgnore, err := ignoreList.shouldIgnore(candidate, "")
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
