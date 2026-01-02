package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"filippo.io/age"

	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

const (
	suffix         = ".rotate"
	fileNameRotate = "privage-key-rotate.txt"
)

// rotateCommand generates a new age key and reencrypts all present encrypted
// fields with the new key.
func rotateCommand(opts setup.Options, args []string) error {
	fs := flag.NewFlagSet("rotate", flag.ContinueOnError)
	var isClean bool
	var slot string
	fs.BoolVar(&isClean, "clean", false, "Delete old Key's encrypted files. Rename new encrypted files and the new key")
	fs.BoolVar(&isClean, "c", false, "alias for -clean")
	fs.StringVar(&slot, "piv-slot", "", "Use the yubikey slot to encrypt the age private key with the RSA Key")
	fs.StringVar(&slot, "p", "", "alias for -piv-slot")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s rotate [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Create a new age key and reencrypt every file with the new key.\n")
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

	return rotate(s, isClean, slot)
}

func rotate(s *setup.Setup, isClean bool, slot string) error {

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	numFiles := numFilesForIdentity(s.Repository, s.Id)
	if numFiles == 0 {
		return fmt.Errorf("found no encrypted files with key %s", s.Id.Path)
	}

	fmt.Printf("Found %d files encrypted with key %s\n", numFiles, s.Id.Path)
	fmt.Println()

	var idRotate id.Identity
	idRotatePath := s.Repository + "/" + fileNameRotate
	var pivSlot uint32
	var err error

	if len(slot) > 0 {
		ps, err := strconv.ParseUint(slot, 16, 32)
		if err != nil {
			return fmt.Errorf("could not convert slot %s to hex: %v", slot, err)
		}

		pivSlot = uint32(ps)
	}

	if pivSlot > 0 {
		idRotate = id.LoadPiv(idRotatePath, pivSlot, "")
	} else {
		idRotate = id.Load(idRotatePath)
	}

	numFilesRotate := 0
	if idRotate.Err == nil {
		numFilesRotate = numFilesForIdentity(s.Repository, idRotate)
		fmt.Printf("Found %d files encrypted with the rotated key %s\n", numFilesRotate, idRotate.Path)

		if numFiles == numFilesRotate {

			// the rotate process is completed. Run clean if flag
			if isClean {
				err = cleanRotate(s, idRotate, slot)
				if err != nil {
					return err
				}

				return nil
			}

			fmt.Println()
			fmt.Println("rotate is completed ✔️")
			fmt.Println("(Use \"privage rotate --clean\" to clean up old encrypted files and rename the new ones.)")
			return nil
		}
	}

	if idRotate.Err != nil {

		// Create
		if pivSlot > 0 {
			err = id.CreatePivRsa(idRotatePath, pivSlot, id.PivAlgoRsa2048)
		} else {
			err = id.Create(idRotatePath)
		}

		if err != nil {
			return fmt.Errorf("could not create age key file: %w", err)
		}

		// Load
		if pivSlot > 0 {
			idRotate = id.LoadPiv(idRotatePath, pivSlot, "")
		} else {
			idRotate = id.Load(idRotatePath)
		}

		fmt.Printf("🔑 Created new age key file %s✔️\n", idRotate.Path)
	}

	// iterate all encrypted files with the current key and reencrypt.
	//
	// maybe we are in a rerun of the command rotate, after a failing process.
	// some age files present in the repo will be encrypted with the old key and
	// some with the new one.
	numReencrypted := 0
	for h := range headerGenerator(s.Repository, s.Id) {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if errors.As(h.Err, &e) {
				continue
			}

			return h.Err
		}

		r, err := contentReader(h, s.Id)
		if err != nil {

			return err
		}

		// Need a setup with the repo and the new idRotate
		sRotate := s.Copy()
		sRotate.Id = idRotate
		err = encryptSave(h, suffix, r, sRotate)
		if err != nil {
			return err
		}
		numReencrypted++
	}

	fmt.Printf("🔐  Reencrypted %d files with new key %s\n", numReencrypted, idRotate.Path)
	fmt.Println()
	fmt.Println("rotate is completed ✔️")

	if !isClean {
		fmt.Println("(Use \"privage rotate --clean\" to clean up old encrypted files and rename the new ones.)")
		return nil
	}

	err = cleanRotate(s, idRotate, slot)
	if err != nil {
		return err
	}

	return nil
}

// cleanRotate removes all encrypted files with the old key
// it also renames all age encrypted files of new key to standard form (without rotated suffix)
// it also renames the keys (old and new).
func cleanRotate(s *setup.Setup, idRotate id.Identity, slot string) error {

	fmt.Println()
	fmt.Println("Cleaning files...")
	fmt.Println()

	// 1) remove all age encrypted fields of old key
	numDeleted := 0
	for h := range headerGenerator(s.Repository, s.Id) {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if !errors.As(h.Err, &e) {
				return h.Err
			}

			continue
		}

		err := os.Remove(h.Path)
		if err != nil {
			fmt.Printf("%8s Error while deleting h.Path %s: %s\n", "", h.Path, err)
			return err
		}

		numDeleted++
	}

	fmt.Printf("Deleted %d files encrypted with key %s\n", numDeleted, s.Id.Path)
	fmt.Println()

	// 2) rename
	numRenamed := 0
	for h := range headerGenerator(s.Repository, idRotate) {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if errors.As(h.Err, &e) {
				fmt.Printf("NoIdentityMatchError: Could not decrupt with curreent key %s\n", h.Err)
				continue
			}

			return h.Err
		}

		stardardPath := strings.TrimSuffix(h.Path, suffix+AgeExtension) + AgeExtension
		err := os.Rename(h.Path, stardardPath)
		if err != nil {
			return err
		}

		numRenamed++
	}

	fmt.Printf("Renamed %d rotated files to %s extension\n", numRenamed, AgeExtension)
	fmt.Println()

	// 3) rename old key to back
	backupKeyPath := id.BackupFilePath(s.Repository)
	err := os.Rename(s.Id.Path, backupKeyPath)
	if err != nil {
		return err
	}

	fmt.Printf("Renamed old key %s to %s\n", s.Id.Path, backupKeyPath)
	fmt.Println()

	// 4) rename new key
	err = os.Rename(idRotate.Path, s.Id.Path)
	if err != nil {
		return err
	}
	fmt.Printf("Renamed new key %s to %s\n", idRotate.Path, s.Id.Path)
	fmt.Println()

	fmt.Printf("The new key is a %s\n", id.FmtType(slot))
	fmt.Printf("⚠ Make sure the config file %s has these lines:\n", s.C.Path)
	fmt.Println()
	if len(slot) > 0 {
		fmt.Println("    identity_type = \"PIV\"")
		fmt.Printf("    identity_piv_slot = \"%s\"\n", slot)
	} else {
		fmt.Println("    identity_type = \"\"")
		fmt.Println("    identity_piv_slot = \"\"")
	}
	fmt.Println()

	fmt.Println("cleaning is completed ✔️")

	return nil
}

func numFilesForIdentity(repoDir string, identity id.Identity) int {
	num := 0
	for h := range headerGenerator(repoDir, identity) {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if errors.As(h.Err, &e) {
				continue
			}
		}

		num++
	}

	return num
}
