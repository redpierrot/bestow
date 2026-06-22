/*
All Rights Reversed (ɔ)
*/

package engine

import "github.com/ThisaruGuruge/bestow/internal/file"

// Type safety for File System Implementations
var _ FileSystem = (*file.Handler)(nil)
var _ FileSystem = (*file.DryRunHandler)(nil)

type ExecuteResult struct {
	Events  []ActionEvent
	Summary *Summary
	DryRun  bool
}

type Summary struct {
	Stowed   int
	Unstowed int
	Replaced int
	BackedUp int
	Adopted  int
	Skipped  int
	UpToDate int
	Reverted int
}

type FileSystem interface {
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

	// IsEmptyDir returns true if the provided path is empty. Returns true if the path is a directory. False, if the path
	// is not a directory. Returns an error if any IO error occurred.
	IsEmptyDir(path string) (bool, error)

	// Exists returns true if the provided path exists.
	Exists(path string) (bool, error)

	// ReadLines reads the content of a file in the given path and returns the lines of text as a list of strings.
	ReadLines(path string) ([]string, error)

	// ExistingFileType returns the type of the existing dest compared to the provided src. Possible values are:
	//   - ExistingRegularFile: dest is a regular file
	//   - ExistingManagedSymlink: dest is a symlink that is managed by bestow
	//   - ExistingForeignSymlink: dest is a symlink that is not managed by bestow
	//   - ExistingDir: dest is a directory
	ExistingFileType(src, dest string) (file.ExistingType, error)
}
