package ui

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

var (
	out      io.Writer
	useColor bool
)

func Init(w io.Writer, enableColor bool) {
	out = w
	useColor = enableColor
}

func print(prefix string, msg string, attrs ...color.Attribute) {
	outMsg := prefix + msg
	if useColor {
		color.New(attrs...).Fprintln(out, outMsg)
		return
	}
	fmt.Fprintln(out, outMsg)
}

func PrintPath(msg string) {
	print("", msg, color.FgGreen)
}
