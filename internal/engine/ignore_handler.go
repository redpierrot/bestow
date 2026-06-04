/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/constant"
)

func readIgnoreFile(source string, fileSystem FileSystem) ([]string, error) {
	ignoreFile := filepath.Join(source, constant.IgnoreFile)
	exists, err := fileSystem.Exists(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ignoreFile, err)
	}
	if !exists {
		return nil, nil
	}
	lines, err := fileSystem.ReadLines(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ignoreFile, err)
	}
	patterns := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		patterns = append(patterns, trimmed)
	}
	return patterns, nil
}
