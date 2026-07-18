/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"errors"
	"fmt"
	"testing"
)

func TestEngineError_HintedError(t *testing.T) {
	tests := []struct {
		name       string
		err        func() *HintedError
		wantUnwrap error
		wantString string
	}{
		{
			name: "HintedError - with op",
			err: func() *HintedError {
				return &HintedError{
					Op:   "read package bestow",
					Hint: "remove file bestow",
					Err:  errPkgIsNotDir,
				}
			},
			wantUnwrap: errPkgIsNotDir,
			wantString: fmt.Sprintf("read package bestow: %s", errPkgIsNotDir),
		},
		{
			name: "HintedError - no op",
			err: func() *HintedError {
				return &HintedError{
					Hint: "remove file bestow",
					Err:  errPkgIsNotDir,
				}
			},
			wantUnwrap: errPkgIsNotDir,
			wantString: errPkgIsNotDir.Error(),
		},
		{
			name: "HintedError - no op",
			err: func() *HintedError {
				return &HintedError{
					Hint: "remove file bestow",
					Err:  errPkgIsNotDir,
				}
			},
			wantUnwrap: errPkgIsNotDir,
			wantString: errPkgIsNotDir.Error(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.err()
			if !errors.Is(err.Unwrap(), tc.wantUnwrap) {
				t.Fatalf("got %v, want %v", err, tc.wantUnwrap)
			}
			if tc.wantString != err.Error() {
				t.Fatalf("got %s, want %s", err.Error(), tc.wantString)
			}
		})
	}
}

func TestEngineError_AggregatedError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		items      []error
		wantString string
		wantErr    []error
	}{
		{
			name:       "no items",
			message:    "found non multiple errors",
			items:      []error{errPkgIsNotDir, errInvalidPattern},
			wantString: "found non multiple errors: [package is not a directory invalid ignore pattern]",
			wantErr:    []error{errPkgIsNotDir, errInvalidPattern},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aggregatedErr := &AggregatedError{
				Msg:   tc.message,
				Items: tc.items,
			}
			if tc.wantString != aggregatedErr.Error() {
				t.Fatalf("got %s, want %s", aggregatedErr, tc.wantString)
			}
			if len(aggregatedErr.Unwrap()) != len(tc.wantErr) {
				t.Fatalf("got errors %d, want %d", len(aggregatedErr.Unwrap()), len(tc.wantErr))
			}
		})
	}
}

func TestEngineError_ConflictError(t *testing.T) {
	tests := []struct {
		name       string
		op         string
		wantString string
		wantErr    error
	}{
		{
			name:       "no op",
			wantString: "multiple files competing for the same destination",
			wantErr:    errMultiFile,
		},
		{
			name:       "with op",
			op:         "oh no",
			wantString: "oh no: multiple files competing for the same destination",
			wantErr:    errMultiFile,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conflictErr := &ConflictError{
				Op:        tc.op,
				Conflicts: []DestinationConflict{},
				Err:       errMultiFile,
			}
			if tc.wantString != conflictErr.Error() {
				t.Fatalf("got %s, want %s", conflictErr, tc.wantString)
			}
			if conflictErr.Unwrap() != tc.wantErr {
				t.Fatalf("got %v, want %v", conflictErr.Unwrap(), tc.wantErr)
			}
		})
	}
}
