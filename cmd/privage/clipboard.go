package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/setup"
)

func clipboardAction(args []string) error {
	fs := flag.NewFlagSet("clipboard", flag.ExitOnError)
	var deleteFlag bool
	fs.BoolVar(&deleteFlag, "delete", false, "Delete the contents of the clipboard")
	fs.BoolVar(&deleteFlag, "d", false, "alias for -delete")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if deleteFlag {
		err := credential.EmptyClipboard()
		if err != nil {
			return fmt.Errorf("could not emtpty the clipboard: %w", err)
		}

		return nil
	}

	if fs.NArg() == 0 {
		return errors.New("clipboard command needs one argument: label")
	}

	s, err := setupEnv(global.KeyFile, global.ConfigFile, global.RepoPath, global.PivSlot)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := fs.Arg(0)

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
				return fmt.Errorf("could not copy to clipboard: %w", err)
			}

			fmt.Printf("The password for `%s` is in the clipboard\n", label)
		}
	}

	return nil
}
