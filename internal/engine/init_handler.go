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
)

type InitContext struct {
	Force      bool
	IgnoreList []string
}

func (e *Engine) Init(ctx *InitContext) (*ExecuteSummary, error) {
	e.logger.Debug("initializing bestow")
	appConfigDir := config.AppConfigHome()
	configFile := filepath.Join(appConfigDir, constant.ConfigFile)
	ignoreFile := filepath.Join(appConfigDir, constant.IgnoreFile)
	if err := e.checkExistingFiles(configFile, ignoreFile, ctx.Force); err != nil {
		return nil, err
	}
	if err := e.fileSystem.CreateDir(appConfigDir); err != nil {
		return nil, err
	}
	configAction, err := e.createConfigFile(e.source, e.destination, configFile)
	if err != nil {
		return nil, err
	}
	ignoreAction, err := e.createIgnoreFile(ignoreFile, ctx.IgnoreList)
	if err != nil {
		return nil, err
	}
	actions := []ActionEvent{*configAction, *ignoreAction}
	return &ExecuteSummary{Actions: actions, OperationSummary: &Summary{}}, nil
}

func (e *Engine) checkExistingFiles(configFile, ignoreFile string, force bool) error {
	if force {
		return nil
	}
	configExists, err := e.fileSystem.Exists(configFile)
	if err != nil {
		return err
	}
	ignoreExists, err := e.fileSystem.Exists(ignoreFile)
	if err != nil {
		return err
	}
	existing := make([]string, 0, 2)
	if configExists {
		existing = append(existing, configFile)
	}
	if ignoreExists {
		existing = append(existing, ignoreFile)
	}
	if len(existing) > 0 {
		fileString := strings.Join(existing, ", ")
		return &HintedError{
			Op:   fmt.Sprintf("exists %s", fileString),
			Hint: "remove the existing files or use --force",
			Err:  ErrFileExists,
		}
	}
	return nil
}

func (e *Engine) createIgnoreFile(ignoreFile string, ignoreList []string) (*ActionEvent, error) {
	e.logger.Debug("creating ignore file", "filepath", ignoreFile)
	e.logger.Debug("initializing ignore list", "ignore-list", ignoreList)
	if err := e.fileSystem.CreateFile(ignoreFile, getIgnoreFileContent(ignoreList)); err != nil {
		return nil, err
	}
	return &ActionEvent{
		Action:    actionCreated,
		Msg:       ignoreFile,
		EventType: EventSuccess,
	}, nil
}

func getIgnoreFileContent(ignoreList []string) string {
	result := []string{"# Global Ignore List for Bestow"}
	for _, item := range ignoreList {
		result = append(result, strings.TrimSpace(item))
	}
	return strings.Join(result, "\n")
}

func (e *Engine) createConfigFile(source, destination, configFile string) (*ActionEvent, error) {
	e.logger.Debug("creating the config file", "path", configFile)
	config, err := config.GetDefaultConfigTemplate(source, destination)
	if err != nil {
		return nil, fmt.Errorf("load config %s %s: %w", source, destination, err)
	}
	if err := e.fileSystem.CreateFile(configFile, config); err != nil {
		return nil, err
	}
	return &ActionEvent{
		Action:    actionCreated,
		Msg:       configFile,
		EventType: EventSuccess,
	}, nil
}
