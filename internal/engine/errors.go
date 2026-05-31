/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"errors"
)

// HintedError represents an error with a hint to the user.
// The Hint filed includes the hint for the error, and it should be handled in
// the outputs where the error is surfaced.
type HintedError struct {
	Op   string
	Hint string
	Err  error
}

func (e *HintedError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return e.Op + ": " + e.Err.Error()
}

func (e *HintedError) Unwrap() error {
	return e.Err
}

// ConflictError represents the errors where stow fails due to conflicts.
// The Conflicts filed includes all conflicts found during the operation, which
// should be handled by the output.
type ConflictError struct {
	Op        string
	Conflicts []DestinationConflict
	Err       error
}

type DestinationConflict struct {
	Destination string
	Sources     []string
}

func (e *ConflictError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return e.Op + ": " + e.Err.Error()
}

func (e *ConflictError) Unwrap() error { return e.Err }

var ErrDestIsDir = errors.New("destination is a directory")
var ErrPkgIsNotDir = errors.New("package is not a directory")
var ErrRootIsNotPkg = errors.New("root (.) is not a package")
var ErrFileExists = errors.New("file already exists")
var ErrMultiFile = errors.New("multiple files competing for the same destination")
var ErrUnsupportedAction = errors.New("unsupported action")
