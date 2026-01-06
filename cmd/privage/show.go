package main

import (
	"fmt"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// showCommand prints in the terminal partially/all the contents of an encrypted
// file.
func showCommand(s *setup.Setup, label string, fieldName string, ui UI) (err error) {
	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	for h := range headerGenerator(s.Repository, s.Id) {
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

			if h.Category != header.CategoryCredential {
				return fmt.Errorf("file '%s' is not a credential. Use 'privage cat %s' to view its contents", label, label)
			}

			cred, err := credential.Decode(r)
			if err != nil {
				return err
			}

			if fieldName != "" {
				val, ok := cred.GetField(fieldName)
				if !ok {
					return fmt.Errorf("field '%s' not found in credential '%s'", fieldName, label)
				}
				if _, err := fmt.Fprint(ui.Out, val); err != nil {
					return err
				}
				return nil
			}

			return cred.FprintBasic(ui.Out)
		}
	}

	return fmt.Errorf("file %q not found in repository", label)
}