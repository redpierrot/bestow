package file

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Does not work for Windows; we don't care!
func isWritable(path string) error {
	for path != string(os.PathSeparator) {
		if err := unix.Access(path, unix.W_OK); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			path = filepath.Dir(path)
			continue
		}
		break
	}
	return nil
}
