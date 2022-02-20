// Copyright (c) 2022 The privage developers

package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	BuildCommit string
	BuildTag    string
)

func main() {

	app := &cli.App{}

	// Pass the setup configuration to the cli app.
	err := setupApp(app)
	if err != nil {
		fatal(err)
	}

	app.Name = "privage"
	app.Version = BuildTag
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s, commit %s\n", app.Name, c.App.Version, BuildCommit)
	}

	app.EnableBashCompletion = true
	app.UseShortOptionHandling = true
	app.Usage = "password manager/encryption tool based on age"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "identity",
			Aliases: []string{"i"},
			Usage:   "Use the identity file at PATH",
		},
	}

	app.Commands = allCommands

	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "privage: %v\n", err)
	os.Exit(1)
}
