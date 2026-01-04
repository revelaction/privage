package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/setup"
)

func clipboardCommand(opts setup.Options, args []string) error {
	fs := flag.NewFlagSet("clipboard", flag.ContinueOnError)
	var deleteFlag bool
	fs.BoolVar(&deleteFlag, "delete", false, "Delete the contents of the clipboard")
	fs.BoolVar(&deleteFlag, "d", false, "alias for -delete")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s clipboard [options] [label]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Copy the credential password to the clipboard.\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  label  The label of the credential to copy\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if deleteFlag {
		err := credential.EmptyClipboard()
		if err != nil {
			return fmt.Errorf("could not emtpty the clipboard: %w", err)
		}

		return nil
	}

	if fs.NArg() == 0 {
		return errors.New("clipboard command needs one argument: label")
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := fs.Arg(0)

	return clipboard(label, s)
}

func clipboard(label string, s *setup.Setup) (err error) {

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

			r, err := contentRead(f, s.Id)
			if err != nil {
				return err
			}

			err = credential.CopyClipboard(r)
			if err != nil {
				return fmt.Errorf("could not copy to clipboard: %w", err)
			}

			fmt.Printf("The password for `%s` is in the clipboard\n", label)
		}
	}

	return nil
}
