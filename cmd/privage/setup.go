package main

import (
	"fmt"
	"os"

	"github.com/revelaction/privage/fs"
	"github.com/revelaction/privage/identity"
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
		s, err := setup.NewFromKeyRepoFlags(opts.KeyFile, opts.RepoPath, opts.PivSlot)
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
		// Case 3: Nothing - try config file first, then identity file
		configPath, err := fs.FindConfigFile()
		if err == nil {
			// Config file found - use it
			s, err := setup.NewFromConfigFile(configPath)
			if err != nil {
				return &setup.Setup{}, err
			}
			return s, nil
		}

		// No config file found - search for identity file
		idPath, err := fs.FindIdentityFile()
		if err != nil {
			return &setup.Setup{}, fmt.Errorf("no config or identity file found: %w", err)
		}

		// Use current directory as repository
		repoPath, err := os.Getwd()
		if err != nil {
			return &setup.Setup{}, fmt.Errorf("could not get current directory: %w", err)
		}

		// Create setup with identity file and current directory as repo
		id := identity.Load(idPath)
		if id.Err != nil {
			return &setup.Setup{}, id.Err
		}

		return &setup.Setup{
			C:          nil, // No config when using auto-discovery
			Id:         id,
			Repository: repoPath,
		}, nil

	default:
		// This should never happen due to Validate()
		return &setup.Setup{}, fmt.Errorf("invalid option state")
	}
}
