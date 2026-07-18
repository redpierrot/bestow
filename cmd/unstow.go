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

var unstowCmd = &cobra.Command{
	Use:     "unstow [packages...]",
	Short:   unstowShort,
	Long:    unstowLong,
	Example: unstowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(*viper.GetViper(), cmd)
		if err != nil {
			return err
		}
		appLogger.Debug("running unstow command", "args", args)
		dryRun, err := boolFlag(cmd.Flags(), flagDryRun)
		if err != nil {
			return err
		}
		engineCfg := engine.EngineConfig{
			Source:      cfg.Source,
			Destination: cfg.Destination,
			ConfigHome:  config.AppConfigHome(),
		}
		eng, err := engine.NewEngine(&engineCfg, dryRun, appLogger)
		if err != nil {
			return err
		}
		cmdCfg := engine.CommandConfig{
			Action: engine.CommandUnstow,
			Args:   args,
		}
		summary, err := eng.Execute(cmd.Context(), &cmdCfg)
		appOutput.PrintResult(summary)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addOperationFlags(unstowCmd.Flags())
	rootCmd.AddCommand(unstowCmd)
}
