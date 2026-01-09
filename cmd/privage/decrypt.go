package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/revelaction/privage/setup"
)

var ErrHeaderNotFound = errors.New("header not found")

// decryptCommand decrypts an encrypted file
func decryptCommand(s *setup.Setup, label string, ui UI) (retErr error) {
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
			found = true
			w, err := os.Create(filepath.Join(s.Repository, label))
			if err != nil {
				return err
			}
			defer func() {
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()

			f, err := os.Open(h.Path)
			if err != nil {
				return err
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && retErr == nil {
					retErr = cerr
				}
			}()

			r, err := contentReader(f, s.Id)
			if err != nil {
				return err
			}

			bufFile := bufio.NewWriter(w)

			if _, err := io.Copy(bufFile, r); err != nil {
				if err != io.ErrUnexpectedEOF {
					return err
				}
			}

			if err := bufFile.Flush(); err != nil {
				return err
			}

			break
		}
	}

	if !found {
		return fmt.Errorf("file %q not found in directory", label)
	}

	_, _ = fmt.Fprintf(ui.Err, "The file %s was decrypted in the directory %s.\n", label, s.Repository)
	_, _ = fmt.Fprintln(ui.Err)
	_, _ = fmt.Fprintln(ui.Err, "(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
	_, _ = fmt.Fprintln(ui.Err, "(Use \"privage reencrypt --clean\" to reencrypt and delete all decrypted files)")

	return nil
}
