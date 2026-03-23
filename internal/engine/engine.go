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

type ActionContext struct {
	Action Action
	Args   []string
	DryRun bool
}

type ExecutionContext struct {
	Source      string
	Destination string
	PackageList []string
	Ignore      IgnoreList
}

func Execute(ctx *ActionContext, cfg *config.Config) error {
	log.Debug("executing", "config", cfg, "context", ctx)
	ignoreList, err := newIgnoreList(cfg.Source)
	executionCtx := &ExecutionContext{
		Source:      cfg.Source,
		Destination: cfg.Destination,
		Ignore:      *ignoreList,
	}
	packageList, err := getPackageList(ctx, cfg, ignoreList)
	if err != nil {
		return &EngineError{
			Message: "failed to read the files",
			Command: ctx.Action,
			Cause:   err,
		}
	}
	executionCtx.PackageList = *packageList
	log.Debug("packages received", "packages", packageList)
	log.Info("found candidates", "context", executionCtx)
	operations, err := populateOperations(*ctx, *executionCtx)
	if err != nil {
		return &EngineError{
			Message: "failed to opulate operations",
			Command: ctx.Action,
			Cause:   err,
		}
	}
	output.Success(fmt.Sprint("operation: ", ctx.Action))
	// I'm sorry, but for sentimental reasons, I will not accept any AI-generated PRs here.
	if err := validateOperations(&operations); err != nil {
		return &EngineError{
			Message: "invalid operation",
			Command: ctx.Action,
			Cause:   err,
		}
	}
	return nil
}

func getPackageList(ctx *ActionContext, cfg *config.Config, ignoreList *IgnoreList) (*[]string, error) {
	source := cfg.Source
	log.Debug("retrieving package list", "source", source)
	var pkgCandidates []string
	if len(ctx.Args) == 0 {
		log.Warn("no packages provided, processing all the packages", "action", ctx.Action)
		var err error
		pkgCandidates, err = file.ListAllDirectories(source)
		if err != nil {
			return nil, &EngineError{
				Message: "failed to read packages from source",
				Command: ctx.Action,
				Cause:   err,
			}
		}
	} else {
		for _, pkgCandidate := range ctx.Args {
			isDir, err := file.IsDir(filepath.Join(source, pkgCandidate))
			if err != nil {
				return nil, &EngineError{
					Message: "failed to read package",
					Command: ctx.Action,
					Package: pkgCandidate,
					Cause:   err,
				}
			}
			if !isDir {
				return nil, &EngineError{
					Message: "failed to read package",
					Package: pkgCandidate,
					Command: ctx.Action,
				}
			}
			pkgCandidates = append(pkgCandidates, pkgCandidate)
		}

		log.Debug("retrieved candidates for packages", "candidates", pkgCandidates)
	}
	packages, err := filterPackages(pkgCandidates, *ignoreList)
	if err != nil {
		return nil, err
	}

	return &packages, nil
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
