/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:     "status [packages...]",
	Short:   statusShort,
	Long:    statusLong,
	Example: statusExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeStatus(viper.GetViper(), cmd, args)
	},
}

func init() {
	addOperationFlags(statusCmd.Flags())
	rootCmd.AddCommand(statusCmd)
}

func executeStatus(v *viper.Viper, cmd *cobra.Command, args []string) error {
	e, err := buildEngine(v, cmd, false)
	if err != nil {
		return err
	}
	status, err := e.Status(args)
	if err != nil {
		return err
	}
	appOutput.PrintStatus(status)
	return nil
}
