/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"errors"
	"fmt"
)

// HintedError represents an error with a hint to the user.
// The Hint field includes the hint for the error, and it should be handled in
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

// AggregatedError represents a collection of errors that needs to be grouped
// together for presenting.
type AggregatedError struct {
	Msg   string
	Items []error
}

func (e *AggregatedError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Items)
}

func (e *AggregatedError) Unwrap() []error {
	return e.Items
}

// ConflictError represents the errors where stow fails due to conflicts.
// The Conflicts field includes all conflicts found during the operation, which
// should be handled by the output.
type ConflictError struct {
	Op        string
	Conflicts []DestinationConflict
	Err       error
}

// DestinationConflict stores the conflicts (multiple sources competing for same destination) occurred while calculating destinations.
// Each destination will have the sources that compete for the destination
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

var (
	// ErrDestIsDir is returned when the destination file path is actually a directory
	ErrDestIsDir = errors.New("destination is a directory")
	// ErrPkgIsNotDir is returned when the provided package path is not a directory
	ErrPkgIsNotDir = errors.New("package is not a directory")
	// ErrRootIsNotPkg is returned when the root is provided as a package path
	ErrRootIsNotPkg = errors.New("root (.) is not a package")
	// ErrFileExists is returned when the provided destination is an existing file
	ErrFileExists = errors.New("file already exists")
	// ErrMultiFile is returned when multiple source files compete for the same destination
	ErrMultiFile = errors.New("multiple files competing for the same destination")
	// ErrUnsupportedAction is returned when the provided action is unsupported
	ErrUnsupportedAction = errors.New("unsupported action")
)
