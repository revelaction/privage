package main

import (
    "errors"
	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// setupEnv initializes:
// - privage configuration file
// - Identity, using conf file if exists
// - Secrets repo path, using conf file if exists
func setupEnv(ctx *cli.Context) (*setup.Setup, error) {


	// conf Can be empty struct
    // main motto:
    // arguments -k, -r are incompatble with configurtaion files -c or conf search.
    //
    // at the end we want  a setup object
    // a key is required
    // if no key found fail 
    //
    // * key
    // if -k arg take it as key, validate it?
    //  if -p key take it as a yubico key
    //  if -c throw error, excluded with -k
    // stop key finding
    // if -k, not -c allowed, no conf search allowed.
    // if -k, and not -r:  repo is current dir
    // 
    // if -c parse it and search for key
    // if no -k search for a conf file, parse it and get its key if not empty
    // if still not -k, stop it with error. 
    //
    // * path 
    // if -r take it.
    // if 
	argKey := ctx.String("key")
	argConf := ctx.String("conf")
	argRepository := ctx.String("repository")
    argPivSlot:= ctx.String("piv")

    if "" != argKey {
        if "" != argConf {
		    return &setup.Setup{}, errors.New("flags -c and -k are incompatible")
        }

        s, err := setup.NewFromArgs(argKey, argRepository, argPivSlot)
        if err != nil {
            return &setup.Setup{},err
        }

        return s, nil

    }

    if "" != argConf {
        s, err := setup.NewFromConfigFile(argConf)
        if err != nil {
            return &setup.Setup{},err
        }

        return s, nil
    }

    path, err := config.FindPath()
    if err != nil {
        return &setup.Setup{},err
    }

    s, err := setup.NewFromConfigFile(path)
    if err != nil {
        return &setup.Setup{},err
    }

    return s, nil

}

