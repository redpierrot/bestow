/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/redpierrot/bestow/internal/config"
	"github.com/redpierrot/bestow/internal/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type stowParams struct {
	source      string
	destination string
	dryRun      bool
	strategy    engine.ResolveStrategy
	packages    []string
}

var stowCmd = &cobra.Command{
	Use:     "stow [packages...]",
	Short:   stowShort,
	Long:    stowLong,
	Example: stowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeStow(viper.GetViper(), cmd, args)
	},
}

func init() {
	addOperationFlags(stowCmd.Flags())
	addConflictResolutionFlags(stowCmd)
	rootCmd.AddCommand(stowCmd)
}

func executeStow(v *viper.Viper, cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(v, cmd)
	if err != nil {
		return err
	}
	appLogger.Debug("running stow command", "args", args)
	params, err := parseStowParams(cfg, cmd, args)
	if err != nil {
		return err
	}
	e, err := buildEngine(v, cmd, params.dryRun)
	if err != nil {
		return err
	}
	cmdCfg := engine.CommandConfig{
		Kind:            engine.CommandStow,
		Args:            params.packages,
		ResolveStrategy: params.strategy,
	}
	summary, err := e.Execute(cmd.Context(), &cmdCfg)

	appOutput.PrintResult(summary)
	if err != nil {
		return err
	}
	return nil
}

func parseStowParams(cfg *config.Config, cmd *cobra.Command, args []string) (*stowParams, error) {
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
		source:      cfg.Source,
		destination: cfg.Destination,
		dryRun:      dryRun,
		strategy:    strategy,
		packages:    args,
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
