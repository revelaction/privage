package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/setup"
)

// catCommand prints in the terminal the contents of an encrypted file.
func catCommand(s *setup.Setup, args []string, ui UI) (err error) {
	fs := flag.NewFlagSet("cat", flag.ContinueOnError)
	fs.SetOutput(ui.Err)
	fs.Usage = func() {
		fmt.Fprintf(ui.Err, "Usage: %s cat [label]\n", os.Args[0])
		fmt.Fprintf(ui.Err, "\nDescription:\n")
		fmt.Fprintf(ui.Err, "  Print the full contents of an encrypted file to stdout.\n")
		fmt.Fprintf(ui.Err, "\nArguments:\n")
		fmt.Fprintf(ui.Err, "  label  The label of the file to show\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) == 0 {
		return errors.New("cat command needs one argument (label)")
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

			if _, err := io.Copy(ui.Out, r); err != nil {
				if err != io.ErrUnexpectedEOF {
					return err
				}
			}

			return nil
		}
	}

	return nil
}