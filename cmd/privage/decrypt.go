package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/setup"
)

// decryptAction decrypts an encrypted file
func decryptAction(args []string) error {

	if len(args) == 0 {
		return errors.New("decrypt command needs one argument (label)")
	}

	s, err := setupEnv(global)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]

	return decrypt(label, s)
}

func decrypt(label string, s *setup.Setup) error {

	for h := range headerGenerator(s.Repository, s.Id) {

		if h.Label == label {

			r, err := contentReader(h, s.Id)
			if err != nil {
				return err
			}

			f, err := os.Create(label)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, r)
			if err != nil {
				return err
			}

			fmt.Printf("The file %s was decrypted in the directory %s.\n\n", label, s.Repository)

			fmt.Println("(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
			fmt.Println("(Use \"privage reencrypt --clean\" to reencrypt all decrypted files and after that delete them)")

			return nil
		}
	}

	return nil
}
