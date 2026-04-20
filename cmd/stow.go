/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/ThisaruGuruge/bestow/internal/log"
	"github.com/spf13/cobra"
)

// stowCmd represents the stow command
var stowCmd = &cobra.Command{
	Use:     "stow [packages...]",
	Short:   StowShort,
	Long:    StowLong,
	Example: StowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug("running stow command", "args", args)
		flagValues := getConflictFlags(cmd)
		conflictResolution, err := conflictResolve(flagValues)
		if err != nil {
			return err
		}

		ctx := engine.CommandContext{
			Action:   engine.ActionStow,
			Args:     args,
			Conflict: conflictResolution,
		}
		engine, err := engine.NewEngine(&ctx, cfg)
		if err != nil {
			return err
		}
		if err := engine.Execute(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addOperationFlags(stowCmd.Flags())
	addConflictResolutionFlags(stowCmd.Flags())

	rootCmd.AddCommand(stowCmd)
}

func getConflictFlags(cmd *cobra.Command) []boolFlagValue {
	// TODO: Handle errors; maybe we can ignore these
	force, _ := cmd.Flags().GetBool(FlagForce)
	adopt, _ := cmd.Flags().GetBool(FlagAdopt)
	backup, _ := cmd.Flags().GetBool(FlagBackup)
	interactive, _ := cmd.Flags().GetBool(FlagInteractive)

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
	return flagValues
}
