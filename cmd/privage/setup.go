package main

import (
	"errors"
	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// setupEnv loads the identity file (secret key) and the repository command
// arguments have preference. Explicite -k, -r arguments have preference over
// the -c argument.  If no arguments provided, standard paths are searched for
// a configuration file .privage.conf
func setupEnv(ctx *cli.Context) (*setup.Setup, error) {

	argKey := ctx.String("key")
	argConf := ctx.String("conf")
	argRepository := ctx.String("repository")
	argPivSlot := ctx.String("piv-slot")

	if "" != argKey {
		if "" != argConf {
			return &setup.Setup{}, errors.New("flags -c and -k are incompatible")
		}

		s, err := setup.NewFromArgs(argKey, argRepository, argPivSlot)
		if err != nil {
			return &setup.Setup{}, err
		}

		return s, nil
	}

	if "" != argConf {
		s, err := setup.NewFromConfigFile(argConf)
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
