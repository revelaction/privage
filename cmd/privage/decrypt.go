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
	
	var ErrHeaderNotFound = errors.New("header not found")
	
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
	
		if err := decrypt(label, streamHeaders, openContent, createFile); err != nil {
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
	func decrypt(label string, streamHeaders HeaderStreamFunc, openContent ContentOpenFunc, createFile FileCreateFunc) (retErr error) {
	
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
	
				if err := bufFile.Flush(); err != nil {
					return err
				}
	
				return nil
			}
		}
	
		return ErrHeaderNotFound
	}
	
