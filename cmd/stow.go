/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/config"
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
		engineCtx := engine.EngineContext{
			Source:      cfg.Source,
			Destination: cfg.Destination,
			ConfigHome:  config.AppConfigHome(),
		}
		eng, err := engine.NewEngine(&engineCtx, dryrun, appLogger)
		if err != nil {
			return err
		}
		cmdCtx := engine.CommandContext{
			Action:           engine.CommandStow,
			Args:             args,
			ConflictStrategy: strategy,
		}
		if err != nil {
			return err
		}
		summary, err := eng.Execute(&cmdCtx)
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
