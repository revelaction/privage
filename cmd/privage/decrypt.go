package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/setup"
)

// decryptAction decrypts an age encrypted file.
func decryptAction(ctx *cli.Context) error {
	label := ctx.Args().First()

	s, err := setupEnv(ctx)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	return decrypt(s, label)
}

// decrypt creates a decrypted copy of an encrypted file contents. It saves the
// copy in the repository directory under the file name label
func decrypt(s *setup.Setup, label string) error {

	for h := range headerGenerator(s.Repository, s.Id) {

		if h.Label == label {

			w, err := os.Create(s.Repository + "/" + label)
			if err != nil {
				return err
			}
			defer w.Close()

			r, err := contentReader(h, s.Id)
			if err != nil {
				return err
			}

			bufFile := bufio.NewWriter(w)

			if _, err := io.Copy(bufFile, r); err != nil {

				if err != io.ErrUnexpectedEOF {
					return err
				}
			}

			bufFile.Flush()

			fmt.Printf("The file %s was decrypted in the directory %s.\n", label, s.Repository)
			fmt.Println()
			fmt.Println("(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
			fmt.Println("(Use \"privage reencrypt --clean\" to reencrypt and delete all decrypted files)")

			return nil
		}
	}

	return nil
}
