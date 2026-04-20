/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/file"
	"github.com/ThisaruGuruge/bestow/internal/log"
	"github.com/spf13/cobra"
)

const (
	flagInitIgnoreList  string = "ignore-list"
	flagInitSource      string = "source"
	flagInitDestination string = "destination"
	flagInitForce       string = "force"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: InitShort,
	Long:  InitLong,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := checkVerbose(cmd); err != nil {
			return fmt.Errorf("failed to check flags: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString(flagInitSource)
		destination, _ := cmd.Flags().GetString(flagInitDestination)
		force, _ := cmd.Flags().GetBool(flagInitForce)
		appConfigDir := config.AppConfigHome()
		if err := file.CreateDir(appConfigDir); err != nil {
			return fmt.Errorf("failed to create the config directory: %w", err)
		}
		if err := createConfigFile(source, destination, force, appConfigDir); err != nil {
			return err
		}
		if err := createIgnoreFile(appConfigDir, force, cmd); err != nil {
			return err
		}
		log.Info("successfully initialized bestow")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringSlice(flagInitIgnoreList, config.DefaultIgnoreList, "list of file/directory names bestow should ignore. This is the global set of values. For repo or package specific ignore lists, use specific .bestowignore files")
	initCmd.Flags().StringP(flagInitSource, "s", "", "source of dotfiles for symlinks; written to 'config.yaml'. (defaults to '$HOME/dotfiles')")
	initCmd.Flags().StringP(flagInitDestination, "d", "", "destination for the dotfiles symlinks; written to 'config.yaml'. (defaults to '$HOME')")
	initCmd.Flags().BoolP(flagInitForce, "f", false, "forcefully overwrite any existing config files for bestow")
}

func createConfigFile(source, destination string, force bool, appConfigDir string) error {
	fullPath := filepath.Join(appConfigDir, constant.ConfigFile)
	log.Debug("creating the config file", "path", fullPath)
	exists, err := file.Exists(fullPath)
	if err != nil {
		return fmt.Errorf("failed to check config file: %w", err)
	}
	if exists {
		if !force {
			return fmt.Errorf("failed to create config file; file '%s' already exists. Use '--force' to overwrite", fullPath)
		}
		log.Warn("config file exists; overwriting", "config-file", fullPath)
	}
	config, err := config.GetDefaultConfigTemplate(source, destination)
	if err != nil {
		return fmt.Errorf("failed to load default configs: %w", err)
	}
	if err := file.CreateFile(constant.ConfigFile, appConfigDir, config); err != nil {
		return fmt.Errorf("failed to write the config file: %w", err)
	}
	log.Debug("successfully created the config file", "path", fullPath)
	return nil
}

func createIgnoreFile(appConfigDir string, force bool, cmd *cobra.Command) error {
	ignoreList, err := cmd.Flags().GetStringSlice(flagInitIgnoreList)
	if err != nil {
		return fmt.Errorf("failed to read the ignore list: %w", err)
	}
	fullPath := filepath.Join(appConfigDir, constant.IgnoreFile)
	exists, err := file.Exists(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read the ignore file: %w", err)
	}
	if exists {
		if !force {
			return fmt.Errorf("failed to create ignore file; file '%s' already exists. Use '--force' to overwrite", fullPath)
		}
		log.Warn("ignore file exists; overwriting", "ignore-file", fullPath)
	}
	log.Debug("initializing ignore list", flagInitIgnoreList, ignoreList)

	if err := file.CreateFile(constant.IgnoreFile, appConfigDir, getIgnoreFileContent(ignoreList)); err != nil {
		return err
	}
	log.Debug("successfully created the ignore file", "path", fullPath)
	return nil
}

func getIgnoreFileContent(ignoreList []string) string {
	result := []string{"# Global Ignore List for Bestow"}
	for _, item := range ignoreList {
		result = append(result, strings.TrimSpace(item))
	}
	return strings.Join(result, "\n")
}
