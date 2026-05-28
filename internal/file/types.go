/*
All Rights Reversed (ɔ)
*/

package file

type ExistingType string

const (
	ExistingManagedSymlink ExistingType = "managed_symlink"
	ExistingForeignSymlink ExistingType = "foreign_symlink"
	ExistingRegularFile    ExistingType = "regular_file"
	ExistingDir            ExistingType = "directory"
	ExistingUnknown        ExistingType = "unknown_type"
)
