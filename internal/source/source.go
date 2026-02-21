package source

import (
	"fmt"
	"os"

	"github.com/ThisaruGuruge/bestow/internal/clog"
)

type Input struct {
	Destination string
	Source      string
}

type Source interface {
	getArgs() (Input, error)
}

func GetArgs() (*Input, error) {
	return getDefaultArgs()
}

func getDefaultArgs() (*Input, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		clog.Error(err)
		return nil, fmt.Errorf("failed reading user home directory: %w", err)
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		clog.Error(err)
		return nil, fmt.Errorf("failed reading the current directory: %w", err)
	}
	return &Input{
		Destination: home,
		Source:      currentDirectory,
	}, nil
}
