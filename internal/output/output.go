package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var uiRenderer = lipgloss.NewRenderer(os.Stdout)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("2")).
	Renderer(uiRenderer)

func Success(message string, args ...any) {
	text := fmt.Sprintf(message, args...)
	fmt.Println(successStyle.Render(text))
}
