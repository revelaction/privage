package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// catCommand prints in the terminal the contents of an encrypted file.
func catCommand(opts setup.Options, args []string) error {

	if len(args) == 0 {
		return errors.New("cat command needs one argument (label)")
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

	return cat(label, streamHeaders, openContent, os.Stdout)
}

func cat(label string, streamHeaders HeaderStreamFunc, openContent ContentOpenFunc, out io.Writer) error {

	for h := range streamHeaders() {

		if h.Label == label {

			r, err := openContent(h)
			if err != nil {
				return err
			}

			if _, err := io.Copy(out, r); err != nil {
				if err != io.ErrUnexpectedEOF {
					return err
				}
			}

			return nil
		}
	}

	return nil
}
