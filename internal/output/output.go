/*
All Rights Reversed (ɔ)
*/

package output

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ThisaruGuruge/bestow/internal/engine"
)

const actionStringLength = 7

var successStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Green)

var stepStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Cyan)

var skipStyle = lipgloss.NewStyle().
	Bold(true).
	Faint(true).
	Foreground(lipgloss.Green)

var warnStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Yellow)

var hintStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Magenta)

var actionStyle = lipgloss.NewStyle().Width(actionStringLength).Align(lipgloss.Right)

var conflictDestStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Blue)
var conflictSrcStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Yellow)

type Level int

const (
	Normal Level = iota
	Quiet
)

type Output struct {
	OutputLevel Level
}

func NewOutput(level Level) *Output {
	return &Output{
		OutputLevel: level,
	}
}

func (o *Output) SetLevel(level Level) {
	o.OutputLevel = level
}

func (o *Output) PrintAction(action engine.ActionEvent, label string) {
	var message string
	formattedAction := actionStyle.Render(action.Action)
	if label == "" {
		message = fmt.Sprintf("%s: %s", formattedAction, action.Msg)
	} else {
		message = fmt.Sprintf("%s %s: %s", label, formattedAction, action.Msg)
	}
	var text string
	switch action.EventType {
	case engine.EventSuccess:
		text = successStyle.Render(message)
	case engine.EventStep:
		text = stepStyle.Render(message)
	case engine.EventWarn:
		text = warnStyle.Render(message)
	case engine.EventSkip:
		text = skipStyle.Render(message)
	case engine.EventIgnore:
		return
	}
	lipgloss.Println(text)
}

func (o *Output) PrintSummary(summary *engine.ExecuteSummary) {
	var label string
	if summary.DryRun {
		label = "[dryrun]"
	}
	if o.OutputLevel != Quiet {
		for _, action := range summary.Actions {
			o.PrintAction(action, label)
		}
		o.printSummaryLine(summary.OperationSummary)
	}
}

func (o *Output) printSummaryLine(summary *engine.Summary) {
	summaryFields := 7
	parts := make([]string, 0, summaryFields)
	if summary.Stowed > 0 {
		parts = append(parts, fmt.Sprintf("stowed: %d", summary.Stowed))
	}
	if summary.Unstowed > 0 {
		parts = append(parts, fmt.Sprintf("unstowed: %d", summary.Unstowed))
	}
	if summary.Replaced > 0 {
		parts = append(parts, fmt.Sprintf("replaced: %d", summary.Replaced))
	}
	if summary.BackedUp > 0 {
		parts = append(parts, fmt.Sprintf("backed up: %d", summary.BackedUp))
	}
	if summary.Adopted > 0 {
		parts = append(parts, fmt.Sprintf("adopted: %d", summary.Adopted))
	}
	if summary.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("skipped: %d", summary.Skipped))
	}
	if summary.UpToDate > 0 {
		parts = append(parts, fmt.Sprintf("up-to-date: %d", summary.UpToDate))
	}
	if len(parts) == 0 {
		return
	}
	lipgloss.Println(strings.Join(parts, "   "))
}

func (o *Output) PrintHint(hint string) {
	message := "[hint] " + hint
	lipgloss.Fprintln(os.Stderr, hintStyle.Render(message))
}

func (o *Output) PrintConflict(conflicts []engine.DestinationConflict) {
	lipgloss.Fprintln(os.Stderr, conflictDestStyle.Render("conflicts:"))
	for _, conflict := range conflicts {
		lipgloss.Fprintln(os.Stderr, conflictDestStyle.Render(conflict.Destination))
		for _, source := range conflict.Sources {
			lipgloss.Fprintln(os.Stderr, conflictSrcStyle.Render("-", source))
		}
	}
}
