package main

import (
	"errors"
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
func rotateCommand(s *setup.Setup, isClean bool, slot string, ui UI) (err error) {
	return rotate(s, isClean, slot, ui)
}

func rotate(s *setup.Setup, isClean bool, slot string, ui UI) (err error) {

	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	numFiles, err := numFilesForIdentity(s.Repository, s.Id)
	if err != nil {
		return fmt.Errorf("failed to count files: %w", err)
	}
	if numFiles == 0 {
		return fmt.Errorf("found no encrypted files with key %s", s.Id.Path)
	}

	_, _ = fmt.Fprintf(ui.Err, "Found %d files encrypted with key %s\n", numFiles, s.Id.Path)
	_, _ = fmt.Fprintln(ui.Err)

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
		numFilesRotate, err = numFilesForIdentity(s.Repository, idRotate)
		if err != nil {
			return fmt.Errorf("failed to count rotated files: %w", err)
		}
		_, _ = fmt.Fprintf(ui.Err, "Found %d files encrypted with the rotated key %s\n", numFilesRotate, idRotate.Path)

		if numFiles == numFilesRotate {

			// the rotate process is completed. Run clean if flag
			if isClean {
				err = cleanRotate(s, idRotate, slot, ui)
				if err != nil {
					return err
				}

				return nil
			}

			_, _ = fmt.Fprintln(ui.Err)
			_, _ = fmt.Fprintln(ui.Err, "rotate is completed ✔️")
			_, _ = fmt.Fprintln(ui.Err, "(Use \"privage rotate --clean\" to clean up old encrypted files and rename the new ones.)")
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

		_, _ = fmt.Fprintf(ui.Err, "🔑 Created new age key file %s✔️\n", idRotate.Path)
	}

	// iterate all encrypted files with the current key and reencrypt.
	//
	// maybe we are in a rerun of the command rotate, after a failing process.
	// some age files present in the repo will be encrypted with the old key and
	// some with the new one.
	numReencrypted := 0
	ch, err := headerGenerator(s.Repository, s.Id)
	if err != nil {
		return err
	}
	for h := range ch {
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

	_, _ = fmt.Fprintf(ui.Err, "🔐  Reencrypted %d files with new key %s\n", numReencrypted, idRotate.Path)
	_, _ = fmt.Fprintln(ui.Err)
	_, _ = fmt.Fprintln(ui.Err, "rotate is completed ✔️")

	if !isClean {
		_, _ = fmt.Fprintln(ui.Err, "(Use \"privage rotate --clean\" to clean up old encrypted files and rename the new ones.)")
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

	_, _ = fmt.Fprintln(ui.Err)
	_, _ = fmt.Fprintln(ui.Err, "Cleaning files...")
	_, _ = fmt.Fprintln(ui.Err)

	// 1) remove all age encrypted fields of old key
	numDeleted := 0
	ch, err := headerGenerator(s.Repository, s.Id)
	if err != nil {
		return err
	}
	for h := range ch {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if !errors.As(h.Err, &e) {
				return h.Err
			}

			continue
		}

		err := os.Remove(h.Path)
		if err != nil {
			_, _ = fmt.Fprintf(ui.Err, "%8s Error while deleting h.Path %s: %s\n", "", h.Path, err)
			return err
		}

		numDeleted++
	}

	_, _ = fmt.Fprintf(ui.Err, "Deleted %d files encrypted with key %s\n", numDeleted, s.Id.Path)
	_, _ = fmt.Fprintln(ui.Err)

	// 2) rename
	numRenamed := 0
	ch2, err := headerGenerator(s.Repository, idRotate)
	if err != nil {
		return err
	}
	for h := range ch2 {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if errors.As(h.Err, &e) {
				_, _ = fmt.Fprintf(ui.Err, "NoIdentityMatchError: Could not decrupt with curreent key %s\n", h.Err)
				continue
			}

			return h.Err
		}

		// The file currently has a path like:  .../hash.rotate.privage
		// We want to remove the .rotate suffix.
		// Since fileName() now uses PrivageExtension, we construct the standard name with it.
		stardardPath := strings.TrimSuffix(h.Path, suffix+PrivageExtension) + PrivageExtension
		err := os.Rename(h.Path, stardardPath)
		if err != nil {
			return err
		}

		numRenamed++
	}

	_, _ = fmt.Fprintf(ui.Err, "Renamed %d rotated files to %s extension\n", numRenamed, PrivageExtension)
	_, _ = fmt.Fprintln(ui.Err)

	// 3) rename old key to back
	backupKeyPath := id.BackupFilePath(s.Repository)
	err = os.Rename(s.Id.Path, backupKeyPath)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(ui.Err, "Renamed old key %s to %s\n", s.Id.Path, backupKeyPath)
	_, _ = fmt.Fprintln(ui.Err)

	// 4) rename new key
	err = os.Rename(idRotate.Path, s.Id.Path)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(ui.Err, "Renamed new key %s to %s\n", idRotate.Path, s.Id.Path)
	_, _ = fmt.Fprintln(ui.Err)

	_, _ = fmt.Fprintf(ui.Err, "The new key is a %s\n", id.FmtType(slot))
	_, _ = fmt.Fprintf(ui.Err, "⚠ Make sure the config file %s has these lines:\n", s.C.Path)
	_, _ = fmt.Fprintln(ui.Err)
	if len(slot) > 0 {
		_, _ = fmt.Fprintln(ui.Err, "    identity_type = \"PIV\"")
		_, _ = fmt.Fprintf(ui.Err, "    identity_piv_slot = \"%s\"\n", slot)
	} else {
		_, _ = fmt.Fprintln(ui.Err, "    identity_type = \"\"")
		_, _ = fmt.Fprintln(ui.Err, "    identity_piv_slot = \"\"")
	}
	_, _ = fmt.Fprintln(ui.Err)

	_, _ = fmt.Fprintln(ui.Err, "cleaning is completed ✔️")

	return nil
}

func numFilesForIdentity(repoDir string, identity id.Identity) (int, error) {
	num := 0
	ch, err := headerGenerator(repoDir, identity)
	if err != nil {
		return 0, err
	}
	for h := range ch {
		if h.Err != nil {
			var e *age.NoIdentityMatchError
			if errors.As(h.Err, &e) {
				continue
			}
		}

		num++
	}

	return num, nil
}
