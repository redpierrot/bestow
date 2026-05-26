/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"fmt"

	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/spf13/cobra"
)

var stowCmd = &cobra.Command{
	Use:     "stow [packages...]",
	Short:   stowShort,
	Long:    stowLong,
	Example: stowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		appLogger.Debug("running stow command", "args", args)
		flagValues, err := getConflictFlags(cmd)
		if err != nil {
			return err
		}
		conflictResolution, err := conflictResolve(flagValues)
		if err != nil {
			return err
		}

		ctx := engine.CommandContext{
			Action:           engine.ActionStow,
			Args:             args,
			ConflictStrategy: conflictResolution,
		}
		//TODO: Handle error?
		dryrun, _ := cmd.Flags().GetBool(FlagDryRun)
		engine, err := engine.NewEngine(cfg, dryrun, appLogger)
		if err != nil {
			return err
		}
		if err := engine.Execute(&ctx, &args); err != nil {
			return err
		}
		appLogger.Info("successfully stowed the packages")
		return nil
	},
}

func init() {
	addOperationFlags(stowCmd.Flags())
	addConflictResolutionFlags(stowCmd.Flags())

	rootCmd.AddCommand(stowCmd)
}

func getConflictFlags(cmd *cobra.Command) ([]boolFlagValue, error) {
	var force, adopt, backup, interactive bool
	var err error
	force, err = cmd.Flags().GetBool(FlagForce)
	if err != nil {
		return nil, fmt.Errorf("failed to read the flag %s: %w", FlagForce, err)
	}
	adopt, err = cmd.Flags().GetBool(FlagAdopt)
	if err != nil {
		return nil, fmt.Errorf("failed to read the flag %s: %w", FlagForce, err)
	}
	backup, err = cmd.Flags().GetBool(FlagBackup)
	if err != nil {
		return nil, fmt.Errorf("failed to read the flag %s: %w", FlagForce, err)
	}
	interactive, err = cmd.Flags().GetBool(FlagInteractive)
	if err != nil {
		return nil, fmt.Errorf("failed to read the flag %s: %w", FlagForce, err)
	}

	flagValues := []boolFlagValue{
		{
			name:     FlagForce,
			value:    force,
			strategy: engine.ResolveForce,
		},
		{
			name:     FlagAdopt,
			value:    adopt,
			strategy: engine.ResolveAdopt,
		},
		{
			name:     FlagBackup,
			value:    backup,
			strategy: engine.ResolveBackup,
		},
		{
			name:     FlagInteractive,
			value:    interactive,
			strategy: engine.ResolveInteractive,
		},
	}
	return flagValues, nil
}
