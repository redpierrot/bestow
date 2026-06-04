/*
All Rights Reversed (ɔ)
*/

package engine

import "github.com/ThisaruGuruge/bestow/internal/file"

// Type safety for File System Implementations
var _ FileSystem = (*file.Handler)(nil)
var _ FileSystem = (*file.NoWriteHandler)(nil)

type ExecuteSummary struct {
	Actions          []ActionEvent
	OperationSummary *Summary
	DryRun           bool
}

type Summary struct {
	Stowed   int
	Unstowed int
	Replaced int
	BackedUp int
	Adopted  int
	Skipped  int
	UpToDate int
}

type FileSystem interface {
	// ListFiles returns a list of all the files in a given parent directory, excluding the directories.
	// The file list includes the full paths of the files found.
	ListFiles(parent string) ([]string, error)

	// ListDirs lists all the subdirectories in a given parent directory.
	ListDirs(parent string) ([]string, error)

	// ListAllFiles lists all the files in a given parent directory, including the files in the subdirectories.
	ListAllFiles(parent string) ([]string, error)

	// CreateFile creates a file in the provided path and writes the provided content to the file.
	CreateFile(path, content string) error

	// CreateDir creates a directory on the provided path, including all the parent directories.
	CreateDir(path string) error

	// Link creates a symlink of the provided src in the provided target.
	Link(src, target string) error

	// Move moves a file from src to target.
	Move(src, target string) error

	// Remove removes the file in the provided path.
	Remove(path string) error

	// IsDir checks whether the provided path is a directory.
	IsDir(path string) (bool, error)

	// IsEmptyDir returns true if the provided path is empty. It will return an error is the provided path is not a
	// directory.
	IsEmptyDir(path string) (bool, error)

	// Exists returns true if the provided path exists.
	Exists(path string) (bool, error)

	// ReadLines reads the content of a file in the given path and returns the lines of text as a list of strings.
	ReadLines(path string) ([]string, error)

	// GetExistingFileType returns the type of the existing dst compared to the provided src. Possible values are:
	//   - ExistingRegularFile: dst is a regular file
	//   - ExistingManagedSymlink: dst is a symlink that is managed by bestow
	//   - ExistingForeignSymlink: dst is a symlink that is not managed by bestow
	//   - ExistingDir: dst is a directory
	GetExistingFileType(src, dest string) (file.ExistingType, error)
}
