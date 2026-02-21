package config

import (
	"github.com/ThisaruGuruge/bestow/internal/clog"
)

var ignoreFileList = []string{".git", ".DS_Store", "README.md", "LICENSE", ".gitignore"}

type Config struct {
	Ignore []string
}

func GetConfigs() *Config {
	defaultConfigs, err := getDefaultConfigs()
	if err != nil {
		clog.Error(err)
	}

	return defaultConfigs
}

func getDefaultConfigs() (*Config, error) {
	return &Config{
		Ignore: ignoreFileList,
	}, nil
}
