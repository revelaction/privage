package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/config"
	id "github.com/revelaction/privage/identity"
)

const (
	GITIGNORE = `# Ignore everything
*

# But not these files...
!.gitignore
!*.age`
)

// initAction generates an age identity in the current dir if no identity was found.
// It generates a .gitignore file in the current directory if not existing.
// It generates a .privage.conf file in the home directory, with the
// identity and secret directory paths.
func initAction(ctx *cli.Context) error {

	s, err := setupEnv(ctx)
	if err != nil {
		return fmt.Errorf("Unable to setup environment configuration: %s", err)
	}

	// If config file exist, we exit
	if len(s.C.Path) > 0 {
		fmt.Printf("📑 Config file already exists: %s... Exiting\n", s.C.Path)
		return nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	//
	// identity
	//
	if len(s.Id.Path) > 0 {
		fmt.Printf("🔑 privage key file already exists: %s... Exiting.\n", s.Id.Path)
		return nil
	}

	identityPath := currentDir + "/" + id.FileName
	identityType := id.TypeAge
	identityAlgo := id.PivAlgoRsa2048 // only RSA2048 supported

	// we receive string hex representation like 9a
	slot := ctx.String("piv-slot")

	// piv key
	if len(slot) > 0 {
		identityType = id.TypePiv

		identitySlot, err := strconv.ParseUint(slot, 16, 32)
		if err != nil {
			return fmt.Errorf("Could not convert slot %s to hex: %v", slot, err)
		}

		err = id.CreatePivRsa(identityPath, uint32(identitySlot), identityAlgo)
		if err != nil {
			return fmt.Errorf("Error creating encrypted age key in slot %s: %w", slot, err)
		}

		fmt.Printf("🔑 Generated encrypted age key file `%s` with PIV slot %s ✔️\n", identityPath, slot)
	} else {
		// normal age key
		err = id.Create(identityPath)
		if err != nil {
			return err
		}

		fmt.Printf("🔑 Generated age key file `%s` ✔️\n", identityPath)
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

		fmt.Printf("📒 .gitignore file already exists: %s... Exiting\n", gitignorePath)
		return nil
	}

	fmt.Printf("📒 Generated `%s` file ✔️\n", gitignorePath)

	//
	// config file
	//
	err = config.Create(identityPath, identityType, slot, currentDir)
	if err != nil {
		return fmt.Errorf("Could not generate config file: %w", err)
	}

	fmt.Printf("📑 Generated config file %s ✔️\n", config.FileName)

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
