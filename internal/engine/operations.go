package engine

import (
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/file"
	"github.com/ThisaruGuruge/bestow/internal/log"
)

type ResolveStrategy string

const (
	ResolveSkip        ResolveStrategy = "skip"
	ResolveForce       ResolveStrategy = "force"
	ResolveAdopt       ResolveStrategy = "adopt"
	ResolveBackup      ResolveStrategy = "backup"
	ResolveInteractive ResolveStrategy = "interactive"
)

type Operation struct {
	Source      string
	Destination string
	BackupPath  string
	Action      FileAction
}

func (e *Engine) populateOperations() ([]Operation, error) {
	result := []Operation{}
	var action Action = e.Action
	var source, destination string
	if action == ActionUnstow {
		source = e.Destination
		destination = e.Source
	} else {
		source = e.Source
		destination = e.Destination
	}

	result, err := e.getRootOperation(result, source, destination)
	if err != nil {
		return nil, err
	}

	for _, pkg := range *e.PackageList {
		pacakgeOperations, err := e.getPackageOperation(pkg)
		if err != nil {
			return nil, err
		}
		result = append(result, pacakgeOperations...)
	}
	return result, nil
}

func (e *Engine) getRootOperation(operations []Operation, src, dest string) ([]Operation, error) {
	rootFileList, err := file.ListFiles(e.Source)
	if err != nil {
		return nil, &EngineError{
			Message: "failed to read files from the source root",
			Package: ".",
			Cause:   err,
		}
	}
	for _, fileName := range rootFileList {
		doIgnore, err := e.Ignore.shouldIgnore(fileName, constant.RootPackageName)
		if err != nil {
			return nil, err
		}
		if doIgnore {
			log.Debug("ignoring the file", "fileName", fileName)
			continue
		}
		srcFile := filepath.Join(src, fileName)
		destFile := filepath.Join(dest, fileName)
		operations = append(operations, Operation{
			Source:      srcFile,
			Destination: destFile,
			Action:      FileActionLink,
		})
	}
	return operations, nil
}

func (e *Engine) getPackageOperation(pkg string) ([]Operation, error) {
	sourceFileList := []string{}
	err := file.ListAllFilesInDir(e.Source, pkg, &sourceFileList)
	if err != nil {
		return nil, &EngineError{
			Message: "failed to read the package contents",
			Cause:   err,
		}
	}
	operations := []Operation{}
	for _, fileName := range sourceFileList {
		doIgnore, err := e.Ignore.shouldIgnore(fileName, pkg)
		if err != nil {
			return nil, err
		}
		if doIgnore {
			log.Debug("ignoring the file", "fileName", fileName, "package", pkg)
			continue
		}
		srcFile := filepath.Join(e.Source, fileName)
		relativePath := filepath.Join(file.GetPathSegments(fileName)[1:]...)
		destFile := filepath.Join(e.Destination, relativePath)

		operations = append(operations, Operation{
			Source:      srcFile,
			Destination: destFile,
			Action:      FileActionLink,
		})
	}
	return operations, nil
}

// TODO: Handle interactive/non-interactive modes.
// Pass config here to check interactivity and conflict resolution strategy
// Should return error in any invalid scenario.
func (e *Engine) resolveOperations(operations *[]Operation) ([]Operation, error) {
	for i := range *operations {
		if err := e.resolveOperation(&(*operations)[i]); err != nil {
			return *operations, err
		}
	}
	return *operations, nil
}

func (e *Engine) resolveOperation(operation *Operation) error {
	destExists, _ := file.Exists(operation.Destination)
	if destExists {
		// TODO: Doesn't make any sense for the static resolver. But we need this when we have interactive mode
		existing, err := file.GetExistingFileType(operation.Source, operation.Destination)
		if err != nil {
			return &EngineError{
				Message: "failed to check exising file type",
				Command: e.Action,
				Cause:   err,
			}
		}
		resolver := StaticResolver{strategy: e.Strategy}
		strategy, _ := resolver.Resolve(e.Source, e.Destination, existing)
		if err := e.resolveFileAction(operation, strategy, existing); err != nil {
			return err
		}
	}
	return nil
}
