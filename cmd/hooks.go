/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	charmlog "github.com/charmbracelet/log"
	"github.com/redpierrot/bestow/internal/config"
	"github.com/redpierrot/bestow/internal/engine"
	"github.com/redpierrot/bestow/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupLogging(cmd *cobra.Command) error {
	verbose, err := boolFlag(cmd.Flags(), flagVerbose)
	if err != nil {
		return err
	}
	quiet, err := boolFlag(cmd.Flags(), flagQuiet)
	if err != nil {
		return err
	}
	if verbose {
		charmLogger.SetLevel(charmlog.DebugLevel)
	}
	if quiet {
		charmLogger.SetLevel(charmlog.ErrorLevel)
		appOutput.SetLevel(output.Quiet)
	}
	return nil
}

func loadConfig(v viper.Viper, cmd *cobra.Command) (*config.Config, error) {
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
	cfg, err := config.NewConfig(&v, appLogger)
	if err != nil {
		return nil, err
	}
	if source, _ := stringFlag(cmd.Flags(), flagSource); source != "" {
		cfg.Source = source
	}
	if destination, _ := stringFlag(cmd.Flags(), flagDestination); destination != "" {
		cfg.Destination = destination
	}
	return cfg, nil
}
