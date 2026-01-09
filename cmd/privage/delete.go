package main

import (
	"fmt"
	"os"

	"github.com/revelaction/privage/setup"
)

// deleteCommand deletes an encrypted file from the repository
func deleteCommand(s *setup.Setup, label string, ui UI) error {
	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	found := false
	ch, err := headerGenerator(s.Repository, s.Id)
	if err != nil {
		return err
	}
	for h := range ch {
		if h.Label == label {
			err := os.Remove(h.Path)
			if err != nil {
				return err
			}
			found = true
			break
		}
	}

	if !found {
		_, _ = fmt.Fprintf(ui.Err, "could not find the encrypted file for %s\n", label)
	} else {
		_, _ = fmt.Fprintf(ui.Err, "deleted encrypted file for %s\n", label)
	}

	return nil
}
