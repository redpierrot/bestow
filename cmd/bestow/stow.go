package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ThisaruGuruge/bestow/internal/clog"
	"github.com/ThisaruGuruge/bestow/internal/config"
	"github.com/ThisaruGuruge/bestow/internal/source"
	"github.com/ThisaruGuruge/bestow/internal/ui"
)

func Execute(config config.Config) {
	args, err := source.GetArgs()
	if err != nil {
		clog.Error(err)
	}
	findDotfiles(&config, args)
}

func findDotfiles(configs *config.Config, args *source.Input) {
	clog.Info("Checking dotfiles")
	clog.Info("Ignore: " + strings.Join(configs.Ignore, ", "))
	clog.Info("Reading files from the dotfiles")
	filesToLink := []string{}

	err := readDir(args.Source, configs.Ignore, &filesToLink)
	if err != nil {
		clog.Error(err)
	}
	for _, file := range filesToLink {
		ui.PrintPath(file)
	}
}

func validateInputs() {}

func checkDirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func readDir(path string, ignore []string, filesToLink *[]string) error {
	isDir, err := checkDirExists(path)
	if err != nil {
		err := fmt.Errorf("Failed to read the files: %w", err)
		clog.Error(err)
		return err
	}
	if !isDir {
		msg := fmt.Sprintf("Directory %s does not exist", path)
		clog.Warn(msg)
		return errors.New(msg)
	}
	files, err := os.ReadDir(path)
	if err != nil {
		clog.Error(err)
		return err
	}
	for _, file := range files {
		path := filepath.Join(path, file.Name())
		if slices.Contains(ignore, file.Name()) {
			clog.Warn(fmt.Sprintf("Ignoring %s as per the ignore list", path))
			continue
		}
		if file.IsDir() {
			clog.Info(fmt.Sprintf("Reading from the '%s'", path))
			readDir(path, ignore, filesToLink)
			continue
		}
		*filesToLink = append(*filesToLink, path)
	}
	return nil
}
