package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/revelaction/privage/config"
	filesystem "github.com/revelaction/privage/fs"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/identity/piv/yubikey"
)

const (
	GITIGNORE = `# Ignore everything
*

# But not these files...
!.gitignore
!*.privage`
)

// initCommand is a pure logic worker for environment initialization.
// It generates an age identity, a .gitignore, and a .privage.conf file.
func initCommand(slot string, ui UI) (err error) {

	// Pre-flight checks
	configPath, err := filesystem.FindConfigFile()
	if err != nil {
		return fmt.Errorf("error searching for config file: %w", err)
	}
	if configPath != "" {
		_, _ = fmt.Fprintf(ui.Err, "📑 Config file already exists: %s... Exiting\n", configPath)
		return nil
	}

	idPath, err := filesystem.FindIdentityFile()
	if err != nil {
		return fmt.Errorf("error searching for identity file: %w", err)
	}
	if idPath != "" {
		_, _ = fmt.Fprintf(ui.Err, "🔑 privage key file already exists: %s... Exiting.\n", idPath)
		return nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	//
	// identity
	//
	identityPath := currentDir + "/" + id.DefaultFileName
	identityType := id.TypeAge

	// piv key
	if len(slot) > 0 {
		identityType = id.TypePiv

		identitySlot, err := strconv.ParseUint(slot, 16, 32)
		if err != nil {
			return fmt.Errorf("could not convert slot %s to hex: %v", slot, err)
		}

		yk, err := yubikey.New()
		if err != nil {
			return fmt.Errorf("could not connect to yubikey: %w", err)
		}
		defer func() {
			if cerr := yk.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		f, err := filesystem.CreateFile(identityPath, 0600)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		err = id.GeneratePiv(f, yk, uint32(identitySlot))
		if err != nil {
			return fmt.Errorf("error creating encrypted age key in slot %s: %w", slot, err)
		}

		_, _ = fmt.Fprintf(ui.Err, "🔑 Generated encrypted age key file `%s` with PIV slot %s ✔️\n", identityPath, slot)
	} else {
		// normal age key
		f, err := filesystem.CreateFile(identityPath, 0600)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()
		err = id.GenerateAge(f)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(ui.Err, "🔑 Generated age key file `%s` ✔️\n", identityPath)
	}

	//
	// gitignore
	//
	// if existing gitignore do not generate
	gitignorePath := currentDir + "/.gitignore"
	err = generateGitignore(gitignorePath)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}

		_, _ = fmt.Fprintf(ui.Err, "📒 .gitignore file already exists: %s... Exiting\n", gitignorePath)
		return nil
	}

	_, _ = fmt.Fprintf(ui.Err, "📒 Generated `%s` file ✔️\n", gitignorePath)

	//
	// config file
	//
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	confPath := homeDir + "/" + config.DefaultFileName
	f, err := os.OpenFile(confPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("could not create config file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	conf := &config.Config{
		IdentityPath:    identityPath,
		IdentityType:    identityType,
		IdentityPivSlot: slot,
		RepositoryPath:  currentDir,
	}

	if err := conf.Encode(f); err != nil {
		return fmt.Errorf("could not encode config file: %w", err)
	}

	_, _ = fmt.Fprintf(ui.Err, "📑 Generated config file %s ✔️\n", confPath)

	return nil
}

func generateGitignore(path string) error {
	//if file exists, error ErrExist
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()

	_, err = f.Write([]byte(GITIGNORE))
	if err != nil {
		return err
	}

	return nil
}
