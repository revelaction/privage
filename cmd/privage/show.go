package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// showCommand prints in the terminal partially/all the contents of an encrypted
// file.
func showCommand(opts setup.Options, args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s show [label] [field]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Show the contents of an encrypted file (formatted if it's a credential).\n")
		fmt.Fprintf(os.Stderr, "  If a field name is provided, only that field's value is printed.\n")
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  label  The label of the file to show\n")
		fmt.Fprintf(os.Stderr, "  field  Optional: specific TOML field to show (e.g., api_key)\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) == 0 {
		return errors.New("show command needs at least one argument (label)")
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]
	var fieldName string
	if len(args) > 1 {
		fieldName = args[1]
	}

	streamHeaders := func() <-chan *header.Header {
		return headerGenerator(s.Repository, s.Id)
	}

	openContent := func(h *header.Header) (io.Reader, error) {
		return contentReader(h, s.Id)
	}

	return show(label, fieldName, streamHeaders, openContent, os.Stdout)
}

func show(label string, fieldName string, streamHeaders HeaderStreamFunc, openContent ContentOpenFunc, out io.Writer) error {

	for h := range streamHeaders() {

		if h.Label == label {

			r, err := openContent(h)
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
				fmt.Fprint(out, val)
				return nil
			}

			return cred.FprintBasic(out)
		}
	}

	return nil
}
