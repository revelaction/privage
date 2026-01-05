package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/revelaction/privage/fs"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/identity/piv/yubikey"
	"github.com/revelaction/privage/setup"
)

func keyCommand(s *setup.Setup, args []string, ui UI) (err error) {
	flagSet := flag.NewFlagSet("key", flag.ContinueOnError)
	flagSet.SetOutput(ui.Err)
	flagSet.Usage = func() {
		fmt.Fprintf(ui.Err, "Usage: %s key\n", os.Args[0])
		fmt.Fprintf(ui.Err, "\nDescription:\n")
		fmt.Fprintf(ui.Err, "  Decrypt the age private key with the PIV key defined in the .privage.conf file.\n")
	}

	if err = flagSet.Parse(args); err != nil {
		return err
	}

	// piv functionality requires conf piv slot
	if s.C == nil || len(s.C.IdentityPivSlot) == 0 {
		return fmt.Errorf("found no piv slot in conf")
	}

	ps, err := strconv.ParseUint(s.C.IdentityPivSlot, 16, 32)
	if err != nil {
		return fmt.Errorf("could not convert slot %s to hex: %v", s.C.IdentityPivSlot, err)
	}

	device, err := yubikey.New()
	if err != nil {
		return fmt.Errorf("could not create yubikey device: %w", err)
	}
	defer func() {
		if cerr := device.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	f, err := fs.OpenFile(s.Id.Path)
	if err != nil {
		return fmt.Errorf("could not open key file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	ageKey, err := id.DecryptPiv(f, device, uint32(ps))
	if err != nil {
		return fmt.Errorf("could not decrypt age key: %w", err)
	}

	_, err = fmt.Fprintf(ui.Out, "%s\n", ageKey)
	if err != nil {
		return fmt.Errorf("could not copy to the console: %w", err)
	}

	return nil
}