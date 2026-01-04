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

func keyCommand(opts setup.Options, args []string) (err error) {
	flagSet := flag.NewFlagSet("key", flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s key\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Decrypt the age private key with the PIV key defined in the .privage.conf file.\n")
	}

	if err = flagSet.Parse(args); err != nil {
		return err
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	// piv functionality requires conf piv slot
	if len(s.C.IdentityPivSlot) == 0 {
		return fmt.Errorf("found no piv slot in conf: %w", s.Id.Err)
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

	_, err = fmt.Fprintf(os.Stdout, "%s\n", ageKey)
	if err != nil {
		return fmt.Errorf("could not copy to the console: %w", err)
	}

	return nil
}
