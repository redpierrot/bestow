/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/ThisaruGuruge/bestow/internal/log"
	"github.com/spf13/cobra"
)

// unstowCmd represents the unstow command
var unstowCmd = &cobra.Command{
	Use:     "unstow [packages...]",
	Short:   UnstowShort,
	Long:    UnstowLong,
	Example: UnstowExamples,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug("running stow command", "args", args)
		ctx := engine.CommandContext{
			Action: engine.ActionUnstow,
			Args:   args,
		}
		engine, err := engine.NewEngine(&ctx, cfg)
		if err != nil {
			return err
		}
		if err := engine.Execute(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addOperationFlags(unstowCmd.Flags())
	rootCmd.AddCommand(unstowCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unstowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unstowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
