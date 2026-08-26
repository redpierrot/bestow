/*
All Rights Reversed (ɔ)
*/

// Package cmd is the main package that acts as the interface between the engine and the input.
package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/redpierrot/bestow/internal/engine"
	"github.com/redpierrot/bestow/internal/output"
)

const rootCmdName = "bestow"
const configFileName = "config.yaml"

var version = "dev"

type App struct {
	logHandler *charmlog.Logger
	logger     *slog.Logger
	out        *output.Output
	v          *viper.Viper
}

func newRootCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "bestow",
		Short:         rootCmdShort,
		Long:          rootCmdLong,
		Example:       rootCmdExamples,
		Version:       version,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configFile, err := stringFlag(cmd.Flags(), flagConfigFile)
			if err != nil {
				return err
			}
			initConfig(app, configFile)
			return setupLogging(cmd, app)
		},
	}
	initRootCommand(cmd)
	cmd.AddCommand(newStowCmd(app))
	cmd.AddCommand(newUnstowCmd(app))
	cmd.AddCommand(newStatusCmd(app))
	cmd.AddCommand(newInitCmd(app))

	return cmd
}

func Execute(ctx context.Context) {
	app := getApp()
	cmd := newRootCmd(app)
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(exitCodeFor(app.out, err))
	}
}

func exitCodeFor(out *output.Output, err error) int {
	if err != nil {
		var hintedError *engine.HintedError
		var conflictError *engine.ConflictError
		var aggregatedError *engine.AggregatedError
		if errors.As(err, &hintedError) && hintedError.Hint != "" {
			out.PrintCommandError(hintedError)
			out.PrintHint(hintedError.Hint)
		} else if errors.As(err, &conflictError) {
			out.PrintCommandError(conflictError)
			out.PrintConflict(conflictError.Conflicts)
		} else if errors.As(err, &aggregatedError) {
			out.PrintAggregatedError(aggregatedError)
		} else {
			out.PrintCommandError(err)
		}
		return 1
	}
	return 0
}

func initRootCommand(cmd *cobra.Command) {
	// Disable showing `completion` in the available commands list while keeping the command available
	cmd.CompletionOptions.HiddenDefaultCmd = true
	// Hide the `help` subcommand from the subcommand list (only allow `-h/--help` flags)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.PersistentFlags().BoolP(flagDryRun, "n", false, "run the command without actually making the file system changes")
	cmd.PersistentFlags().BoolP(flagVerbose, "v", false, "print verbose logs")
	cmd.PersistentFlags().BoolP(flagQuiet, "q", false, "quiet logs; only print the summary")
	cmd.PersistentFlags().String(flagConfigFile, "", "provide custom config file")
	cmd.PersistentFlags().String(flagProfile, "default", "profile to run the command")
	cmd.MarkFlagsMutuallyExclusive(flagQuiet, flagVerbose)
	cobra.EnableTraverseRunHooks = true
}

func getApp() *App {
	// Setting logger in the init method to avoid falling back to default logger.
	opts := charmlog.Options{
		Level:           charmlog.InfoLevel,
		ReportTimestamp: false,
	}
	logHandler := charmlog.NewWithOptions(os.Stderr, opts)
	logger := slog.New(logHandler)
	out := output.NewOutput(os.Stdout, os.Stderr)
	v := viper.New()
	return &App{
		logHandler: logHandler,
		logger:     logger,
		v:          v,
		out:        out,
	}
}
