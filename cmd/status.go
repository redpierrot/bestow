/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/spf13/cobra"
)

func newStatusCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status [packages...]",
		Short:   statusShort,
		Long:    statusLong,
		Example: statusExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeStatus(cmd, app, args)
		},
	}
	addOperationFlags(cmd.Flags())
	return cmd
}

func executeStatus(cmd *cobra.Command, app *App, args []string) error {
	e, err := buildEngine(cmd, false, app)
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
