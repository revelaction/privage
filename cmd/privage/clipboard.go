package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/setup"
)

// clipboardCommand copies the password field of a credential file to the clipboard
func clipboardCommand(s *setup.Setup, args []string, ui UI) (err error) {
	fs := flag.NewFlagSet("clipboard", flag.ContinueOnError)
	fs.SetOutput(ui.Err)
	fs.Usage = func() {
		fmt.Fprintf(ui.Err, "Usage: %s clipboard [options] [label]\n", os.Args[0])
		fmt.Fprintf(ui.Err, "\nDescription:\n")
		fmt.Fprintf(ui.Err, "  Copy the credential password to the clipboard.\n")
		fmt.Fprintf(ui.Err, "\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(ui.Err, "\nArguments:\n")
		fmt.Fprintf(ui.Err, "  label  The label of the credential to copy\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) == 0 {
		return errors.New("clipboard command needs one argument (label)")
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]

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

			if err := credential.CopyClipboard(r); err != nil {
				return err
			}

			fmt.Fprintf(ui.Out, "The password for `%s` is in the clipboard\n", label)

			return nil
		}
	}

	return nil
}
