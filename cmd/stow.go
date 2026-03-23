/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/ThisaruGuruge/bestow/internal/log"
	"github.com/spf13/cobra"
)

// stowCmd represents the stow command
var stowCmd = &cobra.Command{
	Use:   "stow",
	Short: "stows the provided packages from source to destination",
	Long: `stow will create symlinks to the dotfiles from a source to a destination.

each directory in the source is considered a package. Each file inside the package will be symlinked in the destination.
For example;

bestow stow nvim

will create symlinks for each file inside the 'nvim' directory, while maintaining the internal file strucutre.

If no packages are provided, all the pakcages inside the source will be stowed.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug("running stow command", "flags", cmd.Flags(), "args", args)
		ctx := engine.ActionContext{
			Action: engine.ActionStow,
			Args:   args,
		}
		err := engine.Execute(&ctx, cfg)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addOperationFlags(stowCmd.Flags())

	stowCmd.PersistentFlags().BoolP(FlagForce, "f", false, "remove the existing file and create the symlink")
	stowCmd.PersistentFlags().BoolP(FlagAdopt, "a", false, "move the existing file to the source and create the symlinks")

	rootCmd.AddCommand(stowCmd)
}
