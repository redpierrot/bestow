/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"github.com/redpierrot/bestow/internal/file"
)

type State int

const (
	Unknown State = iota
	Unstowed
	Stowed
	Conflict
	numStates
)

type Status struct {
	counts [numStates]int
}

func (s *Status) Count(state State) int {
	if state < 0 || state >= numStates {
		return 0
	}
	return s.counts[state]
}

func (e *Engine) Status(args []string) (*Status, error) {
	candidates, err := e.findCandidates(args)
	if err != nil {
		return nil, err
	}
	status := &Status{}
	for _, candidate := range candidates {
		state, err := e.getState(candidate)
		if err != nil {
			return nil, err
		}
		status.counts[state]++
	}
	return status, nil
}

func (e *Engine) getState(candidate operationCandidate) (State, error) {
	exists, err := e.fileSystem.Exists(candidate.destination)
	if err != nil {
		return Unknown, err
	}
	if !exists {
		return Unstowed, nil
	}
	existingType, err := e.fileSystem.ExistingFileType(candidate.source, candidate.destination)
	if err != nil {
		return Unknown, err
	}
	switch existingType {
	case file.ExistingManagedSymlink:
		return Stowed, nil
	case file.ExistingForeignSymlink, file.ExistingDir, file.ExistingRegularFile:
		return Conflict, nil
	default:
		return Unknown, nil
	}
}
