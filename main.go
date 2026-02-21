package main

import (
	"os"

	cmd "github.com/ThisaruGuruge/bestow/cmd/bestow"
	"github.com/ThisaruGuruge/bestow/internal/clog"
	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/terminal"
	"github.com/ThisaruGuruge/bestow/internal/ui"
	"github.com/fatih/color"
)

func main() {
	color.NoColor = false
	stdoutTty := terminal.IsTerminal(os.Stdout)
	stderrTty := terminal.IsTerminal(os.Stderr)

	ui.Init(os.Stdout, stdoutTty)
	clog.Init(os.Stderr, stderrTty)
	config := config.GetConfigs()
	cmd.Execute(*config)
}
