package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/setup"
)

func deleteAction(ctx *cli.Context) error {

	if ctx.Args().Len() == 0 {
		return errors.New("delete command needs one argument (label)")
	}

	label := ctx.Args().First()

	s, err := setupEnv(ctx)
	if err != nil {
		return fmt.Errorf("Unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("Found no privage key file: %w", s.Id.Err)
	}

	labelExists, err := deleteFile(s, label)

	if err != nil {
		return err
	}

	if !labelExists {
		fmt.Printf("Could not find the encrypted file for %s\n", label)
	} else {
		fmt.Printf("Deleted encrypted file for %s\n", label)
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
