package clog

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/fatih/color"
)

var (
	out      io.Writer
	useColor bool
)

type Handler struct {
	writer io.Writer
	level  slog.Leveler
}

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

func Info(msg string) {
	print("[INFO]: ", msg, color.FgGreen)
}

func Warn(msg string) {
	print("[WARN]: ", msg, color.FgYellow)
}

func Error(err error) {
	print("[ERROR]: ", err.Error(), color.FgRed)
}

func Debug(err error) {
	print("[ERROR] ", err.Error(), color.FgHiMagenta)
}
