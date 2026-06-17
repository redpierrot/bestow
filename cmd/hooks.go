/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"errors"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/ThisaruGuruge/bestow/internal/output"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrIncompatibleFlags = errors.New("mutually exclusive flags")

func setupLogging(cmd *cobra.Command) error {
	verbose, err := getBoolFlag(cmd.Flags(), flagVerbose)
	if err != nil {
		return err
	}
	quiet, err := getBoolFlag(cmd.Flags(), flagQuiet)
	if err != nil {
		return err
	}
	if verbose {
		logHandler.SetLevel(log.DebugLevel)
	}
	if quiet {
		appOutput.SetLevel(output.Quiet)
	}
	return nil
}

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		return nil, &engine.HintedError{
			Op:   "read config",
			Err:  err,
			Hint: "run the command `bestow init` to initialize",
		}
	}
	bindOperationalFlags(cmd, viper.GetViper())
	cfg, err := config.NewConfig(viper.GetViper(), appLogger)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
