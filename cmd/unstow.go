/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/spf13/cobra"
)

var unstowCmd = &cobra.Command{
	Use:     "unstow [packages...]",
	Short:   unstowShort,
	Long:    unstowLong,
	Example: unstowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		appLogger.Debug("running unstow command", "args", args)
		dryrun, err := getBoolFlag(cmd.Flags(), flagDryRun)
		if err != nil {
			return err
		}
		engineCtx := engine.EngineContext{
			Source:      cfg.Source,
			Destination: cfg.Destination,
			ConfigHome:  config.AppConfigHome(),
		}
		eng, err := engine.NewEngine(&engineCtx, dryrun, appLogger)
		if err != nil {
			return err
		}
		ctx := engine.CommandContext{
			Action: engine.ActionUnstow,
			Args:   args,
		}
		summary, err := eng.Execute(&ctx)
		if err != nil {
			return err
		}
		appOutput.PrintSummary(summary)
		return nil
	},
}

func init() {
	addOperationFlags(unstowCmd.Flags())
	rootCmd.AddCommand(unstowCmd)
}
