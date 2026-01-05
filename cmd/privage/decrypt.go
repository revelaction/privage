package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/revelaction/privage/setup"
)

var ErrHeaderNotFound = errors.New("header not found")

// decryptCommand decrypts an encrypted file
func decryptCommand(s *setup.Setup, args []string, ui UI) (retErr error) {
	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	fs.SetOutput(ui.Err)
	fs.Usage = func() {
		fmt.Fprintf(ui.Err, "Usage: %s decrypt [label]\n", os.Args[0])
		fmt.Fprintf(ui.Err, "\nDescription:\n")
		fmt.Fprintf(ui.Err, "  Decrypt a file and write its content in a file named after the label\n")
		fmt.Fprintf(ui.Err, "\nArguments:\n")
		fmt.Fprintf(ui.Err, "  label  The label of the file to decrypt\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) == 0 {
		return errors.New("decrypt command needs one argument (label)")
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]

	found := false
	for h := range headerGenerator(s.Repository, s.Id) {
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
		return fmt.Errorf("file %q not found in repository", label)
	}

	fmt.Fprintf(ui.Out, "The file %s was decrypted in the directory %s.\n", label, s.Repository)
	fmt.Fprintln(ui.Out)
	fmt.Fprintln(ui.Out, "(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
	fmt.Fprintln(ui.Out, "(Use \"privage reencrypt --clean\" to reencrypt and delete all decrypted files)")

	return nil
}
