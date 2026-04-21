package engine

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/ThisaruGuruge/bestow/internal/file"
	"github.com/ThisaruGuruge/bestow/internal/log"
	"github.com/ThisaruGuruge/bestow/internal/output"
)

type ResolveStrategy string

const (
	ResolveSkip        ResolveStrategy = "skip"
	ResolveForce       ResolveStrategy = "force"
	ResolveAdopt       ResolveStrategy = "adopt"
	ResolveBackup      ResolveStrategy = "backup"
	ResolveInteractive ResolveStrategy = "interactive"
)

const rootPackage = "."

type Operation struct {
	Source      string
	Destination string
	BackupPath  string
	Action      FileAction
}

func (e *Engine) populateOperations() ([]Operation, error) {
	result := []Operation{}
	if slices.Contains(*e.PackageList, rootPackage) {
		rootOperations, err := e.getRootOperation(result, e.Source, e.Destination)
		if err != nil {
			return nil, err
		}
		result = append(result, rootOperations...)
	}

	for _, pkg := range *e.PackageList {
		if pkg == rootPackage {
			continue
		}
		log.Debug("populating operations for package", "pacakge", pkg)
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
		doIgnore, err := e.Ignore.shouldIgnore(fileName, rootPackage)
		if err != nil {
			return nil, err
		}
		if doIgnore {
			log.Debug("ignoring the file", "fileName", fileName, "package", rootPackage)
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
	sourceFileList, err := file.ListAllFilesInDir(e.Source, pkg)
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
func (e *Engine) resolveStowOperations(operations *[]Operation) ([]Operation, error) {
	for i := range *operations {
		if err := e.resolveStowOperation(&(*operations)[i]); err != nil {
			return *operations, err
		}
	}
	return *operations, nil
}

func (e *Engine) resolveStowOperation(operation *Operation) error {
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

func (e *Engine) resolveUnstowOperations(operations *[]Operation) ([]Operation, error) {
	for i := range *operations {
		if err := e.resolveUnstowOperation(&(*operations)[i]); err != nil {
			return *operations, err
		}
	}
	return *operations, nil
}

func (e *Engine) resolveUnstowOperation(operation *Operation) error {
	destExists, err := file.Exists(operation.Destination)
	if err != nil {
		return &EngineError{
			Message: "failed to check the destination file",
			Command: e.Action,
			Cause:   err,
		}
	}
	if !destExists {
		log.Debug("file have not been stowed", "source", operation.Source)
		operation.Action = FileActionSkip
		return nil
	}
	existingType, err := file.GetExistingFileType(operation.Source, operation.Destination)
	if err != nil {
		return &EngineError{
			Message: "failed to check the destination file",
			Command: e.Action,
			Cause:   err,
		}
	}
	if existingType != file.ExistingManagedSymlink {
		operation.Action = FileActionSkip
		return nil
	}
	operation.Action = FileActionRemoveLink
	return nil
}

func (e *Engine) stow(operations []Operation) error {
	log.Debug("stowing files")
	// I'm sorry, but for sentimental reasons, I will not accept any AI-generated PRs here.
	operations, err := e.resolveStowOperations(&operations)
	if err != nil {
		return &EngineError{
			Message: "failed to resolve stow operation",
			Command: e.Action,
			Cause:   err,
		}
	}
	operations = filterSkipFiles(operations)
	if len(operations) == 0 {
		return &EngineError{
			Message: "no operations left for stow",
			Command: e.Action,
		}
	}
	for _, operation := range operations {
		if err := e.stowOperation(&operation); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) stowOperation(operation *Operation) error {
	switch operation.Action {
	case FileActionSkip:
		log.Debug("skipping file", "source", operation.Source, "destination", operation.Destination, "action", operation.Action)
		return nil
	case FileActionLink:
		return createLink(operation.Source, operation.Destination)
	case FileActionRemoveLink:
		return updateLink(operation.Source, operation.Destination)
	case FileActionBackupLink:
		return backupLink(operation.Source, operation.Destination)
	case FileActionAdoptLink:
		return adoptLink(operation.Source, operation.Destination)
	}
	return nil
}

func (e *Engine) unstow(operations []Operation) error {
	log.Debug("unstowing files", "pacakges", e.PackageList)
	operations, err := e.resolveUnstowOperations(&operations)
	if err != nil {
		return err
	}
	operations = filterSkipFiles(operations)
	if len(operations) == 0 {
		return &EngineError{
			Message: "no files left for unstow",
			Command: e.Action,
			Cause:   err,
		}
	}
	for _, operation := range operations {
		if err := e.unstowOperation(&operation); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) unstowOperation(operation *Operation) error {
	log.Debug("unstowing file", "source", operation.Source, "destination", operation.Destination)
	exists, err := file.Exists(operation.Destination)
	if err != nil {
		return &EngineError{
			Message: "failed to check the destination file",
			Command: e.Action,
			Cause:   err,
		}
	}
	if !exists {
		log.Debug("file not found for unstow", "path", operation.Destination)
		return nil
	}
	fileType, err := file.GetExistingFileType(operation.Source, operation.Destination)
	if err != nil {
		return &EngineError{
			Message: "failed to check the fily type",
			Command: e.Action,
			Cause:   err,
		}
	}
	if fileType != file.ExistingManagedSymlink {
		log.Warn("existing file is not managed by bestow", "existing_file", operation.Destination)
		return nil
	}
	err = file.Remove(operation.Destination)
	if err != nil {
		return &EngineError{
			Message: "failed to unstow the file",
			Command: e.Action,
			Cause:   err,
		}
	}
	log.Debug("successfully unstowed the file", "source", operation.Source, "destination", operation.Destination)
	return nil
}

func createLink(src, dest string) error {
	if err := file.Link(src, dest); err != nil {
		return &EngineError{
			Message: "failed to stow the file",
			Cause:   err,
		}
	}
	output.Success(fmt.Sprintf("created: %s", dest))
	return nil
}

func updateLink(src, dest string) error {
	if err := file.Remove(dest); err != nil {
		return &EngineError{
			Message: "failed to stow the file",
			Cause:   err,
		}
	}
	if err := createLink(src, dest); err != nil {
		return &EngineError{
			Message: "failed to stow the file",
			Cause:   err,
		}
	}
	return nil
}

func backupLink(src, dest string) error {
	if err := file.Backup(dest); err != nil {
		return &EngineError{
			Message: "failed to backup existing file",
			Cause:   err,
		}
	}
	if err := createLink(src, dest); err != nil {
		return err
	}
	return nil
}

func adoptLink(src, dest string) error {
	if err := file.Copy(dest, src); err != nil {
		return &EngineError{
			Message: "failed to adopt file from destination",
			Cause:   err,
		}
	}
	if err := file.Link(src, dest); err != nil {
		return &EngineError{
			Message: "failed to stow the file",
			Cause:   err,
		}
	}
	return nil
}
