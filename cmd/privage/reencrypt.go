package main

import (
	"fmt"
	"os"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// reencryptCommand reencrypts modified files
func reencryptCommand(s *setup.Setup, isForce, isClean bool, ui UI) error {
	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	return reencrypt(s, isForce, isClean, ui)
}

// reencrypt reencrypts modified (decrypted) files in the Repository directory.
func reencrypt(s *setup.Setup, isForce, isClean bool, ui UI) error {

	headers := []*header.Header{}

	for h := range headerGenerator(s.Repository, s.Id) {
		headers = append(headers, h)
	}

	toEncrypt := []*header.Header{}
	for _, h := range headers {
		//if label exist as file add to list to encrypt
		if _, err := os.Stat(s.Repository + "/" + h.Label); !os.IsNotExist(err) {
			toEncrypt = append(toEncrypt, h)
		}
	}

	if len(toEncrypt) == 0 {
		_, _ = fmt.Fprintln(ui.Err, "Found no files to reencrypt.")
		return nil
	}

	// show only, if not force
	if !isForce && !isClean {

		_, _ = fmt.Fprintln(ui.Err, "Found the following files to be reencrypted:")
		logFilesToBeProcessed(toEncrypt, ui)
		_, _ = fmt.Fprintln(ui.Err, "(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
		_, _ = fmt.Fprintln(ui.Err, "(Use \"privage reencrypt --clean\" to reencrypt and delete all decrypted files)")
		return nil
	}

	for _, h := range toEncrypt {

		f, err := os.Open(s.Repository + "/" + h.Label)
			if err != nil {
				return err
			}

		// if is credential category -> validate as toml
		if header.CategoryCredential == h.Category {
			err := credential.ValidateFile(s.Repository + "/" + h.Label)
				if err != nil {
					return fmt.Errorf("invalid credential file %s. toml error: %w", h.Label, err)
				}
		}

		//encrypt and save the file
		err = encryptSave(h, "", f, s)
			if err != nil {
				return err
			}

			if err := f.Close(); err != nil {
				return err
			}
	}

	_, _ = fmt.Fprintln(ui.Err, "The following files were reencrypted:")
	logFilesToBeProcessed(toEncrypt, ui)

	if isClean {
		return clean(s, true, ui)
	}

	return nil
}

func clean(s *setup.Setup, isForce bool, ui UI) error {
	headers := []*header.Header{}

	for h := range headerGenerator(s.Repository, s.Id) {
		headers = append(headers, h)
	}

	toClean := []*header.Header{}
	for _, h := range headers {

		//if label exist as file, then add to list to encrypt
		if _, err := os.Stat(s.Repository + "/" + h.Label); !os.IsNotExist(err) {
			toClean = append(toClean, h)
		}
	}

	if len(toClean) == 0 {
		_, _ = fmt.Fprintln(ui.Err, "There are no decrypted files to de deleted.")
		return nil
	}

	if !isForce {

		_, _ = fmt.Fprintln(ui.Err, "The following decrypted files will be deleted because they already exist as encrypted:")
		logFilesToBeProcessed(toClean, ui)

		_, _ = fmt.Fprintln(ui.Err, "Use `privage clean --force` to clean")
		return nil
	}

	for _, h := range toClean {

		// contents as []byte
		err := os.Remove(s.Repository + "/" + h.Label)
			if err != nil {
				return err
			}
	}

	_, _ = fmt.Fprintln(ui.Err, "The following files were deleted:")
	logFilesToBeProcessed(toClean, ui)

	return nil
}

func logFilesToBeProcessed(toEncrypt []*header.Header, ui UI) {
	_, _ = fmt.Fprintln(ui.Err)
	for _, h := range toEncrypt {
		_, _ = fmt.Fprintf(ui.Err, "%8s%s\n", "", h)
	}

	_, _ = fmt.Fprintln(ui.Err)
}
