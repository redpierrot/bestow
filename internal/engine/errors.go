/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"strings"
)

type EngineError struct {
	Message   string
	Cause     error
	Hint      string
	Conflicts []DestinationConflict
}

type DestinationConflict struct {
	destination string
	sources     []string
}

func (e *EngineError) Error() string {
	message := e.Message
	if e.Cause != nil {
		message = fmt.Sprintf("%s: %v", message, e.Cause)
	}
	for _, c := range e.Conflicts {
		message = fmt.Sprintf("%s\n  %s <- [%s]", message, c.destination, strings.Join(c.sources, ", "))
	}
	if e.Hint != "" {
		message = fmt.Sprintf("%s: [Hint] %s", message, e.Hint)
	}
	return message
}

func (e *EngineError) Unwrap() error { return e.Cause }
