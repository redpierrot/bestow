/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newStatusCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status [packages...]",
		Short:   statusShort,
		Long:    statusLong,
		Example: statusExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeStatus(viper.GetViper(), cmd, args, app)
		},
	}
	addOperationFlags(cmd.Flags())
	return cmd
}

func executeStatus(v *viper.Viper, cmd *cobra.Command, args []string, app *App) error {
	e, err := buildEngine(v, cmd, false, app)
	if err != nil {
		return err
	}
	status, err := e.Status(args)
	if err != nil {
		return err
	}
	app.out.PrintStatus(status)
	return nil
}
