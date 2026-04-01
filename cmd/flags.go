package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/engine"
	"github.com/ThisaruGuruge/bestow/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	flagSource      = "source"
	flagDestination = "destination"
)

type boolFlagValue struct {
	name     string
	value    bool
	strategy engine.ResolveStrategy
}

func addOperationFlags(fs *pflag.FlagSet) {
	fs.StringP(flagSource, "s", "", "root directory of the source files (Eg.: `dotfiles` repo)")
	fs.StringP(flagDestination, "d", "", "destination directory of the symlinks (Eg.: `$HOME` directory)")
}

func bindOperationalFlags(cmd *cobra.Command, v *viper.Viper) {
	if f := cmd.Flags().Lookup(FlagProfile); f != nil {
		v.BindPFlag(FlagProfile, f)
	}
	profile := v.GetString(FlagProfile)
	if profile == "" {
		profile = "default"
	}
	prefix := fmt.Sprintf("profiles.%s", profile)
	if f := cmd.Flags().Lookup(flagSource); f != nil {
		v.BindPFlag(prefix+".source", f)
	}
	if f := cmd.Flags().Lookup(flagDestination); f != nil {
		v.BindPFlag(prefix+".destination", f)
	}
}

func checkVerbose(cmd *cobra.Command) error {
	verbose, err := cmd.Flags().GetBool(FlagVerbose)
	if err != nil {
		return err
	}
	if verbose {
		log.SetLevel(log.LevelDebug)
	}
	return nil
}

func conflictResolve(flagValues []boolFlagValue) (engine.ResolveStrategy, error) {
	enabledFlags := []boolFlagValue{}
	for _, flagValue := range flagValues {
		if flagValue.value {
			enabledFlags = append(enabledFlags, flagValue)
		}
	}
	if len(enabledFlags) > 1 {
		flags := []string{}
		for _, flag := range enabledFlags {
			flags = append(flags, flag.name)
		}
		return engine.ResolveSkip, errors.New(fmt.Sprintf("flags %s are mutually exclusive", strings.Join(flags, ", ")))
	}
	if len(enabledFlags) == 1 {
		return enabledFlags[0].strategy, nil
	}
	return engine.ResolveSkip, nil
}

func addConflictResolutionFlags(flags *pflag.FlagSet) {
	flags.BoolP(FlagForce, "f", false, "remove the existing file and create the symlink")
	flags.BoolP(FlagAdopt, "a", false, "move the existing file to the source and create the symlink")
	flags.BoolP(FlagBackup, "b", false, "rename the existing file to <filename>.bak and create the symlink")
	flags.BoolP(FlagInteractive, "i", false, "resolve conflicts interactively")
}
