package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/setup"
)

// catAction prints in the terminal the contents of an encrypted file.
func catAction(args []string) error {

	if len(args) == 0 {
		return errors.New("cat command needs one argument (label)")
	}

	s, err := setupEnv(global)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]

	return cat(label, s)
}

func cat(label string, s *setup.Setup) error {

	for h := range headerGenerator(s.Repository, s.Id) {

		if h.Label == label {

			r, err := contentReader(h, s.Id)
			if err != nil {
				return err
			}

			if _, err := io.Copy(os.Stdout, r); err != nil {
				if err != io.ErrUnexpectedEOF {
					return err
				}
			}

			return nil
		}
	}

	return nil
}
