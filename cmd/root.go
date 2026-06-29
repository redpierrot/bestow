/*
All Rights Reversed (ɔ)
*/

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

// TODO: Add `config` subcommand (to override the init command)
var rootCmd = &cobra.Command{
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

func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		var hintedError *engine.HintedError
		var conflictError *engine.ConflictError
		var aggregatedError *engine.AggregatedError
		if errors.As(err, &hintedError) && hintedError.Hint != "" {
			appOutput.PrintCommandError(hintedError)
			appOutput.PrintHint(hintedError.Hint)
		} else if errors.As(err, &conflictError) {
			appOutput.PrintCommandError(conflictError)
			appOutput.PrintConflict(conflictError.Conflicts)
		} else if errors.As(err, &aggregatedError) {
			appOutput.PrintAggregatedError(aggregatedError)
		} else {
			appOutput.PrintCommandError(err)
		}
		os.Exit(1)
	}
}

func init() {
	// Setting logger in the init method to avoid falling back to default logger.
	opts := charmlog.Options{
		Level:           charmlog.InfoLevel,
		ReportTimestamp: false,
	}
	charmLogger = charmlog.NewWithOptions(os.Stderr, opts)
	appLogger = slog.New(charmLogger)

	appOutput = output.NewOutput(output.Normal)

	cobra.OnInitialize(initConfig)
	// Disable showing `completion` in the available commands list while keeping the command available
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	// Hide the `help` subcommand from the subcommand list (only allow `-h/--help` flags)
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().BoolP(flagDryRun, "n", false, "run the command without actually making the file system changes")
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "print verbose logs")
	rootCmd.PersistentFlags().BoolP(flagQuiet, "q", false, "quiet logs; only print the summary")
	rootCmd.PersistentFlags().StringVar(&configFile, flagConfigFile, "", "provide custom config file")
	rootCmd.PersistentFlags().String(flagProfile, "default", "profile to run the command")

	rootCmd.MarkFlagsMutuallyExclusive(flagQuiet, flagVerbose)
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
