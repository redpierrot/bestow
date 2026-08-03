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

var statusCmd = &cobra.Command{
	Use:     "status [packages...]",
	Short:   statusShort,
	Long:    statusLong,
	Example: statusExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(*viper.GetViper(), cmd)
		if err != nil {
			return err
		}
		appLogger.Debug("running status command", "args", args)

		engineCfg := engine.EngineConfig{
			Source:      cfg.Source,
			Destination: cfg.Destination,
			ConfigHome:  config.AppConfigHome(),
		}
		eng, err := engine.NewEngine(&engineCfg, true, appLogger)
		if err != nil {
			return err
		}
		status, err := eng.Status(args)
		if err != nil {
			return err
		}
		appOutput.PrintStatus(status)
		return nil
	},
}

func init() {
	addOperationFlags(statusCmd.Flags())

	rootCmd.AddCommand(statusCmd)
}
