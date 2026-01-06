package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// addCommand is a pure logic worker for adding encrypted files.
// It assumes that category and label have been validated by the driver in main.go.
func addCommand(s *setup.Setup, cat string, label string, ui UI) error {

	// Check label exists
	if labelExists(label, s.Id) {
		return fmt.Errorf("second argument (label) %q already exist", label)
	}

	h := &header.Header{Label: label}

	switch cat {
	case header.CategoryCredential:
		h.Category = header.CategoryCredential
		if err := addCredential(h, s, ui); err != nil {
			return err
		}
	default:
		h.Category = cat
		if err := addCustomCategory(h, s, ui); err != nil {
			return err
		}
	}

	return nil
}

// addCredential creates a encrypted credential file in the repository directory.
func addCredential(h *header.Header, s *setup.Setup, ui UI) error {

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

	// Data output goes to ui.Out (handled by showCommand)
	if err := showCommand(s, h.Label, "", ui); err != nil {
		return err
	}

	// Diagnostic/Instruction output goes to ui.Err
	fmt.Fprintln(ui.Err, "You can edit the credentials file by running these commands:")
	fmt.Fprintln(ui.Err)
	fmt.Fprintf(ui.Err, "   privage decrypt %s\n", h.Label)
	fmt.Fprintf(ui.Err, "   vim %s # or your favorite editor\n", h.Label)
	fmt.Fprintf(ui.Err, "   privage reencrypt\n")
	fmt.Fprintln(ui.Err)

	return nil
}

// addCustomCategory creates an encrypted file of the contents of a file
// present in the repository directory
func addCustomCategory(h *header.Header, s *setup.Setup, ui UI) (err error) {

	content, err := os.Open(h.Label)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := content.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	err = encryptSave(h, "", content, s)
	if err != nil {
		return err
	}

	fmt.Fprintf(ui.Err, "Added file '%s' to category '%s' ✔️\n", h.Label, h.Category)

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
