package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"filippo.io/age"

	"github.com/revelaction/privage/fs"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/identity/piv/yubikey"
	"github.com/revelaction/privage/setup"
)

const (
	suffix         = ".rotate"
	fileNameRotate = "privage-key-rotate.txt"
)

// rotateCommand generates a new age key and reencrypts all present encrypted
// fields with the new key.
func rotateCommand(s *setup.Setup, args []string, ui UI) (err error) {
	fs := flag.NewFlagSet("rotate", flag.ContinueOnError)
	fs.SetOutput(ui.Err)
	var isClean bool
	var slot string
	fs.BoolVar(&isClean, "clean", false, "Delete old Key's encrypted files. Rename new encrypted files and the new key")
	fs.BoolVar(&isClean, "c", false, "alias for -clean")
	fs.StringVar(&slot, "piv-slot", "", "Use the yubikey slot to encrypt the age private key with the RSA Key")
	fs.StringVar(&slot, "p", "", "alias for -piv-slot")
	fs.Usage = func() {
		fmt.Fprintf(ui.Err, "Usage: %s rotate [options]\n", os.Args[0])
		fmt.Fprintf(ui.Err, "\nDescription:\n")
		fmt.Fprintf(ui.Err, "  Create a new age key and reencrypt every file with the new key.\n")
		fmt.Fprintf(ui.Err, "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err = fs.Parse(args); err != nil {
		return err
	}

	return rotate(s, isClean, slot, ui)
}

func rotate(s *setup.Setup, isClean bool, slot string, ui UI) (err error) {

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	numFiles := numFilesForIdentity(s.Repository, s.Id)
	if numFiles == 0 {
		return fmt.Errorf("found no encrypted files with key %s", s.Id.Path)
	}

	fmt.Fprintf(ui.Err, "Found %d files encrypted with key %s\n", numFiles, s.Id.Path)
	fmt.Fprintln(ui.Err)

	var idRotate id.Identity
	idRotatePath := s.Repository + "/" + fileNameRotate
	var pivSlot uint32

	if len(slot) > 0 {
		ps, err := strconv.ParseUint(slot, 16, 32)
		if err != nil {
			return fmt.Errorf("could not convert slot %s to hex: %v", slot, err)
		}

		pivSlot = uint32(ps)
	}

	if pivSlot > 0 {
		device, err := yubikey.New()
		if err != nil {
			return fmt.Errorf("could not create yubikey device: %w", err)
		}
		defer func() {
			if cerr := device.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		f, err := fs.OpenFile(idRotatePath)
		if err != nil {
			return fmt.Errorf("could not open key file %s: %w", idRotatePath, err)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		idRotate = id.LoadPiv(f, idRotatePath, device, pivSlot)
	} else {
		f, err := fs.OpenFile(idRotatePath)
		if err != nil {
			return fmt.Errorf("could not open key file %s: %w", idRotatePath, err)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()
		idRotate = id.LoadAge(f, idRotatePath)
	}

	numFilesRotate := 0
	if idRotate.Err == nil {
		numFilesRotate = numFilesForIdentity(s.Repository, idRotate)
		fmt.Fprintf(ui.Err, "Found %d files encrypted with the rotated key %s\n", numFilesRotate, idRotate.Path)

		if numFiles == numFilesRotate {

			// the rotate process is completed. Run clean if flag
			if isClean {
				err = cleanRotate(s, idRotate, slot, ui)
				if err != nil {
					return err
				}

				return nil
			}

			fmt.Fprintln(ui.Err)
			fmt.Fprintln(ui.Err, "rotate is completed ✔️")
			fmt.Fprintln(ui.Err, "(Use \"privage rotate --clean\" to clean up old encrypted files and rename the new ones.)")
			return nil
		}
	}

	if idRotate.Err != nil {

		// Create
		if pivSlot > 0 {
			f, err := fs.CreateFile(idRotatePath, 0600)
			if err != nil {
				return fmt.Errorf("could not create key file %s: %w", idRotatePath, err)
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()

			device, err := yubikey.New()
			if err != nil {
				return fmt.Errorf("could not create yubikey device: %w", err)
			}
			defer func() {
				if cerr := device.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()

			err = id.GeneratePiv(f, device, pivSlot)
		} else {
			var f io.WriteCloser
			f, err = fs.CreateFile(idRotatePath, 0600)
			if err != nil {
				return fmt.Errorf("could not create key file %s: %w", idRotatePath, err)
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()
			err = id.GenerateAge(f)
		}

		if err != nil {
			return fmt.Errorf("could not create age key file: %w", err)
		}

		// Load
		if pivSlot > 0 {
			device, err := yubikey.New()
			if err != nil {
				return fmt.Errorf("could not create yubikey device: %w", err)
			}
			defer func() {
				if cerr := device.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()

			f, err := fs.OpenFile(idRotatePath)
			if err != nil {
				return fmt.Errorf("could not open key file %s: %w", idRotatePath, err)
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()

			idRotate = id.LoadPiv(f, idRotatePath, device, pivSlot)
		} else {
			f, err := fs.OpenFile(idRotatePath)
			if err != nil {
				return fmt.Errorf("could not open key file %s: %w", idRotatePath, err)
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()
			idRotate = id.LoadAge(f, idRotatePath)
		}

		fmt.Fprintf(ui.Err, "🔑 Created new age key file %s✔️\n", idRotate.Path)
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

		err = func() (err error) {
			f, err := os.Open(h.Path)
			if err != nil {
				return err
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && err == nil {
					err = cerr
				}
			}()

			r, err := contentReader(f, s.Id)
			if err != nil {
				return err
			}

			// Need a setup with the repo and the new idRotate
			sRotate := s.Copy()
			sRotate.Id = idRotate
			return encryptSave(h, suffix, r, sRotate)
		}()

		if err != nil {
			return err
		}
		numReencrypted++
	}

	fmt.Fprintf(ui.Err, "🔐  Reencrypted %d files with new key %s\n", numReencrypted, idRotate.Path)
	fmt.Fprintln(ui.Err)
	fmt.Fprintln(ui.Err, "rotate is completed ✔️")

	if !isClean {
		fmt.Fprintln(ui.Err, "(Use \"privage rotate --clean\" to clean up old encrypted files and rename the new ones.)")
		return nil
	}

	err = cleanRotate(s, idRotate, slot, ui)
	if err != nil {
		return err
	}

	return nil
}

// cleanRotate removes all encrypted files with the old key
// it also renames all age encrypted files of new key to standard form (without rotated suffix)
// it also renames the keys (old and new).
func cleanRotate(s *setup.Setup, idRotate id.Identity, slot string, ui UI) error {

	fmt.Fprintln(ui.Err)
	fmt.Fprintln(ui.Err, "Cleaning files...")
	fmt.Fprintln(ui.Err)

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
			fmt.Fprintf(ui.Err, "%8s Error while deleting h.Path %s: %s\n", "", h.Path, err)
			return err
		}

		numDeleted++
	}

	fmt.Fprintf(ui.Err, "Deleted %d files encrypted with key %s\n", numDeleted, s.Id.Path)
	fmt.Fprintln(ui.Err)

	// 2) rename
	numRenamed := 0
	for h := range headerGenerator(s.Repository, idRotate) {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if errors.As(h.Err, &e) {
				fmt.Fprintf(ui.Err, "NoIdentityMatchError: Could not decrupt with curreent key %s\n", h.Err)
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

	fmt.Fprintf(ui.Err, "Renamed %d rotated files to %s extension\n", numRenamed, AgeExtension)
	fmt.Fprintln(ui.Err)

	// 3) rename old key to back
	backupKeyPath := id.BackupFilePath(s.Repository)
	err := os.Rename(s.Id.Path, backupKeyPath)
	if err != nil {
		return err
	}

	fmt.Fprintf(ui.Err, "Renamed old key %s to %s\n", s.Id.Path, backupKeyPath)
	fmt.Fprintln(ui.Err)

	// 4) rename new key
	err = os.Rename(idRotate.Path, s.Id.Path)
	if err != nil {
		return err
	}
	fmt.Fprintf(ui.Err, "Renamed new key %s to %s\n", idRotate.Path, s.Id.Path)
	fmt.Fprintln(ui.Err)

	fmt.Fprintf(ui.Err, "The new key is a %s\n", id.FmtType(slot))
	fmt.Fprintf(ui.Err, "⚠ Make sure the config file %s has these lines:\n", s.C.Path)
	fmt.Fprintln(ui.Err)
	if len(slot) > 0 {
		fmt.Fprintln(ui.Err, "    identity_type = \"PIV\"")
		fmt.Fprintf(ui.Err, "    identity_piv_slot = \"%s\"\n", slot)
	} else {
		fmt.Fprintln(ui.Err, "    identity_type = \"\"")
		fmt.Fprintln(ui.Err, "    identity_piv_slot = \"\"")
	}
	fmt.Fprintln(ui.Err)

	fmt.Fprintln(ui.Err, "cleaning is completed ✔️")

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
