package main

import (
	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// setupEnv initializes:
// - privage configuration file
// - Identity, using conf file if exists
// - Secrets repo path, using conf file if exists
func setupEnv(ctx *cli.Context) (*setup.Setup, error) {

	argConf := ctx.String("conf")
	// conf Can be empty struct
	conf, err := config.New(argConf)
	if err != nil {
		return &setup.Setup{},err
	}

    // override conf file with arg params
	argKey := ctx.String("key")
    if "" != argKey {
        conf.IdentityPath=argKey
    }
	argRepo := ctx.String("repository")
    if "" != argRepo {
        conf.RepositoryPath=argRepo
    }



	// find identity and repository for encrypted files
	s, err := setup.New(conf)
	if err != nil {
		return &setup.Setup{},err
	}

	return s, nil
}
