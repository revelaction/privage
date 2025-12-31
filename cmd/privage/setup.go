package main

import (
	"errors"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// setupEnv loads the identity file (secret key) and the repository command
// arguments have preference. Explicite -k, -r arguments have preference over
// the -c argument.  If no arguments provided, standard paths are searched for
// a configuration file .privage.conf
func setupEnv(opts setup.Options) (*setup.Setup, error) {

	if opts.KeyFile != "" {
		if opts.ConfigFile != "" {
			return &setup.Setup{}, errors.New("flags -c and -k are incompatible")
		}

		s, err := setup.NewFromArgs(opts.KeyFile, opts.RepoPath, opts.PivSlot)
		if err != nil {
			return &setup.Setup{}, err
		}

		return s, nil
	}

	if opts.ConfigFile != "" {
		s, err := setup.NewFromConfigFile(opts.ConfigFile)
		if err != nil {
			return &setup.Setup{}, err
		}

		return s, nil
	}

	path, err := config.FindPath()
	if err != nil {
		return &setup.Setup{}, err
	}

	s, err := setup.NewFromConfigFile(path)
	if err != nil {
		return &setup.Setup{}, err
	}

	return s, nil

}
