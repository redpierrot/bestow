package file

type FileSystem interface {
	ListAllFilesInDir(parent string, dirName string) ([]string, error)
	ListFiles(path string) ([]string, error)
	ListAllDirectories(source string) ([]string, error)
	CreateFile(fileName string, path string, data string) error
	CreateDir(path string) error
	IsDir(path string) (bool, error)
	IsEmpty(path string) (bool, error)
	Exists(path string) (bool, error)
	ReadLines(path string) ([]string, error)
	GetPathSegments(path string) []string
	GetExistingFileType(src, dest string) (ExistingType, error)
	Link(src, dest string) error
	Remove(path string) error
	RemoveEmptyDirectories(path string) error
	RemoveDirIfEmpty(path string) error
	Backup(path string) error
	Copy(src, dest string) error
}
