package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/credential"
	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// reencryptAction reencrypts modified files
func reencryptAction(ctx *cli.Context) error {

	isForce := ctx.Bool("force")
	isClean := ctx.Bool("clean")

    s, err := setupEnv(ctx)
    if err != nil {
        return fmt.Errorf("Unable to setup environment configuration: %s", err)
	}

	if s.Id.Id == nil {
		return fmt.Errorf("Found no privage key file: %w", s.Id.Err)
	}

	return reencrypt(s, isForce, isClean)
}

// reencrypt reencrypts modified (decrypted) files in the Repository directory.
func reencrypt(s *setup.Setup, isForce, isClean bool) error {

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
		fmt.Println("Found no files to reencrypt.")
		return nil
	}

	// show only, if not force
	if !isForce && !isClean {

		fmt.Println("Found the following files to be reencrypted:")
		logFilesToBeProcessed(toEncrypt)
		fmt.Println("(Use \"privage reencrypt --force\" to reencrypt all decrypted files)")
		fmt.Println("(Use \"privage reencrypt --clean\" to reencrypt all decrypted files and after that delete them)")
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
				return fmt.Errorf("Invalid credential file %s. toml error: %w", h.Label, err)
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

	fmt.Println("The following files were reencrypted:")
	logFilesToBeProcessed(toEncrypt)

	if isClean {
		return clean(s, true)
	}

	return nil
}

func clean(s *setup.Setup, isForce bool) error {
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
		fmt.Println("There are no decrypted files to de deleted.")
		return nil
	}

	if !isForce {

		fmt.Println("The following decrypted files will be deleted because they already exist as encrypted:")
		logFilesToBeProcessed(toClean)

		fmt.Println("Use `privage clean --force` to clean")
		return nil
	}

	for _, h := range toClean {

		// contents as []byte
		err := os.Remove(s.Repository + "/" + h.Label)
		if err != nil {
			return err
		}
	}

	fmt.Println("The following files were deleted:")
	logFilesToBeProcessed(toClean)

	return nil
}

func logFilesToBeProcessed(toEncrypt []*header.Header) {
	fmt.Println()
	for _, h := range toEncrypt {
		fmt.Printf("%8s%s\n", "", h)
	}

	fmt.Println()

	return
}
