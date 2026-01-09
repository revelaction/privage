package main

import (
	"fmt"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/setup"
)

// clipboardCommand copies the password field of a credential file to the clipboard
func clipboardCommand(s *setup.Setup, label string, ui UI) (err error) {
	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	ch, err := headerGenerator(s.Repository, s.Id)
	if err != nil {
		return err
	}
	for h := range ch {
		if h.Label == label {
			f, err := os.Open(h.Path)
			if err != nil {
				return err
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()

			r, err := contentReader(f, s.Id)
			if err != nil {
				return err
			}

			if err := credential.CopyClipboard(r); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(ui.Err, "The password for `%s` is in the clipboard\n", label)

			return nil
		}
	}

	return nil
}
