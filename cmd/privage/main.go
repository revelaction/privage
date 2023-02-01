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
			Name:    "conf",
			Aliases: []string{"c"},
			Usage:   "Use file as privage configuration file",
		},
		&cli.StringFlag{
			Name:    "key",
			Aliases: []string{"k"},
			Usage:   "Use file path for private key",
		},
		&cli.StringFlag{
			Name:    "piv-slot",
			Aliases: []string{"p"},
			Usage:   "The PIV slot for decryption of the age key",
		},
		&cli.StringFlag{
			Name:    "repository",
			Aliases: []string{"r"},
			Usage:   "Use file path as path for the repository",
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
