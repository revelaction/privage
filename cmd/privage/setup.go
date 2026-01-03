package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// setupEnv loads the identity file (secret key) and the repository command
// arguments have preference. Explicite -k, -r arguments have preference over
// the -c argument.  If no arguments provided, standard paths are searched for
// a configuration file .privage.conf
func setupEnv(opts setup.Options) (*setup.Setup, error) {
	// Validate options first
	if err := opts.Validate(); err != nil {
		return &setup.Setup{}, err
	}

	// Determine which case we're in using explicit methods
	switch {
	case opts.WithKeyRepo():
		// Case 1: -k -r with optional -p
		s, err := setup.NewFromArgs(opts.KeyFile, opts.RepoPath, opts.PivSlot)
		if err != nil {
			return &setup.Setup{}, err
		}
		return s, nil

	case opts.WithConfig():
		// Case 2: -c only
		s, err := setup.NewFromConfigFile(opts.ConfigFile)
		if err != nil {
			return &setup.Setup{}, err
		}
		return s, nil

	case opts.NoFlags():
		// Case 3: Nothing - search for config file
		path, err := findConfigPath()
		if err != nil {
			return &setup.Setup{}, err
		}

		s, err := setup.NewFromConfigFile(path)
		if err != nil {
			return &setup.Setup{}, err
		}
		return s, nil

	default:
		// This should never happen due to Validate()
		return &setup.Setup{}, fmt.Errorf("invalid option state")
	}
}

func findConfigPath() (string, error) {
	locations := []func() (string, error){
		os.UserHomeDir,
		os.Getwd,
	}

	for _, getDir := range locations {
		dir, err := getDir()
		if err != nil {
			continue
		}

		path := filepath.Join(dir, config.DefaultFileName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not find configuration file %s in home or current directory", config.DefaultFileName)
}
