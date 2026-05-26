/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/ThisaruGuruge/bestow/internal/output"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const (
	flagInitIgnoreList  string = "ignore-list"
	flagInitSource      string = "source"
	flagInitDestination string = "destination"
	flagInitForce       string = "force"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   InitShort,
	Long:    initLong,
	Example: initExamples,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := checkVerbose(cmd)
		if err != nil {
			return err
		}
		if verbose {
			logHandler.SetLevel(log.DebugLevel)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString(flagInitSource)
		destination, _ := cmd.Flags().GetString(flagInitDestination)
		force, _ := cmd.Flags().GetBool(flagInitForce)
		config := config.Config{
			Source:      source,
			Destination: destination,
		}
		// TODO: Handle error?
		dryrun, _ := cmd.Flags().GetBool(FlagDryRun)
		eng, err := engine.NewEngine(&config, dryrun, appLogger)
		if err != nil {
			return err
		}
		ignoreList, err := cmd.Flags().GetStringSlice(flagInitIgnoreList)
		ctx := engine.CommandContext{
			Action:     engine.ActionInit,
			Force:      force,
			IgnoreList: ignoreList,
		}
		if err := eng.Execute(&ctx, &[]string{}); err != nil {
			return err
		}
		output.Success("successfully initialized bestow")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringSlice(flagInitIgnoreList, config.DefaultIgnoreList, "list of file/directory names bestow should ignore. This is the global set of values. For repo or package specific ignore lists, use specific .bestowignore files")
	initCmd.Flags().StringP(flagInitSource, "s", "", "source of dotfiles for symlinks; written to 'config.yaml'")
	initCmd.MarkFlagRequired(flagInitSource)
	initCmd.Flags().StringP(flagInitDestination, "d", "", "destination for the dotfiles symlinks; written to 'config.yaml'. (defaults to '$HOME')")
	initCmd.Flags().BoolP(flagInitForce, "f", false, "forcefully overwrite any existing config files for bestow")
}
