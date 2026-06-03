/*
All Rights Reversed (ɔ)
*/

package output

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Green)

var stepStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Cyan)

var warnStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Yellow)

var hintStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Magenta)

func PrintAction(label, action, msg string, t Type) {
	var message string
	if label == "" {
		message = fmt.Sprintf("%s %s", action, msg)
	} else {
		message = fmt.Sprintf("%s %s %s", label, action, msg)
	}
	var text string
	switch t {
	case TypeSuccess:
		text = successStyle.Render(message)
	case TypeStep:
		text = stepStyle.Render(message)
	case TypeWarn:
		text = warnStyle.Render(message)
	}
	lipgloss.Println(text)
}

func PrintSummary(summary *Summary) {
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
	if summary.Backed > 0 {
		parts = append(parts, fmt.Sprintf("backed up: %d", summary.Backed))
	}
	if summary.Adopted > 0 {
		parts = append(parts, fmt.Sprintf("adopted: %d", summary.Adopted))
	}
	if summary.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("skipped: %d", summary.Skipped))
	}
	if summary.UpToDate > 0 {
		parts = append(parts, fmt.Sprintf("up to date: %d", summary.UpToDate))
	}
	if len(parts) == 0 {
		return
	}
	lipgloss.Println(strings.Join(parts, "   "))
}

func PrintHint(hint string) {
	message := "[hint] " + hint
	lipgloss.Fprintln(os.Stderr, hintStyle.Render(message))
}
