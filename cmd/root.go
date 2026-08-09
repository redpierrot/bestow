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
	"path/filepath"
	"strings"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/redpierrot/bestow/internal/config"
	"github.com/redpierrot/bestow/internal/engine"
	"github.com/redpierrot/bestow/internal/output"
)

const rootCmdName = "bestow"
const configFileName = "config.yaml"

var version = "dev"

var (
	configFile  string
	charmLogger *charmlog.Logger
	appLogger   *slog.Logger
	appOutput   *output.Output
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "bestow",
		Short:         rootCmdShort,
		Long:          rootCmdLong,
		Example:       rootCmdExamples,
		Version:       version,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return setupLogging(cmd)
		},
	}
	initRootCommand(cmd)
	cmd.AddCommand(newStowCmd())
	cmd.AddCommand(newUnstowCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newInitCmd())

	return cmd
}

func Execute(ctx context.Context) {
	cmd := newRootCmd()
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(exitCodeFor(appOutput, err))
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
	// Setting logger in the init method to avoid falling back to default logger.
	opts := charmlog.Options{
		Level:           charmlog.InfoLevel,
		ReportTimestamp: false,
	}
	charmLogger = charmlog.NewWithOptions(os.Stderr, opts)
	appLogger = slog.New(charmLogger)

	appOutput = output.NewOutput(os.Stdout, os.Stderr)

	cobra.OnInitialize(initConfig)
	// Disable showing `completion` in the available commands list while keeping the command available
	cmd.CompletionOptions.HiddenDefaultCmd = true
	// Hide the `help` subcommand from the subcommand list (only allow `-h/--help` flags)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	cmd.PersistentFlags().BoolP(flagDryRun, "n", false, "run the command without actually making the file system changes")
	cmd.PersistentFlags().BoolP(flagVerbose, "v", false, "print verbose logs")
	cmd.PersistentFlags().BoolP(flagQuiet, "q", false, "quiet logs; only print the summary")
	cmd.PersistentFlags().StringVar(&configFile, flagConfigFile, "", "provide custom config file")
	cmd.PersistentFlags().String(flagProfile, "default", "profile to run the command")

	cmd.MarkFlagsMutuallyExclusive(flagQuiet, flagVerbose)
	cobra.EnableTraverseRunHooks = true
}

func initConfig() {
	appLogger.Debug("initializing config")
	if configFile != "" {
		appLogger.Debug("custom config file provided", "path", configFile)
		viper.SetConfigFile(configFile)
	} else {
		configFilePath := filepath.Join(config.AppConfigHome(), configFileName)
		appLogger.Debug("no custom config file provided; using default", "path", configFilePath)
		viper.SetConfigFile(configFilePath)
	}
	viper.SetEnvPrefix(strings.ToUpper(rootCmdName))
	viper.AutomaticEnv()
}
