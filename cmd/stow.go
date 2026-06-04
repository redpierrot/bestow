/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/spf13/cobra"
)

var stowCmd = &cobra.Command{
	Use:     "stow [packages...]",
	Short:   stowShort,
	Long:    stowLong,
	Example: stowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		appLogger.Debug("running stow command", "args", args)
		var force, adopt, backup bool
		force, err = getBoolFlag(cmd.Flags(), flagForce)
		if err != nil {
			return err
		}
		adopt, err = getBoolFlag(cmd.Flags(), flagAdopt)
		if err != nil {
			return err
		}
		backup, err = getBoolFlag(cmd.Flags(), flagBackup)
		if err != nil {
			return err
		}
		var strategy engine.ResolveStrategy
		if force {
			strategy = engine.ResolveForce
		}
		if adopt {
			strategy = engine.ResolveAdopt
		}
		if backup {
			strategy = engine.ResolveBackup
		}

		dryrun, err := getBoolFlag(cmd.Flags(), flagDryRun)
		if err != nil {
			return err
		}
		ctx := engine.CommandContext{
			Action:           engine.ActionStow,
			Args:             args,
			ConflictStrategy: strategy,
		}
		eng, err := engine.NewEngine(cfg, dryrun, appLogger)
		if err != nil {
			return err
		}
		summary, err := eng.Execute(&ctx)
		if err != nil {
			return err
		}
		appOutput.PrintSummary(summary)
		return nil
	},
}

func init() {
	addOperationFlags(stowCmd.Flags())
	addConflictResolutionFlags(stowCmd)

	rootCmd.AddCommand(stowCmd)
}
