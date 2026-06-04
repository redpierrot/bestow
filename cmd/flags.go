/*
All Rights Reversed (ɔ)
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	flagSource      = "source"
	flagDestination = "destination"
)

const (
	flagVerbose    string = "verbose"
	flagQuiet      string = "quiet"
	flagDryRun     string = "dry-run"
	flagConfigFile string = "config-file"
	flagProfile    string = "profile"
	flagForce      string = "force"
	flagAdopt      string = "adopt"
	flagBackup     string = "backup"
)

func addOperationFlags(fs *pflag.FlagSet) {
	fs.StringP(flagSource, "s", "", "root directory of the source files (e.g. `dotfiles` repo)")
	fs.StringP(flagDestination, "d", "", "destination directory of the symlinks (e.g. `$HOME` directory)")
}

func bindOperationalFlags(cmd *cobra.Command, v *viper.Viper) {
	if f := cmd.Flags().Lookup(flagProfile); f != nil {
		_ = v.BindPFlag(flagProfile, f)
	}
	profile := v.GetString(flagProfile)
	if profile == "" {
		profile = "default"
	}
	prefix := fmt.Sprintf("profiles.%s", profile)
	if f := cmd.Flags().Lookup(flagSource); f != nil {
		_ = v.BindPFlag(prefix+".source", f)
	}
	if f := cmd.Flags().Lookup(flagDestination); f != nil {
		_ = v.BindPFlag(prefix+".destination", f)
	}
}

func addConflictResolutionFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP(flagForce, "f", false, "remove the existing file and create the symlink")
	cmd.Flags().BoolP(flagAdopt, "a", false, "move the existing file to the source and create the symlink")
	cmd.Flags().BoolP(flagBackup, "b", false, "rename the existing file to <filename>.bestow.bak and create the symlink")

	cmd.MarkFlagsMutuallyExclusive(flagForce, flagAdopt, flagBackup)
}

func getBoolFlag(fs *pflag.FlagSet, name string) (bool, error) {
	val, err := fs.GetBool(name)
	if err != nil {
		return false, fmt.Errorf("parse flag %s: %w", name, err)
	}
	return val, nil
}

func getStringFlag(fs *pflag.FlagSet, name string) (string, error) {
	val, err := fs.GetString(name)
	if err != nil {
		return "", fmt.Errorf("parse flag %s: %w", name, err)
	}
	return val, nil
}
