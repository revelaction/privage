package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/revelaction/privage/setup"
)

// deleteCommand deletes an encrypted file from the repository
func deleteCommand(s *setup.Setup, args []string, ui UI) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(ui.Err)
	fs.Usage = func() {
		fmt.Fprintf(ui.Err, "Usage: %s delete [label]\n", os.Args[0])
		fmt.Fprintf(ui.Err, "\nDescription:\n")
		fmt.Fprintf(ui.Err, "  Delete an encrypted file.\n")
		fmt.Fprintf(ui.Err, "\nArguments:\n")
		fmt.Fprintf(ui.Err, "  label  The label of the file to delete\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) == 0 {
		return fmt.Errorf("delete command needs one argument (label)")
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]

	found := false
	for h := range headerGenerator(s.Repository, s.Id) {
		if h.Label == label {
			err := os.Remove(h.Path)
			if err != nil {
				return err
			}
			found = true
			break
		}
	}

	if !found {
		fmt.Fprintf(ui.Err, "could not find the encrypted file for %s\n", label)
	} else {
		fmt.Fprintf(ui.Err, "deleted encrypted file for %s\n", label)
	}

	return nil
}