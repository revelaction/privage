package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

var ErrHeaderNotFound = errors.New("header not found")

// decryptCommand decrypts an encrypted file
func decryptCommand(opts setup.Options, args []string) error {
	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s decrypt [label]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Decrypt a file and write its content in a file named after the label\n")
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  label  The label of the file to decrypt\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

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

	readContent := func(r io.Reader) (io.Reader, error) {
		return contentReader(r, s.Id)
	}

	createFile := func(name string) (io.WriteCloser, error) {
		return os.Create(filepath.Join(s.Repository, name))
	}

	if err := decrypt(label, streamHeaders, readContent, createFile); err != nil {
		if errors.Is(err, ErrHeaderNotFound) {
			return fmt.Errorf("file %q not found in repository", label)
		}
		return err
	}

	fmt.Printf("The file %s was decrypted in the directory %s.\n", label, s.Repository)
	fmt.Println()
	fmt.Println("(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
	fmt.Println("(Use \"privage reencrypt --clean\" to reencrypt and delete all decrypted files)")

	return nil
}

// decrypt creates a decrypted copy of an encrypted file contents. It saves the
// copy in the repository directory under the file name label
func decrypt(label string, streamHeaders HeaderStreamFunc, readContent ContentReaderFunc, createFile FileCreateFunc) (retErr error) {

	for h := range streamHeaders() {

		if h.Label == label {

			w, err := createFile(label)
			if err != nil {
				return err
			}
			defer func() {
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()

			f, err := os.Open(h.Path)
			if err != nil {
				return err
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && retErr == nil {
					retErr = cerr
				}
			}()

			r, err := readContent(f)
			if err != nil {
				return err
			}

			bufFile := bufio.NewWriter(w)

			if _, err := io.Copy(bufFile, r); err != nil {

				if err != io.ErrUnexpectedEOF {
					return err
				}
			}

			if err := bufFile.Flush(); err != nil {
				return err
			}

			return nil
		}
	}

	return ErrHeaderNotFound
}
