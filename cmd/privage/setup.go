package main

import (
	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// setupApp checks if these exists:
// - privage configuration file
// - Identity, using conf file if exists
// - Secrets repo path, using conf file if exists
func setupApp(app *cli.App) error {

	metadata := map[string]interface{}{}

	// conf Can be empty struct
	conf, err := config.New()
	if err != nil {
		return err
	}

	// find identity and repository for encrypted files
	setup, err := setup.New(conf)
	if err != nil {
		return err
	}

	metadata["setup"] = setup
	app.Metadata = metadata
	return nil
}
