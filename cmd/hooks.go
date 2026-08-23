/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"path/filepath"
	"strings"

	charmlog "github.com/charmbracelet/log"
	"github.com/redpierrot/bestow/internal/config"
	"github.com/redpierrot/bestow/internal/engine"
	"github.com/redpierrot/bestow/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupLogging(cmd *cobra.Command, app *App) error {
	verbose, err := boolFlag(cmd.Flags(), flagVerbose)
	if err != nil {
		return err
	}
	quiet, err := boolFlag(cmd.Flags(), flagQuiet)
	if err != nil {
		return err
	}
	if verbose {
		app.logHandler.SetLevel(charmlog.DebugLevel)
	}
	if quiet {
		app.logHandler.SetLevel(charmlog.ErrorLevel)
		app.out.SetLevel(output.Quiet)
	}
	return nil
}

func initConfig(app *App, configFile string) {
	app.logger.Debug("initializing config")
	if configFile != "" {
		app.logger.Debug("custom config file provided", "path", configFile)
		viper.SetConfigFile(configFile)
	} else {
		configFilePath := filepath.Join(config.AppConfigHome(), configFileName)
		app.logger.Debug("no custom config file provided; using default", "path", configFilePath)
		viper.SetConfigFile(configFilePath)
	}
	viper.SetEnvPrefix(strings.ToUpper(rootCmdName))
	viper.AutomaticEnv()
}

func loadProfile(v *viper.Viper, cmd *cobra.Command, app *App) (*config.Profile, error) {
	if err := v.ReadInConfig(); err != nil {
		return nil, &engine.HintedError{
			Op:   "read config",
			Err:  err,
			Hint: "run the command `bestow init` to initialize",
		}
	}
	// Profile flag is bound before the config is loaded so the viper configs does not pollute with provided profile keys
	if f := cmd.Flags().Lookup(flagProfile); f != nil {
		_ = v.BindPFlag(flagProfile, f)
	}
	profile, err := config.GetProfile(v, app.logger)
	if err != nil {
		return nil, err
	}
	if source, _ := stringFlag(cmd.Flags(), flagSource); source != "" {
		profile.Source = source
	}
	if destination, _ := stringFlag(cmd.Flags(), flagDestination); destination != "" {
		profile.Destination = destination
	}
	return profile, nil
}

func buildEngine(v *viper.Viper, cmd *cobra.Command, dryRun bool, app *App) (*engine.Engine, error) {
	profile, err := loadProfile(v, cmd, app)
	if err != nil {
		return nil, err
	}
	engineConfig := &engine.EngineConfig{
		Source:      profile.Source,
		Destination: profile.Destination,
		DryRun:      dryRun,
		ConfigHome:  config.AppConfigHome(),
	}
	return engine.NewEngine(engineConfig, app.logger)
}
