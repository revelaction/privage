package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// decryptCommand decrypts an encrypted file
func decryptCommand(opts setup.Options, args []string) error {

	if len(args) == 0 {
		return errors.New("decrypt command needs one argument (label)")
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	label := args[0]

	streamHeaders := func() <-chan *header.Header {
		return headerGenerator(s.Repository, s.Id)
	}

	openContent := func(h *header.Header) (io.Reader, error) {
		return contentReader(h, s.Id)
	}

	createFile := func(name string) (io.WriteCloser, error) {
		return os.Create(filepath.Join(s.Repository, name))
	}

	return decrypt(label, s.Repository, streamHeaders, openContent, createFile, os.Stdout)
}

// decrypt creates a decrypted copy of an encrypted file contents. It saves the
// copy in the repository directory under the file name label
func decrypt(
	label string,
	repoPath string,
	streamHeaders HeaderStreamFunc,
	openContent ContentOpenFunc,
	createFile FileCreateFunc,
	out io.Writer,
) error {

	for h := range streamHeaders() {

		if h.Label == label {

			w, err := createFile(label)
			if err != nil {
				return err
			}
			defer w.Close()

			r, err := openContent(h)
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

			fmt.Fprintf(out, "The file %s was decrypted in the directory %s.\n", label, repoPath)
			fmt.Fprintln(out)
			fmt.Fprintln(out, "(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
			fmt.Fprintln(out, "(Use \"privage reencrypt --clean\" to reencrypt and delete all decrypted files)")

			return nil
		}
	}

	return nil
}