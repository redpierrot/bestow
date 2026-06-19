/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"fmt"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/spf13/cobra"
)

// These flags are defined seprately because semantically, the init flags are different than the common operation flags.
const (
	flagInitIgnoreList  string = "ignore-list"
	flagInitSource      string = "source"
	flagInitDestination string = "destination"
	flagInitForce       string = "force"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   initShort,
	Long:    initLong,
	Example: initExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		source, err := getStringFlag(cmd.Flags(), flagInitSource)
		if err != nil {
			return err
		}
		destination, err := getStringFlag(cmd.Flags(), flagInitDestination)
		if err != nil {
			return err
		}
		force, err := getBoolFlag(cmd.Flags(), flagInitForce)
		if err != nil {
			return err
		}
		engineCfg := engine.EngineConfig{
			Source:      source,
			Destination: destination,
			ConfigHome:  config.AppConfigHome(),
		}
		dryRun, err := getBoolFlag(cmd.Flags(), flagDryRun)
		if err != nil {
			return err
		}
		eng, err := engine.NewEngine(&engineCfg, dryRun, appLogger)
		if err != nil {
			return err
		}
		ignoreList, err := cmd.Flags().GetStringSlice(flagInitIgnoreList)
		if err != nil {
			return fmt.Errorf("parse flag %s: %w", flagInitIgnoreList, err)
		}
		ctx := engine.InitContext{
			Force:      force,
			IgnoreList: ignoreList,
		}
		summary, err := eng.Init(&ctx)
		if err != nil {
			return err
		}
		appOutput.PrintSummary(summary)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP(flagInitSource, "s", "", "source directory of the files for symlinks; written to 'config.yaml'")
	initCmd.MarkFlagRequired(flagInitSource)
	initCmd.Flags().StringP(flagInitDestination, "d", "", "destination for the symlinks; written to 'config.yaml'. (defaults to user home directory)")
	initCmd.Flags().StringSlice(flagInitIgnoreList, config.DefaultIgnoreList, "list of file/directory names bestow should ignore. This is the global set of values. For repo or package specific ignore lists, use specific .bestowignore files")
	initCmd.Flags().BoolP(flagInitForce, "f", false, "forcefully overwrite any existing config files for bestow")

	initCmd.Flags().SortFlags = false
	initCmd.PersistentFlags().SortFlags = false
}
