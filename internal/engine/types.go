/*
All Rights Reversed (ɔ)
*/

package engine

import "github.com/redpierrot/bestow/internal/file"

// Type safety for File System Implementations
var _ FileSystem = (*file.Handler)(nil)
var _ FileSystem = (*file.DryRunHandler)(nil)

// ExecuteResult stores the result of a given execution
type ExecuteResult struct {
	// Events is the list of events (file operations) performed in the execution
	Events []ActionEvent
	// Summary is the summary of each operation
	Summary *Summary
	// DryRun whether the operation is a dry run
	DryRun bool
}

// Summary stores a summary of all the operations performed during an execution
type Summary struct {
	counts   [numActionKinds]int
	reverted int
}

func (s *Summary) Count(k ActionKind) int {
	return s.counts[k]
}

func (s *Summary) Reverted() int {
	return s.reverted
}

// FileSystem defines the operations needed for the engine to perform file operations
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
