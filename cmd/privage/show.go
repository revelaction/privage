package main

import (
	"errors"
	"fmt"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// showCommand prints in the terminal partially/all the contents of an encrypted
// file.
func showCommand(opts setup.Options, args []string) error {

	if len(args) == 0 {
		return errors.New("show command needs one argument")
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]

	return show(label, s)
}

func show(label string, s *setup.Setup) error {

	for h := range headerGenerator(s.Repository, s.Id) {

		if h.Label == label {

			r, err := contentReader(h, s.Id)
			if err != nil {

				return err
			}

			if h.Category != header.CategoryCredential {
				return fmt.Errorf("file '%s' is not a credential. Use 'privage cat %s' to view its contents", label, label)
			}

			err = credential.LogFields(r)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}
