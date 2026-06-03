/*
All Rights Reversed (ɔ)
*/

package file

type ExistingType int

const (
	ExistingManagedSymlink ExistingType = iota
	ExistingForeignSymlink
	ExistingRegularFile
	ExistingDir
	ExistingUnknown
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
