/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/redpierrot/bestow/internal/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type stowParams struct {
	dryRun   bool
	strategy engine.ResolveStrategy
}

func newStowCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stow [packages...]",
		Short:   stowShort,
		Long:    stowLong,
		Example: stowExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeStow(cmd, viper.GetViper(), app, args)
		},
	}
	addOperationFlags(cmd.Flags())
	addConflictResolutionFlags(cmd)
	return cmd
}

func executeStow(cmd *cobra.Command, v *viper.Viper, app *App, args []string) error {
	params, err := parseStowParams(cmd, args)
	if err != nil {
		return err
	}
	e, err := buildEngine(v, cmd, params.dryRun, app)
	if err != nil {
		return err
	}
	summary, err := e.Stow(cmd.Context(), args, params.strategy)

	app.out.PrintResult(summary)
	if err != nil {
		return err
	}
	return nil
}

func parseStowParams(cmd *cobra.Command, args []string) (*stowParams, error) {
	var force, adopt, backup bool
	force, err := boolFlag(cmd.Flags(), flagForce)
	if err != nil {
		return nil, err
	}
	adopt, err = boolFlag(cmd.Flags(), flagAdopt)
	if err != nil {
		return nil, err
	}
	backup, err = boolFlag(cmd.Flags(), flagBackup)
	if err != nil {
		return nil, err
	}
	var strategy = resolveStrategy(force, adopt, backup)
	dryRun, err := boolFlag(cmd.Flags(), flagDryRun)
	if err != nil {
		return nil, err
	}
	return &stowParams{
		dryRun:   dryRun,
		strategy: strategy,
	}, nil
}

func resolveStrategy(force, adopt, backup bool) engine.ResolveStrategy {
	switch {
	case force:
		return engine.ResolveForce
	case adopt:
		return engine.ResolveAdopt
	case backup:
		return engine.ResolveBackup
	default:
		return engine.ResolveSkip
	}
}
