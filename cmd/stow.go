/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/redpierrot/bestow/internal/config"
	"github.com/redpierrot/bestow/internal/engine"
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
		force, err = boolFlag(cmd.Flags(), flagForce)
		if err != nil {
			return err
		}
		adopt, err = boolFlag(cmd.Flags(), flagAdopt)
		if err != nil {
			return err
		}
		backup, err = boolFlag(cmd.Flags(), flagBackup)
		if err != nil {
			return err
		}
		var strategy = engine.ResolveSkip
		if force {
			strategy = engine.ResolveForce
		}
		if adopt {
			strategy = engine.ResolveAdopt
		}
		if backup {
			strategy = engine.ResolveBackup
		}

		dryRun, err := boolFlag(cmd.Flags(), flagDryRun)
		if err != nil {
			return err
		}
		engineCfg := engine.EngineConfig{
			Source:      cfg.Source,
			Destination: cfg.Destination,
			ConfigHome:  config.AppConfigHome(),
		}
		eng, err := engine.NewEngine(&engineCfg, dryRun, appLogger)
		if err != nil {
			return err
		}
		cmdCfg := engine.CommandConfig{
			Action:           engine.CommandStow,
			Args:             args,
			ConflictStrategy: strategy,
		}
		summary, err := eng.Execute(cmd.Context(), &cmdCfg)

		// TODO: Decide on "What to Print" when error occurrs
		appOutput.PrintResult(summary)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addOperationFlags(stowCmd.Flags())
	addConflictResolutionFlags(stowCmd)

	rootCmd.AddCommand(stowCmd)
}
