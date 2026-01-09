package main

import (
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/setup"
)

// catCommand is a pure logic worker. It does not know about flags or usage.
// It assumes the label has been validated and the setup is successful.
func catCommand(s *setup.Setup, label string, ui UI) (err error) {

	if s.Id.Id == nil {
		return fmt.Errorf("%w: %v", ErrNoIdentity, s.Id.Err)
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

			if _, err := io.Copy(ui.Out, r); err != nil {
				if err != io.ErrUnexpectedEOF {
					return err
				}
			}

			return nil
		}
	}

	return fmt.Errorf("%w: %q", ErrFileNotFound, label)
}
