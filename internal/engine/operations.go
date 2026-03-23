package engine

import (
	"fmt"
	"path/filepath"

	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/file"
	"github.com/ThisaruGuruge/bestow/internal/log"
)

type ConflictResolution string

const (
	ConflictSkip        ConflictResolution = "skip"
	ConflictForce       ConflictResolution = "force"
	ConflictAdopt       ConflictResolution = "adopt"
	ConflictBackup      ConflictResolution = "backup"
	ConflictInteractive ConflictResolution = "interactive"
)

type Operation struct {
	Source      string
	Destination string
	Package     string
	Steps       []Step
}

type Step struct {
	SourceFilePath      string
	DestinationFilePath string
	Conflict            ConflictResolution
	BackupFile          string
}

func populateOperations(actionCtx ActionContext, executionCtx ExecutionContext) ([]Operation, error) {
	result := []Operation{}
	var action Action = actionCtx.Action
	var source, destination string
	if action == ActionUnstow {
		source = executionCtx.Destination
		destination = executionCtx.Source
	} else {
		source = executionCtx.Source
		destination = executionCtx.Destination
	}

	rootOperation := Operation{
		Source:      source,
		Destination: destination,
		Package:     constant.RootPackageName,
	}
	if err := populateStepsForRoot(&rootOperation, executionCtx.Ignore); err != nil {
		return nil, err
	}
	result = append(result, rootOperation)

	for _, pkg := range executionCtx.PackageList {
		operation := Operation{
			Source:      source,
			Destination: destination,
			Package:     pkg,
		}
		populateStepsForPackage(&operation, executionCtx.Ignore)
		result = append(result, operation)
	}
	return result, nil
}

func populateStepsForRoot(operation *Operation, ignoreList IgnoreList) error {
	rootFileList, err := file.ListFiles(operation.Source)
	if err != nil {
		return &EngineError{
			Message: "failed to read files from the source root",
			Package: ".",
			Cause:   err,
		}
	}
	steps := []Step{}
	for _, fileName := range rootFileList {
		doIgnore, err := ignoreList.shouldIgnore(fileName, constant.RootPackageName)
		if err != nil {
			return err
		}
		if doIgnore {
			log.Debug("ignoring the file", "fileName", fileName)
			continue
		}
		steps = append(steps, Step{
			SourceFilePath:      fileName,
			DestinationFilePath: fileName,
		})
	}
	operation.Steps = steps
	return nil
}

func populateStepsForPackage(operation *Operation, ignoreList IgnoreList) error {
	sourceFileList := []string{}

	err := file.ListAllFilesInDir(operation.Source, operation.Package, &sourceFileList)
	if err != nil {
		return &EngineError{
			Message: "failed to read the package contents",
			Cause:   err,
		}
	}
	steps := []Step{}
	for _, fileName := range sourceFileList {
		doIgnore, err := ignoreList.shouldIgnore(fileName, operation.Package)
		if err != nil {
			return err
		}
		if doIgnore {
			log.Debug("ignoring the file", "fileName", fileName, "package", operation.Package)
			continue
		}
		pathSegments := file.GetPathSegments(fileName)[1:]
		destinationFilePath := filepath.Join(pathSegments...)
		steps = append(steps, Step{
			SourceFilePath:      fileName,
			DestinationFilePath: destinationFilePath,
		})
	}
	operation.Steps = steps
	return nil
}

// TODO: Handle interactive/non-interactive modes.
// Pass config here to check interactivity and conflict resolution strategy
// Should return error in any invalid scenario.
func validateOperations(operations *[]Operation) error {
	for _, operation := range *operations {
		validateOperation(&operation)
	}
	return nil
}

func validateOperation(operation *Operation) error {
	for _, step := range operation.Steps {
		validateStep(operation.Source, operation.Destination, step, conflict)
	}
	return nil
}

// TODO: We should read files and have a struct to defien out custom file type.
// It should include whether file exists, isLink, isDir, etc. Seems like we might need that repeatedly.
// No point of reading the same file twice, I suppose.
func validateStep(source, destination string, step Step, conflict ConflictResolution) error {
	destinationFileName := filepath.Join(destination, step.DestinationFilePath)
	sourceFileName := filepath.Join(source, step.SourceFilePath)

	exists, err := file.Exists(destinationFileName)
	if err != nil {
		return &EngineError{
			Message: "validation failed",
			Cause:   err,
		}
	}
	if exists {
		log.Warn("File exists")
		//TODO: check interactivity and conflict resolution
		return nil
	}
	log.Info(fmt.Sprintf("[Link] %s -> %s", sourceFileName, destinationFileName))

	return nil
}
