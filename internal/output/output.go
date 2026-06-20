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

type Level int

const (
	Normal Level = iota
	Quiet
)

type Output struct {
	level        Level
	successStyle lipgloss.Style
	warnStyle    lipgloss.Style
	errStyle     lipgloss.Style
	stepStyle    lipgloss.Style
	skipStyle    lipgloss.Style
	hintStyle    lipgloss.Style
	undoStyle    lipgloss.Style
	actionStyle  lipgloss.Style
}

func NewOutput(l Level) *Output {
	hasDarkBg := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark := lipgloss.LightDark(hasDarkBg)
	return &Output{
		level:        l,
		successStyle: lipgloss.NewStyle().Bold(true).Foreground(lightDark(lipgloss.Green, lipgloss.BrightGreen)),
		warnStyle:    lipgloss.NewStyle().Bold(true).Foreground(lightDark(lipgloss.Yellow, lipgloss.BrightYellow)),
		errStyle:     lipgloss.NewStyle().Bold(true).Foreground(lightDark(lipgloss.Red, lipgloss.BrightRed)),
		stepStyle:    lipgloss.NewStyle().Bold(true).Foreground(lightDark(lipgloss.Cyan, lipgloss.BrightCyan)).Faint(true),
		skipStyle:    lipgloss.NewStyle().Italic(true).Foreground(lightDark(lipgloss.BrightBlack, lipgloss.White)).Faint(true),
		hintStyle:    lipgloss.NewStyle().Italic(true).Foreground(lightDark(lipgloss.Blue, lipgloss.BrightBlue)),
		undoStyle:    lipgloss.NewStyle().Bold(true).Foreground(lightDark(lipgloss.Magenta, lipgloss.BrightMagenta)),
		actionStyle:  lipgloss.NewStyle().Width(actionStringLength).Align(lipgloss.Right).Transform(strings.ToUpper),
	}
}

func (o *Output) SetLevel(level Level) {
	o.level = level
}

func (o *Output) PrintAction(action engine.ActionEvent, label string) {
	var message string
	formattedAction := o.actionStyle.Render(action.Action)
	if label == "" {
		message = fmt.Sprintf("%s: %s", formattedAction, action.Msg)
	} else {
		message = fmt.Sprintf("%s %s: %s", label, formattedAction, action.Msg)
	}
	var text string
	switch action.EventType {
	case engine.EventSuccess:
		text = o.successStyle.Render(message)
	case engine.EventStep:
		text = o.stepStyle.Render(message)
	case engine.EventWarn:
		text = o.warnStyle.Render(message)
	case engine.EventSkip:
		text = o.skipStyle.Render(message)
	case engine.EventUndo:
		text = o.undoStyle.Render(message)
	case engine.EventIgnore:
		return
	}
	lipgloss.Println(text)
}

func (o *Output) PrintSummary(summary *engine.ExecuteResult) {
	if summary == nil {
		return
	}
	var label string
	if summary.DryRun {
		label = "[dryrun]"
	}
	if o.level != Quiet {
		for _, action := range summary.Events {
			o.PrintAction(action, label)
		}
		o.printSummaryLine(summary.Summary)
	}
}

func (o *Output) printSummaryLine(summary *engine.OpsSummary) {
	if summary == nil {
		return
	}
	summaryFields := 8
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
	if summary.Reverted > 0 {
		parts = append(parts, fmt.Sprintf("reverted: %d", summary.Reverted))
	}
	if len(parts) == 0 {
		lipgloss.Println("no operations to execute")
		return
	}
	lipgloss.Println(strings.Join(parts, "   "))
}

func (o *Output) PrintHint(hint string) {
	message := "[hint] " + hint
	lipgloss.Fprintln(os.Stderr, o.hintStyle.Render(message))
}

func (o *Output) PrintConflict(conflicts []engine.DestinationConflict) {
	lipgloss.Fprintln(os.Stderr, o.errStyle.Render("conflicts:"))
	for _, conflict := range conflicts {
		lipgloss.Fprintln(os.Stderr, o.errStyle.Render(conflict.Destination))
		for _, source := range conflict.Sources {
			lipgloss.Fprintln(os.Stderr, o.warnStyle.Render("-", source))
		}
	}
}

func (o *Output) PrintAggregatedError(err *engine.AggregatedError) {
	lipgloss.Fprintln(os.Stderr, o.errStyle.Render(err.Msg))
	for _, item := range err.Items {
		lipgloss.Fprintln(os.Stderr, o.errStyle.Render(fmt.Sprintf("  %s", item.Error())))
	}
}

func (o *Output) PrintCommandError(err error) {
	lipgloss.Fprintln(os.Stderr, o.errStyle.Render(err.Error()))
}
