/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"errors"
	"fmt"
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
)

const rootCmdName = "bestow"

const (
	FlagVerbose     string = "verbose"
	FlagDryRun      string = "dry-run"
	FlagConfigFile  string = "config-file"
	FlagProfile     string = "profile"
	FlagForce       string = "force"
	FlagAdopt       string = "adopt"
	FlagBackup      string = "backup"
	FlagInteractive string = "interactive"
)

var version = "dev"

var cfgFile string
var cfg *config.Config
var cfgFileFound bool

var (
	logHandler *log.Logger
	appLogger  *slog.Logger
)
var initConfigError error

// TODO: Add `config` subsommand (to override the init command)
var rootCmd = &cobra.Command{
	Use:           "bestow",
	Short:         rootCmdShort,
	Long:          rootCmdLong,
	Example:       rootCmdExamples,
	Version:       version,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if initConfigError != nil {
			return fmt.Errorf("failed to read configs: %w", initConfigError)
		}
		bindOperationalFlags(cmd, viper.GetViper())
		verbose, err := checkVerbose(cmd)
		if err != nil {
			return fmt.Errorf("failed to check flags: %w", err)
		}
		if verbose {
			logHandler.SetLevel(log.DebugLevel)
		}
		if !cfgFileFound {
			appLogger.Warn("config file not found; using default values", "hint", "run 'bestow init' to initialize the configs")
		}
		cfg, err = config.GetConfig(viper.GetViper(), appLogger)
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		var hintedError *engine.HintedError
		var conflictError *engine.ConflictError
		if errors.As(err, &hintedError) && hintedError.Hint != "" {
			appLogger.Error(hintedError.Error())
			appLogger.Info(fmt.Sprintf("Hint: %s", hintedError.Hint))
		} else if errors.As(err, &conflictError) {
			appLogger.Error(conflictError.Error())
			appLogger.Warn("[conflicts]")
			for _, conflict := range conflictError.Conflicts {
				appLogger.Warn("  ", "destination", conflict.Destination, "sources", strings.Join(conflict.Sources, ", "))
			}
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
	cobra.OnInitialize(initConfig)
	// disable showing `completion` in the available commands list while keeping the command available
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	// Hide the `help` subcommand from the subcommand list (only allow `-h/--help` flags)
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().Bool(FlagDryRun, false, "run the command without actually making the file system changes")
	rootCmd.PersistentFlags().Bool(FlagVerbose, false, "print verbose logs")
	rootCmd.PersistentFlags().StringVar(&cfgFile, FlagConfigFile, "", "provide custom config file")
	rootCmd.PersistentFlags().String(FlagProfile, "default", "profile to run the command")
}

func initConfig() {
	appLogger.Debug("initilizing config")
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
	if err := viper.ReadInConfig(); err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			cfgFileFound = false
		} else {
			initConfigError = err
		}
	} else {
		cfgFileFound = true
	}
}
