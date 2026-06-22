/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"path/filepath"
	"strings"
)

func readIgnoreFile(source string, reader IgnoreReader) ([]string, error) {
	ignoreFile := filepath.Join(source, ignoreFileName)
	exists, err := reader.Exists(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ignoreFile, err)
	}
	if !exists {
		return nil, nil
	}
	lines, err := reader.ReadLines(ignoreFile)
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
