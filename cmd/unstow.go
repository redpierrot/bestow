/*
All Rights Reversed (ɔ)
*/

package cmd

import (
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
			return executeUnstow(cmd, viper.GetViper(), app, args)
		},
	}
	addOperationFlags(cmd.Flags())
	cmd.Flags().BoolP(flagKeepEmptyDirs, "k", false, "keep empty parent directories when unstow")
	return cmd
}

func executeUnstow(cmd *cobra.Command, v *viper.Viper, app *App, args []string) error {
	dryRun, err := boolFlag(cmd.Flags(), flagDryRun)
	if err != nil {
		return err
	}
	e, err := buildEngine(v, cmd, dryRun, app)
	if err != nil {
		return err
	}
	keepEmptyParents, err := boolFlag(cmd.Flags(), flagKeepEmptyDirs)
	if err != nil {
		return err
	}
	summary, err := e.Unstow(cmd.Context(), args, keepEmptyParents)
	app.out.PrintResult(summary)
	if err != nil {
		return err
	}
	return nil
}
