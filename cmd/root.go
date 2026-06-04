/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/ThisaruGuruge/bestow/internal/output"
)

const rootCmdName = "bestow"

var version = "dev"

var cfgFile string

var (
	logHandler *log.Logger
	appLogger  *slog.Logger
	appOutput  *output.Output
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		var hintedError *engine.HintedError
		var conflictError *engine.ConflictError
		if errors.As(err, &hintedError) && hintedError.Hint != "" {
			appLogger.Error(hintedError.Error())
			appOutput.PrintHint(hintedError.Hint)
		} else if errors.As(err, &conflictError) {
			appLogger.Error(conflictError.Error())
			appOutput.PrintConflict(conflictError.Conflicts)
		} else {
			appLogger.Error(err.Error())
		}
		os.Exit(1)
	}
}

func init() {
	// Setting logger in the init method to avoid falling back to default logger.
	opts := log.Options{
		Level:           log.InfoLevel,
		ReportTimestamp: false,
	}
	logHandler = log.NewWithOptions(os.Stderr, opts)
	appLogger = slog.New(logHandler)

	appOutput = output.NewOutput(output.Normal)

	cobra.OnInitialize(initConfig)
	// Disable showing `completion` in the available commands list while keeping the command available
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	// Hide the `help` subcommand from the subcommand list (only allow `-h/--help` flags)
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().BoolP(flagDryRun, "n", false, "run the command without actually making the file system changes")
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "print verbose logs")
	rootCmd.PersistentFlags().BoolP(flagQuiet, "q", false, "quiet logs; only print the summary")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, flagConfigFile, "c", "", "provide custom config file")
	rootCmd.PersistentFlags().String(flagProfile, "default", "profile to run the command")

	rootCmd.MarkFlagsMutuallyExclusive(flagQuiet, flagVerbose)
}

func initConfig() {
	appLogger.Debug("initializing config")
	if cfgFile != "" {
		appLogger.Debug("custom config file provided", "path", cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		configFilePath := filepath.Join(config.AppConfigHome(), constant.ConfigFile)
		appLogger.Debug("no custom config file provided; using default", "path", configFilePath)
		viper.SetConfigFile(configFilePath)
	}
	viper.SetEnvPrefix(strings.ToUpper(rootCmdName))
	viper.AutomaticEnv()
}
