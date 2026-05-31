/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/output"
)

// TODO: Check both files before touching them
// Check dryrun before executing.
func (e *Engine) init(ctx *CommandContext) error {
	e.Logger.Debug("initializing bestow")
	appConfigDir := config.AppConfigHome()
	if ctx.DryRun {
		output.Success("[create]", "directory", appConfigDir)

	} else {
		if err := e.FileSystem.CreateDir(appConfigDir); err != nil {
			return err
		}
	}
	if err := e.createConfigFile(e.Source, e.Destination, ctx.Force, appConfigDir); err != nil {
		return err
	}
	if err := e.createIgnoreFile(appConfigDir, ctx.Force, ctx.IgnoreList); err != nil {
		return err
	}
	return nil
}

func (e *Engine) createIgnoreFile(appConfigDir string, force bool, ignoreList []string) error {
	e.Logger.Debug("creating ignore file")
	ignoreFile := filepath.Join(appConfigDir, constant.IgnoreFile)
	exists, err := e.FileSystem.Exists(ignoreFile)
	if err != nil {
		return err
	}
	if exists {
		if !force {
			return &HintedError{
				Op:   fmt.Sprintf("create ignorefile %s", ignoreFile),
				Hint: "use --force to overwrite",
				Err:  ErrFileExists,
			}
		}
		e.Logger.Warn("ignore file exists; overwriting", "ignore-file", ignoreFile)
	}
	e.Logger.Debug("initializing ignore list", "ignore-list", ignoreList)
	if err := e.FileSystem.CreateFile(ignoreFile, getIgnoreFileContent(ignoreList)); err != nil {
		return err
	}
	output.Success("ignore file created successfully", "path", ignoreFile)
	return nil
}

func getIgnoreFileContent(ignoreList []string) string {
	result := []string{"# Global Ignore List for Bestow"}
	for _, item := range ignoreList {
		result = append(result, strings.TrimSpace(item))
	}
	return strings.Join(result, "\n")
}

func (e *Engine) createConfigFile(source, destination string, force bool, appConfigDir string) error {
	configFile := filepath.Join(appConfigDir, constant.ConfigFile)
	e.Logger.Debug("creating the config file", "path", configFile)
	exists, err := e.FileSystem.Exists(configFile)
	if err != nil {
		return err
	}
	if exists {
		if !force {
			return &HintedError{
				Op:   fmt.Sprintf("create configfile %s", configFile),
				Hint: "use --force to overwrite",
				Err:  ErrFileExists,
			}
		}
		e.Logger.Warn("config file exists; overwriting", "config-file", configFile)
	}
	config, err := config.GetDefaultConfigTemplate(source, destination)
	if err != nil {
		return fmt.Errorf("load config %s %s: %w", source, destination, err)
	}
	if err := e.FileSystem.CreateFile(configFile, config); err != nil {
		return err
	}
	output.Success("config file created successfully", "path", configFile)
	return nil
}
