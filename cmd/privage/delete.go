package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/revelaction/privage/setup"
)

func deleteCommand(opts setup.Options, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s delete [label]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Delete an encrypted file.\n")
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  label  The label of the file to delete\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) == 0 {
		return errors.New("delete command needs one argument (label)")
	}

	label := args[0]

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	labelExists, err := deleteFile(s, label)

	if err != nil {
		return err
	}

	if !labelExists {
		fmt.Printf("could not find the encrypted file for %s\n", label)
	} else {
		fmt.Printf("deleted encrypted file for %s\n", label)
	}

	return nil
}

// deleteFile deletes an encrypted file.
func deleteFile(s *setup.Setup, label string) (bool, error) {

	labelExists := false
	for h := range headerGenerator(s.Repository, s.Id) {

		if h.Label == label {

			filePath := s.Repository + "/" + fileName(h, s.Id, "")
			err := os.Remove(filePath)
			if err != nil {
				return false, err
			}

			labelExists = true
			break
		}
	}

	return labelExists, nil
}
