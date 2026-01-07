package main

import (
	"fmt"
	"strconv"

	"github.com/revelaction/privage/fs"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/identity/piv/yubikey"
	"github.com/revelaction/privage/setup"
)

func keyCommand(s *setup.Setup, ui UI) (err error) {
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
