/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"fmt"

	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/spf13/cobra"
)

var unstowCmd = &cobra.Command{
	Use:     "unstow [packages...]",
	Short:   unstowShort,
	Long:    unstowLong,
	Example: unstowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		appLogger.Debug("running unstow command", "args", args)
		dryrun, err := cmd.Flags().GetBool(FlagDryRun)
		if err != nil {
			return fmt.Errorf("parse %s: %w", FlagDryRun, err)
		}
		eng, err := engine.NewEngine(cfg, dryrun, appLogger)
		if err != nil {
			return err
		}
		ctx := engine.CommandContext{
			Action: engine.ActionUnstow,
			Args:   args,
		}
		if err := eng.Execute(&ctx); err != nil {
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
