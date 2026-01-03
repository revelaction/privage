package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/revelaction/privage/config"
	filesystem "github.com/revelaction/privage/fs"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

const (
	GITIGNORE = `# Ignore everything
*

# But not these files...
!.gitignore
!*.age`
)

// initCommand generates an age identity in the current dir if no identity was found.
// It generates a .gitignore file in the current directory if not existing.
// It generates a .privage.conf file in the home directory, with the
// identity and secret directory paths.
func initCommand(opts setup.Options, args []string) (err error) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	var slot string
	fs.StringVar(&slot, "piv-slot", "", "Use the yubikey slot key to encrypt the age private key")
	fs.StringVar(&slot, "p", "", "alias for -piv-slot")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s init [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Add a .gitignore, age/yubikey key file to the current directory. Add a config file in the home directory.\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
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

	identityPath := currentDir + "/" + id.DefaultFileName
	identityType := id.TypeAge
	identityAlgo := id.PivAlgoRsa2048 // only RSA2048 supported

	// piv key
	if len(slot) > 0 {
		identityType = id.TypePiv

		identitySlot, err := strconv.ParseUint(slot, 16, 32)
		if err != nil {
			return fmt.Errorf("could not convert slot %s to hex: %v", slot, err)
		}

		err = id.CreatePivRsa(identityPath, uint32(identitySlot), identityAlgo)
		if err != nil {
			return fmt.Errorf("error creating encrypted age key in slot %s: %w", slot, err)
		}

		fmt.Printf("🔑 Generated encrypted age key file `%s` with PIV slot %s ✔️\n", identityPath, slot)
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
		err = id.New(f)
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	confPath := homeDir + "/" + config.DefaultFileName
	f, err := os.OpenFile(confPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("could not create config file: %w", err)
	}
	defer f.Close()

	conf := &config.Config{
		IdentityPath:    identityPath,
		IdentityType:    identityType,
		IdentityPivSlot: slot,
		RepositoryPath:  currentDir,
	}

	if err := conf.Encode(f); err != nil {
		return fmt.Errorf("could not encode config file: %w", err)
	}

	fmt.Printf("📑 Generated config file %s ✔️\n", config.DefaultFileName)

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
