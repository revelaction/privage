package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// showAction prints in the terminal partially/all the contents of an encrypted
// file.
func showAction(ctx *cli.Context) error {

	if ctx.Args().Len() == 0 {
		return errors.New("show command needs one argument")
	}

    s, err := setupEnv(ctx)
    if err != nil {
        return fmt.Errorf("Unable to setup environment configuration: %s", err)
	}

	// Flag to determine if we dump the file to stout (all) or only selected
	// parts of the file (for credential files we want only login and password
	// toml fields).
	isAll := ctx.Bool("all")

	if s.Id.Id == nil {
		return fmt.Errorf("Found no privage key file: %w", s.Id.Err)
	}

	label := ctx.Args().First()

	return show(label, s, isAll)
}

func show(label string, s *setup.Setup, isAll bool) error {

	for h := range headerGenerator(s.Repository, s.Id) {

		if h.Label == label {

			r, err := contentReader(h, s.Id)
			if err != nil {

				return err
			}

			switch h.Category {
			case header.CategoryCredential:

				// Check flag
				if !isAll {
					err := credential.LogFields(r)
					if err != nil {
						return err
					}

					return nil
				}

				// Show all if flag --all
				fallthrough

			default:
				if _, err := io.Copy(os.Stdout, r); err != nil {
					if err != io.ErrUnexpectedEOF {
						return err
					}
				}
			}

			return nil
		}
	}

	return nil
}
