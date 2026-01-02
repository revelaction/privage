package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// addCommand adds a new privage encrypted file
//
// the first argument is a category. A category can be anything. There is a
// predefined category (credential) that generates credential files.
// the second one (label) is:
// - a label for credentials
// - a existing file in the current directory
func addCommand(opts setup.Options, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s add [category] [label]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Add a new encrypted file.\n")
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  category  A category (e.g., 'credential' or any custom string)\n")
		fmt.Fprintf(os.Stderr, "  label     A label for credentials, or an existing file path\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if len(args) != 2 {
		return errors.New("add command needs two arguments: <category> <label>")
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	cat := args[0]
	if len(cat) > header.MaxLenghtCategory {
		return errors.New("first argument (category) length is greater than max allowed")
	}

	label := args[1]
	if len(label) > header.MaxLenghtLabel {
		return errors.New("second argument (label) length is greater than max allowed")
	}

	// Check label exists
	if labelExists(label, s.Id) {
		return errors.New("second argument (label) already exist")
	}

	h := &header.Header{Label: label}

	switch cat {
	case header.CategoryCredential:
		h.Category = header.CategoryCredential
		addCredential(h, s)
	default:
		// if custom category, label must be a existing file in the current directory
		if _, err := os.Stat(label); os.IsNotExist(err) {
			return errors.New("second argument (label) must be a existing file in this directory")
		}

		h.Category = cat
		addCustomCategory(h, s)
	}

	return nil
}

// addCredential creates a encrypted credential file in the repository directory.
//
// credential files are toml files
func addCredential(h *header.Header, s *setup.Setup) error {

	cred, err := credential.New(s.C)
	if err != nil {
		return fmt.Errorf("could not create credential: %w", err)
	}

	var buf bytes.Buffer
	if err := cred.Encode(&buf); err != nil {
		return fmt.Errorf("could not encode credential: %w", err)
	}

	err = encryptSave(h, "", &buf, s)
	if err != nil {
		return err
	}

	streamHeaders := func() <-chan *header.Header {
		return headerGenerator(s.Repository, s.Id)
	}

	openContent := func(h *header.Header) (io.Reader, error) {
		return contentReader(h, s.Id)
	}

	show(h.Label, "", streamHeaders, openContent, os.Stdout)

	fmt.Println("You can edit the credentials file by running these commands:")
	fmt.Println()
	fmt.Printf("   privage decrypt %s\n", h.Label)
	fmt.Printf("   vim %s # or your favorite editor\n", h.Label)
	fmt.Println("   privage reencrypt")
	fmt.Println()

	return nil
}

// addCustomCategory creates an encrypted file of the contents of a file
// present in the repository directory
func addCustomCategory(h *header.Header, s *setup.Setup) error {

	content, err := os.Open(h.Label)
	if err != nil {
		return err
	}
	defer content.Close()

	err = encryptSave(h, "", content, s)
	if err != nil {
		return err
	}

	return nil
}

func labelExists(label string, identity id.Identity) bool {
	labels := map[string]struct{}{}

	for h := range headerGenerator(".", identity) {
		if _, ok := labels[h.Label]; !ok {
			labels[h.Label] = struct{}{}
		}
	}

	for k := range labels {
		if k == label {
			return true
		}
	}

	return false
}
