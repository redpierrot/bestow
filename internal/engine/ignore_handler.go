/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"path/filepath"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/file"
)

func readIgnoreFile(source string, patterns *[]string, fileSystem file.System) error {
	filePath := filepath.Join(source, constant.IgnoreFile)
	exists, err := fileSystem.Exists(filePath)
	if err != nil {
		return &EngineError{
			Message: "error occurred while reading the ignore file",
			Cause:   err,
		}
	}
	if !exists {
		return nil
	}
	lines, err := fileSystem.ReadLines(filePath)
	if err != nil {
		return &EngineError{
			Message: "failed to read the ignore file",
			Cause:   err,
		}
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		*patterns = append(*patterns, trimmed)
	}
	return nil
}
