package file

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type ExistingType string

const (
	ExistingManagedSymlink ExistingType = "ExistingManagedSymlink"
	ExistingForeignSymlink ExistingType = "ExistingForeignSymlink"
	ExistingRegularFile    ExistingType = "ExistingRegularFile"
	ExistingDir            ExistingType = "ExistingDirectory"
)

const BackupFileExtension = "bak"

const filePermissions = 0755

type FileHandler struct {
	CreatedDirs map[string]bool
	Logger      *slog.Logger
	Dryrun      bool
}

func NewFileHandler(l *slog.Logger) FileHandler {
	return FileHandler{
		CreatedDirs: make(map[string]bool),
		Logger:      l.With("component", "file"),
	}
}

// ListAllFiles Lists all the files in a given directory, including the files in subdirectories.
// This does not include any directory inside the given path.
//
// Parameters:
//   - parent: The parent directory
//   - dirName: The name of the directory where the file list is needed
//
// Returns:
//   - []string: The list of files as complete paths
//   - error: A `FileError` if any of the step fails to complete the intended task
func (h *FileHandler) ListAllFiles(parent string, dirName string) ([]string, error) {
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
			subItems, err := h.ListAllFiles(parent, dirPath)
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

// Lists the files in a given directory, excluding the directories.
//
// Parameters:
//
//	path: Path of the parent directory
//
// Returns:
//
//	[]string: List of files in the provided parent directory
//	error: A "FileError" caused by any reason
func (h *FileHandler) ListFiles(path string) ([]string, error) {
	isDir, err := h.IsDir(path)
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
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		result = append(result, file.Name())
	}
	h.Logger.Debug("found files in the directory", "directory", path, "files", result)
	return result, nil
}

func (h *FileHandler) ListDirs(path string) ([]string, error) {
	h.Logger.Debug("listing all the directories", "source", path)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, &FileError{
			Message: "failed to read the files",
			Path:    path,
			Cause:   err,
		}
	}
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			result = append(result, file.Name())
		}
	}
	h.Logger.Debug("finished searching directories", "dirs", result)
	return result, nil
}

func (h *FileHandler) CreateFile(fileName string, path string, data string) error {
	fullFileName := filepath.Join(path, fileName)
	h.Logger.Debug("writing to file", "file", fullFileName)
	if err := os.WriteFile(fullFileName, []byte(data), 0644); err != nil {
		return &FileError{
			Message: "failed to write to file",
			Path:    fullFileName,
			Cause:   err,
		}
	}
	h.Logger.Debug("successfully written to fiile", "path", fullFileName)
	return nil
}

func (h *FileHandler) CreateDir(path string) error {
	logger := h.Logger.With("path", path)
	logger.Debug("Creating directory")
	if h.CreatedDirs[path] {
		logger.Debug("directory already created")
		return nil
	}
	exists, err := h.Exists(path)
	if err != nil {
		return &FileError{
			Message: "failed to read the path",
			Path:    path,
			Cause:   err,
		}
	}
	if exists {
		logger.Debug("directory already exists")
		h.CreatedDirs[path] = true
		return nil
	}
	if err := os.MkdirAll(path, filePermissions); err != nil {
		return &FileError{
			Message: "failed to create directory",
			Path:    path,
			Cause:   err,
		}
	}
	h.CreatedDirs[path] = true
	logger.Debug("created directory", "path", path)
	return nil
}

func (h *FileHandler) IsDir(path string) (bool, error) {
	stat, err := h.getFileInfo(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

func (h *FileHandler) Exists(path string) (bool, error) {
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

func (h *FileHandler) ReadLines(path string) ([]string, error) {
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

func (h *FileHandler) GetPathSegments(path string) []string {
	parent, child := filepath.Split(path)
	if parent == "" || parent == "/" {
		return []string{child}
	}
	return append(h.GetPathSegments(filepath.Clean(parent)), child)
}

func (h *FileHandler) GetExistingFileType(src, dest string) (ExistingType, error) {
	h.Logger.Debug("checking existing file type", "source", src, "destination", dest)
	stat, err := os.Lstat(dest)
	if err != nil {
		return ExistingRegularFile, &FileError{
			Message: "failed to read the path",
			Path:    dest,
			Cause:   err,
		}
	}
	if stat.Mode().IsRegular() {
		h.Logger.Debug("found regular file")
		return ExistingRegularFile, nil
	}
	if stat.IsDir() {
		h.Logger.Debug("found directory", "path", dest)
		return ExistingDir, nil
	}

	h.Logger.Debug("found symlink", "path", dest)
	srcInfo, err := h.getFileInfo(src)
	if err != nil {
		return ExistingForeignSymlink, err
	}
	destInfo, err := h.getFileInfo(dest)
	if err != nil {
		return ExistingForeignSymlink, err
	}
	if os.SameFile(srcInfo, destInfo) {
		h.Logger.Debug("found managed symlink", "source", src, "destination", dest)
		return ExistingManagedSymlink, nil
	}
	return ExistingForeignSymlink, nil
}

func (h *FileHandler) getFileInfo(path string) (os.FileInfo, error) {
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

func (h *FileHandler) Link(src, dest string) error {
	destParent := filepath.Dir(dest)
	if err := h.CreateDir(destParent); err != nil {
		return err
	}
	if err := os.Symlink(src, dest); err != nil {
		return &FileError{
			Message: "failed to create symlink",
			Path:    dest,
			Cause:   err,
		}
	}
	h.Logger.Debug("link created", "source", src, "destination", dest)
	return nil
}

func (h *FileHandler) Move(src, target string) error {
	if err := os.Rename(src, target); err != nil {
		return err
	}
	return nil
}

func (h *FileHandler) Remove(path string) error {
	if err := os.Remove(path); err != nil {
		return &FileError{
			Message: "failed to remove the existing symlink/file",
			Path:    path,
			Cause:   err,
		}
	}
	h.Logger.Debug("successfully removed the file", "file_name", path)
	return nil
}

func (h *FileHandler) RemoveEmptyDirectories(path string) error {
	h.Logger.Debug("removing all the empty directories in the parent", "parent", path)
	dirs, err := h.ListDirs(path)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		err = h.RemoveDirIfEmpty(filepath.Join(path, dir))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *FileHandler) RemoveDirIfEmpty(path string) error {
	isDir, err := h.IsDir(path)
	if err != nil {
		return err
	}
	if !isDir {
		return nil
	}
	files, err := h.ListFiles(path)
	if err != nil {
		return err
	}
	if len(files) > 0 {
		h.Logger.Debug("directory is not empty", "path", path)
		return nil
	}
	dirs, err := h.ListDirs(path)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if err := h.RemoveDirIfEmpty(filepath.Join(path, dir)); err != nil {
			return err
		}
	}
	isEmpty, err := h.IsEmpty(path)
	if err != nil {
		return err
	}
	if isEmpty {
		h.Logger.Debug("directory is empty, removing", "path", path)
		return h.Remove(path)
	}
	return nil
}

func (h *FileHandler) Copy(src, dest string) error {
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

func (h *FileHandler) IsEmpty(path string) (bool, error) {
	isDir, err := h.IsDir(path)
	if err != nil {
		return false, err
	}
	if !isDir {
		return false, &FileError{
			Message: "provided path is not a directory",
			Path:    path,
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return false, &FileError{
			Message: "failed to read from the path",
			Path:    path,
			Cause:   err,
		}
	}
	defer f.Close()

	_, err = f.ReadDir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, nil
}
