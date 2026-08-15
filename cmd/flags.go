/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagSource      = "source"
	flagDestination = "destination"
)

const (
	flagVerbose       = "verbose"
	flagQuiet         = "quiet"
	flagDryRun        = "dry-run"
	flagConfigFile    = "config-file"
	flagProfile       = "profile"
	flagForce         = "force"
	flagAdopt         = "adopt"
	flagBackup        = "backup"
	flagKeepEmptyDirs = "keep-empty-dirs"
)

func addOperationFlags(fs *pflag.FlagSet) {
	fs.StringP(flagSource, "s", "", "root directory of the source files (e.g. `dotfiles` repo)")
	fs.StringP(flagDestination, "d", "", "destination directory of the symlinks (e.g. `$HOME` directory)")
	fs.SortFlags = false
}

func addConflictResolutionFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP(flagForce, "f", false, "remove the existing file and create the symlink")
	cmd.Flags().BoolP(flagAdopt, "a", false, "move the existing file to the source and create the symlink")
	cmd.Flags().BoolP(flagBackup, "b", false, "rename the existing file to <filename>.bestow.bak and create the symlink")

	cmd.MarkFlagsMutuallyExclusive(flagForce, flagAdopt, flagBackup)
}

func boolFlag(fs *pflag.FlagSet, name string) (bool, error) {
	val, err := fs.GetBool(name)
	if err != nil {
		return false, fmt.Errorf("parse flag %s: %w", name, err)
	}
	return val, nil
}

func stringFlag(fs *pflag.FlagSet, name string) (string, error) {
	val, err := fs.GetString(name)
	if err != nil {
		return "", fmt.Errorf("parse flag %s: %w", name, err)
	}
	return val, nil
}
