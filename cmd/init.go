/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/redpierrot/bestow/internal/config"
	"github.com/redpierrot/bestow/internal/engine"
	"github.com/spf13/cobra"
)

// These flags are defined seprately because semantically, the init flags are different than the common operation flags.
const (
	flagInitIgnoreList  = "ignore-list"
	flagInitSource      = "source"
	flagInitDestination = "destination"
	flagInitForce       = "force"
)

func newInitCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Short:   initShort,
		Long:    initLong,
		Example: initExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeInit(cmd, app)
		},
	}
	cmd.Flags().StringP(flagInitSource, "s", "", "source directory of the files for symlinks; written to 'config.yaml'")
	_ = cmd.MarkFlagRequired(flagInitSource)
	cmd.Flags().StringP(flagInitDestination, "d", "", "destination for the symlinks; written to 'config.yaml'. (defaults to user home directory)")
	cmd.Flags().StringSlice(flagInitIgnoreList, config.DefaultIgnoreList, "list of file/directory names bestow should ignore. This is the global set of values. For repo or package specific ignore lists, use specific .bestowignore files")
	cmd.Flags().BoolP(flagInitForce, "f", false, "forcefully overwrite any existing config files for bestow")

	cmd.Flags().SortFlags = false
	cmd.PersistentFlags().SortFlags = false

	// To avoid showing the long default ignore list on help text
	cmd.Flags().Lookup(flagInitIgnoreList).DefValue = "common dotfile ignore patterns"
	return cmd
}

func executeInit(cmd *cobra.Command, app *App) error {
	source, err := stringFlag(cmd.Flags(), flagInitSource)
	if err != nil {
		return err
	}
	if source == "" {
		app.logger.Warn("no source provided, using current directory")
		source, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	destination, err := stringFlag(cmd.Flags(), flagInitDestination)
	if err != nil {
		return err
	}
	force, err := boolFlag(cmd.Flags(), flagInitForce)
	if err != nil {
		return err
	}
	dryRun, err := boolFlag(cmd.Flags(), flagDryRun)
	if err != nil {
		return err
	}
	engineCfg := engine.EngineConfig{
		Source:      source,
		Destination: destination,
		ConfigHome:  config.AppConfigHome(),
		DryRun:      dryRun,
	}
	eng, err := engine.NewEngine(&engineCfg, app.logger)
	if err != nil {
		return err
	}
	ignoreList, err := cmd.Flags().GetStringSlice(flagInitIgnoreList)
	if err != nil {
		return fmt.Errorf("parse flag %s: %w", flagInitIgnoreList, err)
	}
	cfg := engine.InitConfig{
		Force:      force,
		IgnoreList: ignoreList,
		ConfigFile: configFileName,
	}
	summary, err := eng.Init(&cfg)
	if err != nil {
		return err
	}
	app.out.PrintResult(summary)
	return nil

}
