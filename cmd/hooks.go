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
)

const (
	envSource      = "BESTOW_SOURCE"
	envDestination = "BESTOW_DESTINATION"
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
		app.v.SetConfigFile(configFile)
	} else {
		configFilePath := filepath.Join(config.AppConfigHome(), configFileName)
		app.logger.Debug("no custom config file provided; using default", "path", configFilePath)
		app.v.SetConfigFile(configFilePath)
	}
	app.v.SetEnvPrefix(strings.ToUpper(rootCmdName))
	app.v.AutomaticEnv()
}

func loadProfile(cmd *cobra.Command, app *App) (*config.Profile, error) {
	if err := app.v.ReadInConfig(); err != nil {
		return nil, &engine.HintedError{
			Op:   "read config",
			Err:  err,
			Hint: "run the command `bestow init` to initialize",
		}
	}
	// Profile flag is bound before the config is loaded so the viper configs does not pollute with provided profile keys
	if f := cmd.Flags().Lookup(flagProfile); f != nil {
		_ = app.v.BindPFlag(flagProfile, f)
	}
	profileName := app.v.GetString(config.ProfileKey)
	if profileName == "" {
		profileName = config.DefaultProfile
	}
	profileConfig, err := config.ProfileConfig(app.v, profileName)
	if err != nil {
		return nil, err
	}
	_ = profileConfig.BindPFlag(flagSource, cmd.Flags().Lookup(flagSource))
	_ = profileConfig.BindPFlag(flagDestination, cmd.Flags().Lookup(flagDestination))
	_ = profileConfig.BindEnv(flagSource, envSource)
	_ = profileConfig.BindEnv(flagDestination, envDestination)

	return config.GetProfile(profileName, profileConfig, app.logger)
}

func buildEngine(cmd *cobra.Command, dryRun bool, app *App) (*engine.Engine, error) {
	profile, err := loadProfile(cmd, app)
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
