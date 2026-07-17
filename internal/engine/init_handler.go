/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/redpierrot/bestow/internal/config"
)

const (
	ignoreFileName = ".bestowignore"
)

// InitConfig stores the configuration options for the `init` command
type InitConfig struct {
	Force      bool
	IgnoreList []string
	ConfigFile string
}

// Init initializes the bestow command.
// If the existing files ($CONFIG_HOME/bestow/config.yaml, $CONFIG_HOME/bestow/.bestowignore) are there the command will fail unless the `--force` flag is set
func (e *Engine) Init(cfg *InitConfig) (*ExecuteResult, error) {
	e.logger.Debug("initializing bestow")
	configFile := filepath.Join(e.configHome, cfg.ConfigFile)
	ignoreFile := filepath.Join(e.configHome, ignoreFileName)
	if err := e.checkExistingFiles(configFile, ignoreFile, cfg.Force); err != nil {
		return nil, err
	}
	if err := e.fileSystem.CreateDir(e.configHome); err != nil {
		return nil, err
	}
	configFileContent, err := config.FromTemplate(e.source, e.destination)
	if err != nil {
		return nil, fmt.Errorf("load config %s %s: %w", e.source, e.destination, err)
	}
	configAction, err := e.createFile(configFile, configFileContent)
	if err != nil {
		return nil, err
	}
	ignoreFileContent := buildIgnoreFile(cfg.IgnoreList)
	ignoreAction, err := e.createFile(ignoreFile, ignoreFileContent)
	if err != nil {
		return nil, err
	}
	actions := []ActionEvent{*configAction, *ignoreAction}
	return &ExecuteResult{Events: actions}, nil
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
			Err:  errFileExists,
		}
	}
	return nil
}

func (e *Engine) createFile(path, content string) (*ActionEvent, error) {
	e.logger.Debug("creating the file", "path", path)
	if err := e.fileSystem.CreateFile(path, content); err != nil {
		return nil, err
	}
	return &ActionEvent{
		Action:    fileOpCreated,
		Msg:       path,
		EventType: EventSuccess,
	}, nil
}

func buildIgnoreFile(ignoreList []string) string {
	var sb strings.Builder
	sb.WriteString("# Global Ignore List for Bestow")
	sb.WriteByte('\n')
	for _, item := range ignoreList {
		sb.WriteString(strings.TrimSpace(item))
		sb.WriteByte('\n')
	}
	return sb.String()
}
