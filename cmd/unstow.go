/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/redpierrot/bestow/internal/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newUnstowCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unstow [packages...]",
		Short:   unstowShort,
		Long:    unstowLong,
		Example: unstowExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeUnstow(viper.GetViper(), cmd, args, app)
		},
	}
	addOperationFlags(cmd.Flags())
	return cmd
}

func executeUnstow(v *viper.Viper, cmd *cobra.Command, args []string, app *App) error {
	dryRun, err := boolFlag(cmd.Flags(), flagDryRun)
	if err != nil {
		return err
	}
	e, err := buildEngine(v, cmd, dryRun, app)
	if err != nil {
		return err
	}
	cmdCfg := engine.CommandConfig{
		Kind: engine.CommandUnstow,
		Args: args,
	}
	summary, err := e.Execute(cmd.Context(), &cmdCfg)
	app.out.PrintResult(summary)
	if err != nil {
		return err
	}
	return nil
}
