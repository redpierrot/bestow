package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/log"
)

var createdDirs = make(map[string]bool)

type FileError struct {
	Message string
	Path    string
	Cause   error
}

func (e *FileError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Message, e.Path, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Path)
}

func (e *FileError) Unwrap() error { return e.Cause }

type ExistingType string

const (
	ExistingManagedSymlink ExistingType = "ExistingManagedSymlink"
	ExistingForeignSymlink ExistingType = "ExistingForeignSymlink"
	ExistingRegularFile    ExistingType = "ExistingRegularFile"
	ExistingDir            ExistingType = "ExistingDirectory"
)

const BackupFileExtension = "bak"

// Lists all the files in a given directory. The direcrtory path should be given as
// the parent directory name and the directory name.
// It will throw errors if the paths are incorrect or there are permission issues/IO issues
// when reading the directory.
// This calls itself recursively to get all the files (including the files inside subdirectories).
// The result is added to the `fileList` provided.
// No directory will be listed in the file list, only the files.
func ListAllFilesInDir(parent string, dirName string) ([]string, error) {
	directoryPath := filepath.Join(parent, dirName)
	if directoryPath == "" {
		return nil, &FileError{
			Message: "path name is empty",
			Path:    parent,
		}
	}
	stat, statErr := os.Stat(directoryPath)
	if os.IsNotExist(statErr) {
		return nil, &FileError{
			Message: "provided directory path not found",
			Path:    directoryPath,
			Cause:   statErr,
		}
	}
	if !stat.IsDir() {
		return nil, &FileError{
			Message: "provided path is not a directory",
			Path:    directoryPath,
		}
	}

	files, err := os.ReadDir(directoryPath)
	if err != nil {
		return nil, &FileError{
			Message: "failed to read the content of the directory",
			Path:    directoryPath,
			Cause:   err,
		}
	}
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			dirPath := filepath.Join(dirName, file.Name())
			subItems, err := ListAllFilesInDir(parent, dirPath)
			if err != nil {
				return nil, err
			}
			result = append(result, subItems...)
		} else {
			fileName := filepath.Join(dirName, file.Name())
			result = append(result, fileName)
		}
	}
	return result, nil
}

// Lists all the files in a given directory
// will return error if the provided path:
// - does not exist
// - is not a directory
// - is not readable/accessible
func ListFiles(path string) ([]string, error) {
	isDir, err := IsDir(path)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, &FileError{
			Message: "provided path is not a directory",
			Path:    path,
		}
	}
	files, err := os.ReadDir(path)
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		result = append(result, file.Name())
	}
	return result, nil
}

func ListAllDirectories(source string) ([]string, error) {
	log.Debug("listing all the directories", "source", source)
	files, err := os.ReadDir(source)
	if err != nil {
		return nil, &FileError{
			Message: "failed to read the files",
			Path:    source,
			Cause:   err,
		}
	}
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			result = append(result, file.Name())
		}
	}
	log.Debug("finished searching directories", "dirs", result)
	return result, nil
}

func CreateFile(fileName string, path string, data string) error {
	fullFileName := filepath.Join(path, fileName)
	log.Debug("writing to file", "file", fullFileName)
	if err := os.WriteFile(fullFileName, []byte(data), 0644); err != nil {
		return &FileError{
			Message: "failed to write to file",
			Path:    fullFileName,
			Cause:   err,
		}
	}
	log.Debug("successfully written to fiile", "path", fullFileName)
	return nil
}

func CreateDir(path string) error {
	if createdDirs[path] {
		return nil
	}
	log.Debug("creating dir", "path", path)
	exists, err := Exists(path)
	if err != nil {
		return &FileError{
			Message: "failed to read the path",
			Path:    path,
			Cause:   err,
		}
	}
	if exists {
		log.Debug("directory already exists", "path", path)
		return nil
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return &FileError{
			Message: "failed to create directory",
			Path:    path,
			Cause:   err,
		}
	}
	createdDirs[path] = true
	log.Info("created directory", "path", path)
	return nil
}

func IsDir(path string) (bool, error) {
	stat, err := getFileInfo(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, &FileError{
			Message: "failed to read the file",
			Path:    path,
			Cause:   err,
		}
	}
	return true, nil
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, &FileError{
			Message: "error occurred while opening the file",
			Path:    path,
			Cause:   err,
		}
	}
	defer file.Close()
	result := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, &FileError{
			Message: "failed to read the file",
			Path:    path,
			Cause:   err,
		}
	}
	return result, nil
}

func GetPathSegments(path string) []string {
	parent, child := filepath.Split(path)
	if parent == "" || parent == "/" {
		return []string{child}
	}
	return append(GetPathSegments(filepath.Clean(parent)), child)
}

func GetExistingFileType(src, dest string) (ExistingType, error) {
	log.Debug("checking existing file type", "source", src, "destination", dest)
	stat, err := os.Lstat(dest)
	if err != nil {
		return ExistingRegularFile, &FileError{
			Message: "failed to read the path",
			Path:    dest,
			Cause:   err,
		}
	}
	if stat.Mode().IsRegular() {
		log.Debug("found regular file")
		return ExistingRegularFile, nil
	}
	if stat.IsDir() {
		log.Debug("found directory", "path", dest)
		return ExistingDir, nil
	}

	log.Debug("found symlink", "path", dest)
	srcInfo, err := getFileInfo(src)
	if err != nil {
		return ExistingForeignSymlink, err
	}
	destInfo, err := getFileInfo(dest)
	if err != nil {
		return ExistingForeignSymlink, err
	}
	if os.SameFile(srcInfo, destInfo) {
		log.Debug("found managed symlink", "source", src, "destination", dest)
		return ExistingManagedSymlink, nil
	}
	return ExistingForeignSymlink, nil
}

func getFileInfo(path string) (os.FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, &FileError{
			Message: "failed to read from the path",
			Path:    path,
			Cause:   err,
		}
	}
	return stat, nil
}

func Link(src, dest string) error {
	destParent := filepath.Dir(dest)
	if err := CreateDir(destParent); err != nil {
		return err
	}
	if err := os.Symlink(src, dest); err != nil {
		return &FileError{
			Message: "failed to create symlink",
			Path:    dest,
			Cause:   err,
		}
	}
	log.Info("link created", "source", src, "destination", dest)
	return nil
}

func Remove(path string) error {
	if err := os.Remove(path); err != nil {
		return &FileError{
			Message: "failed to remove the existing symlink/file",
			Path:    path,
			Cause:   err,
		}
	}
	return nil
}

func Backup(path string) error {
	log.Debug("backing up file", "path", path)
	fileName := filepath.Base(path)
	newFileName := strings.Join([]string{fileName, constant.AppName, BackupFileExtension}, ".")
	newFullPath := filepath.Join(filepath.Dir(path), newFileName)
	if err := os.Rename(path, newFullPath); err != nil {
		return &FileError{
			Message: "failed to rename the file",
			Path:    path,
			Cause:   err,
		}
	}
	log.Info("successfully backed up the file", "old", path, "new", newFullPath)
	return nil
}
func Copy(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return &FileError{
			Message: "failed to read the file",
			Path:    src,
			Cause:   err,
		}
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return &FileError{
			Message: "failed to open the file",
			Path:    dest,
			Cause:   err,
		}
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
