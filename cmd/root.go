/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/log"
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

var initConfigError error

var rootCmd = &cobra.Command{
	Use:           "bestow",
	Short:         RootCmdShort,
	Long:          RootCmdLong,
	Example:       RootCmdExamples,
	Version:       version,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if initConfigError != nil {
			return fmt.Errorf("failed to read configs: %w", initConfigError)
		}
		bindOperationalFlags(cmd, viper.GetViper())
		if err := checkVerbose(cmd); err != nil {
			return fmt.Errorf("failed to check flags: %w", err)
		}
		if !cfgFileFound {
			log.Warn("config file not found; using default values", "hint", "run 'bestow init' to create one")
		}
		var err error
		cfg, err = config.GetConfig(viper.GetViper())
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	log.SetLogger(log.NewCharmLogger())
	err := rootCmd.Execute()
	if err != nil {
		log.Error("Operation failed", "error", err)
		os.Exit(1)
	}
}

func init() {
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
	log.Debug("initilizing config")
	if cfgFile != "" {
		log.Debug("custom config file provided", "path", cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		configFilePath := filepath.Join(config.AppConfigHome(), constant.ConfigFile)
		log.Debug("no custom config file provided; using default", "path", configFilePath)
		viper.SetConfigFile(configFilePath)
	}
	if err := viper.ReadInConfig(); err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			cfgFileFound = false
		}
	} else {
		cfgFileFound = true
		initConfigError = err
	}

	viper.SetEnvPrefix(strings.ToUpper(rootCmdName))
	viper.AutomaticEnv()
}
