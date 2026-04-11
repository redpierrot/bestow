package engine

import (
	"fmt"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/file"
	"github.com/ThisaruGuruge/bestow/internal/log"
	"github.com/ThisaruGuruge/bestow/internal/output"
)

type Action string

const (
	ActionStow   Action = "stow"
	ActionUnstow Action = "unstow"
)

const RootDir string = "."

type EngineError struct {
	Message string
	Command Action
	Package string
	Cause   error
}

func (e *EngineError) Error() string {
	msg := e.Message
	if e.Command != "" {
		msg += fmt.Sprintf(": [%s]", e.Command)
	}
	if e.Package != "" {
		msg += fmt.Sprintf(": [%s]", e.Package)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	return msg
}

func (e *EngineError) Unwrap() error { return e.Cause }

type CommandContext struct {
	Action   Action
	Args     []string
	DryRun   bool
	Conflict ResolveStrategy
}

type ExecutionContext struct {
	Source      string
	Destination string
	PackageList []string
	Ignore      IgnoreList
}

type Engine struct {
	Source      string
	Destination string
	Ignore      IgnoreList
	Action      Action
	PackageList *[]string
	Strategy    ResolveStrategy
	DryRun      bool
}

func NewEngine(ctx *CommandContext, cfg *config.Config) (*Engine, error) {
	ignoreList, err := newIgnoreList(cfg.Source)
	if err != nil {
		return nil, &EngineError{
			Message: "failed to initialize the action",
			Command: ctx.Action,
			Cause:   err,
		}
	}
	packages, err := populatePackageList(cfg.Source, ctx.Args, *ignoreList)
	if err != nil {
		return nil, &EngineError{
			Message: "failed to initialize the action",
			Command: ctx.Action,
			Cause:   err,
		}
	}
	return &Engine{
		Source:      cfg.Source,
		Destination: cfg.Destination,
		Ignore:      *ignoreList,
		Action:      ctx.Action,
		PackageList: &packages,
		Strategy:    ctx.Conflict,
		DryRun:      ctx.DryRun,
	}, nil
}

func (e *Engine) Execute() error {
	operations, err := e.buildOperations()
	if err != nil {
		return err
	}
	e.executeOperations(operations)
	return nil
}

func (e *Engine) executeOperations(operations []Operation) error {
	switch e.Action {
	case ActionStow:
		e.stow(operations)
	case ActionUnstow:
		e.unstow(operations)
	}
	return nil
}

func (e *Engine) executeDryRun(operations []Operation) {
	for _, operation := range operations {
		output.Success(fmt.Sprintf("[%s]: %s -> %s", operation.Action, operation.Source, operation.Destination))
	}
}

func (e *Engine) buildOperations() ([]Operation, error) {
	operations, err := e.populateOperations()
	if err != nil {
		return nil, &EngineError{
			Message: "failed to populate operations",
			Command: e.Action,
			Cause:   err,
		}
	}
	return operations, nil
}

func populatePackageList(src string, args []string, ignore IgnoreList) ([]string, error) {
	log.Debug("populating package list", "source", src)
	var pkgCandidates []string
	var err error
	if len(args) == 0 {
		pkgCandidates, err = getAllPackages(src)
		if err != nil {
			return nil, err
		}
		// Add the root package
		pkgCandidates = append([]string{"."}, pkgCandidates...)
	} else {
		pkgCandidates, err = getPackagesFromArgs(src, args)
		if err != nil {
			return nil, err
		}
	}
	packages, err := filterPackages(pkgCandidates, ignore)
	if err != nil {
		return nil, err
	}
	log.Debug("package list populated", "package_list", packages)
	return packages, nil
}

func getAllPackages(src string) ([]string, error) {
	candidates, err := file.ListAllDirectories(src)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}

func getPackagesFromArgs(src string, candidates []string) ([]string, error) {
	result := []string{}
	for _, candidate := range candidates {
		pkgPath := filepath.Clean(candidate)
		isDir, err := file.IsDir(filepath.Join(src, pkgPath))
		if err != nil {
			return nil, err
		}
		if !isDir {
			return nil, err
		}
		result = append(result, pkgPath)
	}
	return result, nil
}

func filterPackages(candidates []string, ignoreList IgnoreList) ([]string, error) {
	log.Debug("filtering packages", "candidates", candidates, "filter", ignoreList.items)
	result := []string{}
	for _, candidate := range candidates {
		shouldIgnore, err := ignoreList.shouldIgnore(candidate, constant.RootPackageName)
		if err != nil {
			return nil, err
		}
		if shouldIgnore {
			log.Debug("Ignoring package candidate", "candidate", candidate)
			continue
		}
		result = append(result, candidate)
	}
	return result, nil
}

func filterSkipFiles(operations []Operation) []Operation {
	result := []Operation{}
	for _, operation := range operations {
		if operation.Action == FileActionSkip {
			continue
		}
		result = append(result, operation)
	}
	return result
}
