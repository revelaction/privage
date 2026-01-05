package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// showCommand prints in the terminal partially/all the contents of an encrypted
// file.
func showCommand(s *setup.Setup, args []string, ui UI) (err error) {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(ui.Err)
	fs.Usage = func() {
		fmt.Fprintf(ui.Err, "Usage: %s show [label] [field]\n", os.Args[0])
		fmt.Fprintf(ui.Err, "\nDescription:\n")
		fmt.Fprintf(ui.Err, "  Show the contents of an encrypted file (formatted if it's a credential).\n")
		fmt.Fprintf(ui.Err, "  If a field name is provided, only that field's value is printed.\n")
		fmt.Fprintf(ui.Err, "\nArguments:\n")
		fmt.Fprintf(ui.Err, "  label  The label of the file to show\n")
		fmt.Fprintf(ui.Err, "  field  Optional: specific TOML field to show (e.g., api_key)\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) == 0 {
		return errors.New("show command needs at least one argument (label)")
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]
	var fieldName string
	if len(args) > 1 {
		fieldName = args[1]
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

	return nil
}