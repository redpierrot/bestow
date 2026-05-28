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
func (e *Engine) init(ctx *CommandContext) error {
	e.Logger.Debug("initializing bestow")
	appConfigDir := config.AppConfigHome()
	if err := e.FileSystem.CreateDir(appConfigDir); err != nil {
		return fmt.Errorf("failed to create the config directory: %w", err)
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
	fullPath := filepath.Join(appConfigDir, constant.IgnoreFile)
	exists, err := e.FileSystem.Exists(fullPath)
	if err != nil {
		return &EngineError{
			Message: "failed to read the ignore file",
			Cause:   err,
			Hint:    "use --force to overwrite",
		}
	}
	if exists {
		if !force {
			return &EngineError{
				Message: fmt.Sprintf("failed to create ignore file; file '%s' already exists", fullPath),
				Cause:   err,
				Hint:    "Use '--force' to overwrite",
			}
		}
		e.Logger.Warn("ignore file exists; overwriting", "ignore-file", fullPath)
	}
	e.Logger.Debug("initializing ignore list", "ignore-list", ignoreList)

	ignoreFile := filepath.Join(appConfigDir, constant.IgnoreFile)
	if err := e.FileSystem.CreateFile(ignoreFile, getIgnoreFileContent(ignoreList)); err != nil {
		return err
	}
	output.Success("ignore file created successfully", "path", fullPath)
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
	fullPath := filepath.Join(appConfigDir, constant.ConfigFile)
	e.Logger.Debug("creating the config file", "path", fullPath)
	exists, err := e.FileSystem.Exists(fullPath)
	if err != nil {
		return &EngineError{
			Message: "failed to check config file path: %w",
			Cause:   err,
		}
	}
	if exists {
		if !force {
			return &EngineError{
				Message: fmt.Sprintf("failed to create config file; file '%s' already exists", fullPath),
				Hint:    "use --force to overwrite",
			}
		}
		e.Logger.Warn("config file exists; overwriting", "config-file", fullPath)
	}
	config, err := config.GetDefaultConfigTemplate(source, destination)
	if err != nil {
		return &EngineError{
			Message: "failed to load default configs: %w",
			Cause:   err,
		}
	}
	configFile := filepath.Join(appConfigDir, constant.ConfigFile)
	if err := e.FileSystem.CreateFile(configFile, config); err != nil {
		return &EngineError{
			Message: "failed to write to the config file: %w",
			Cause:   err,
		}
	}
	output.Success("config file created successfully", "path", fullPath)
	return nil
}
