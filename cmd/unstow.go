/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/spf13/cobra"
)

var unstowCmd = &cobra.Command{
	Use:     "unstow [packages...]",
	Short:   unstowShort,
	Long:    unstowLong,
	Example: unstowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		appLogger.Debug("running stow command", "args", args)
		ctx := engine.CommandContext{
			Action: engine.ActionUnstow,
			Args:   args,
		}
		// TODO: Handle error?
		dryrun, _ := cmd.Flags().GetBool(FlagDryRun)
		engine, err := engine.NewEngine(cfg, dryrun, appLogger)
		if err != nil {
			return err
		}
		if err := engine.Execute(&ctx, &args); err != nil {
			return err
		}
		appLogger.Info("successfully unstowed the packages")
		return nil
	},
}

func init() {
	addOperationFlags(unstowCmd.Flags())
	rootCmd.AddCommand(unstowCmd)
}
