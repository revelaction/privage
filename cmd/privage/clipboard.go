package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/setup"
)

func clipboardAction(ctx *cli.Context) error {

	if ctx.Bool("delete") {
		err := credential.EmptyClipboard()
		if err != nil {
			return fmt.Errorf("Could not emtpty the clipboard: %w", err)
		}

		return nil
	}

	if ctx.Args().Len() == 0 {
		return errors.New("clipboard command needs one argument: label")
	}

	s, err := setupEnv(ctx)
	if err != nil {
		return fmt.Errorf("Unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("Found no privage key file: %w", s.Id.Err)
	}

	label := ctx.Args().First()

	return clipboard(label, s)
}

func clipboard(label string, s *setup.Setup) error {

	for h := range headerGenerator(s.Repository, s.Id) {

		if h.Label == label {

			r, err := contentReader(h, s.Id)
			if err != nil {
				return err
			}

			err = credential.CopyClipboard(r)
			if err != nil {
				return fmt.Errorf("Could not copy to clipboard: %w", err)
			}

			fmt.Printf("The password for `%s` is in the clipboard\n", label)
		}
	}

	return nil
}
