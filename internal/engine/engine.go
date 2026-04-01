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

type EngineError struct {
	Message string
	Command Action
	Package string
	Cause   error
}

func (e *EngineError) Error() string {
	msg := e.Message
	if e.Command != "" {
		msg += fmt.Sprint(": [%s]", e.Command)
	}
	if e.Package != "" {
		msg += fmt.Sprint(": [%s]", e.Package)
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
	}, nil
}

func (e *Engine) Execute() error {
	operations, err := e.buildOperations()
	if err != nil {
		return err
	}
	// TODO: Check dry run and execute the correct one
	// TODO: Check the length of the operations array after filtering and fail if len = 0
	e.executeDryRun(operations)
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
			Message: "failed to opulate operations",
			Command: e.Action,
			Cause:   err,
		}
	}
	// I'm sorry, but for sentimental reasons, I will not accept any AI-generated PRs here.
	operations, err = e.resolveOperations(&operations)
	if err != nil {
		return operations, &EngineError{
			Message: "invalid operation",
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
		isDir, err := file.IsDir(filepath.Join(src, candidate))
		if err != nil {
			return nil, err
		}
		if !isDir {
			return nil, err
		}
		result = append(result, candidate)
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
