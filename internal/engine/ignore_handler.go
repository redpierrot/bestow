/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/constant"
	"github.com/ThisaruGuruge/bestow/internal/file"
)

func readIgnoreFile(source string, patterns *[]string, fileSystem file.System) error {
	ignoreFile := filepath.Join(source, constant.IgnoreFile)
	exists, err := fileSystem.Exists(ignoreFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", ignoreFile, err)
	}
	if !exists {
		return nil
	}
	lines, err := fileSystem.ReadLines(ignoreFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", ignoreFile, err)
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
