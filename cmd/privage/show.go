package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// showCommand prints in the terminal partially/all the contents of an encrypted
// file.
func showCommand(s *setup.Setup, label string, fieldName string, ui UI) (err error) {
	if s.Id.Id == nil {
		return fmt.Errorf("%w: %v", ErrNoIdentity, s.Id.Err)
	}

	ch, err := headerGenerator(s.Repository, s.Id)
	if err != nil {
		return err
	}
	var match *header.Header
	for h := range ch {
		if h.Label == label {
			match = h
			break
		}
		if strings.Contains(h.Label, label) {
			if match != nil {
				return fmt.Errorf("%w: %q", ErrAmbiguousLabel, label)
			}
			match = h
		}
	}

	if match == nil {
		return fmt.Errorf("%w: %q", ErrFileNotFound, label)
	}

	f, err := os.Open(match.Path)
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

	if match.Category != header.CategoryCredential {
		return fmt.Errorf("%w: file '%s' is not a credential. Use 'privage cat %s' to view its contents", ErrNotCredential, label, label)
	}

	cred, err := credential.Decode(r)
	if err != nil {
		return err
	}

	if fieldName != "" {
		val, ok := cred.GetField(fieldName)
		if !ok {
			return fmt.Errorf("%w: field '%s' not found in credential '%s'", ErrFieldNotFound, fieldName, label)
		}
		if _, err := fmt.Fprint(ui.Out, val); err != nil {
			return err
		}
		return nil
	}

	return cred.FprintBasic(ui.Out)
}
