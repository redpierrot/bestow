package file

type FileSystem interface {
	// Lists all the files in a given directory, excluding the sub directories.
	//
	// Parameters:
	//     path: Path of the parent directory
	//
	// Returns:
	//     []string: List of files in the provided parent directory
	//     error: A FileError caused by any reason
	ListFiles(path string) ([]string, error)
	// Lists all the directories in a given directory
	ListDirs(source string) ([]string, error)
	// Lists all the files in a given directory, including the files in the subdirectories.
	ListAllFiles(parent string, dirName string) ([]string, error)

	CreateFile(fileName string, path string, data string) error
	CreateDir(path string) error
	Link(src, dest string) error
	Copy(src, dest string) error
	Move(src, target string) error
	Remove(path string) error

	IsDir(path string) (bool, error)
	IsEmpty(path string) (bool, error)
	Exists(path string) (bool, error)

	ReadLines(path string) ([]string, error)
	GetPathSegments(path string) []string
	GetExistingFileType(src, dest string) (ExistingType, error)
	RemoveEmptyDirectories(path string) error
	RemoveDirIfEmpty(path string) error
}
