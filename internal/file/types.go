/*
All Rights Reversed (ɔ)
*/

package file

// ExistingType defines the different types of existing files
type ExistingType int

const (
	// ExistingUnknown provided source and destination is an unknown type
	ExistingUnknown ExistingType = iota
	// ExistingManagedSymlink the path is a symlink managed (by bestow); i.e.: the destination and the source are the same
	ExistingManagedSymlink
	// ExistingForeignSymlink the path is a symlink, not managed by bestow
	ExistingForeignSymlink
	// ExistingRegularFile the path is a regular file
	ExistingRegularFile
	// ExistingDir the path is a directory
	ExistingDir
)

func (e ExistingType) String() string {
	switch e {
	case ExistingManagedSymlink:
		return "managed symlink"
	case ExistingForeignSymlink:
		return "foreign symlink"
	case ExistingRegularFile:
		return "regular file"
	case ExistingDir:
		return "directory"
	default:
		return "unknown"
	}
}
