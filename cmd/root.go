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
	FlagVerbose    string = "verbose"
	FlagDryRun     string = "dry-run"
	FlagConfigFile string = "config-file"
	FlagProfile    string = "profile"
	FlagForce      string = "force"
	FlagAdopt      string = "adopt"
)

var version = "dev"

var cfgFile string
var cfg *config.Config
var cfgFileFound bool

var initConfigError error

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bestow",
	Short: "bestow (BEtter STOW) is a modern dotfiles manager for Linux and MacOS",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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
