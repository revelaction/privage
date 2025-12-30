package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// addAction adds a new privage encrypted file
//
// the first argument is a category. A category can be anything. There is a
// predefined category (credential) that generates credential files.
// the second one (label) is:
// - a label for credentials
// - a existing file in the current directory
func addAction(ctx *cli.Context) error {

	s, err := setupEnv(ctx)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	if ctx.Args().Len() != 2 {
		return errors.New("usage <category> <label>")
	}

	cat := ctx.Args().First()
	if len(cat) > header.MaxLenghtCategory {
		return errors.New("first argument (category) length is greater than max allowed")
	}

	label := ctx.Args().Get(1)
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

	password, err := credential.GeneratePassword()
	if err != nil {
		return fmt.Errorf("could not generate password: %w", err)
	}

	// login
	var login string
	if len(s.C.Login) > 0 {
		login = s.C.Login
	} else if len(s.C.Email) > 0 {
		login = s.C.Email
	} else {
		login = ""
	}

	// email
	var email string
	if len(s.C.Email) > 0 {
		email = s.C.Email
	} else {
		email = ""
	}

	content := fmt.Sprintf(credential.Template, login, password, email)
	r := bytes.NewReader([]byte(content))

	err = encryptSave(h, "", r, s)
	if err != nil {
		return err
	}

	show(h.Label, s)

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
