package main

import (
	"fmt"
	"os"
	"strconv"

	id "github.com/revelaction/privage/identity"
)

func keyAction(args []string) error {

	s, err := setupEnv(global)
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

	ageKey, err := id.LoadRaw(s.Id.Path, uint32(ps), "")
	if err != nil {
		return fmt.Errorf("could not decrypt age key: %w", err)
	}

	_, err = fmt.Fprintf(os.Stdout, "%s\n", ageKey)
	if err != nil {
		return fmt.Errorf("could not copy to the console: %w", err)
	}

	return nil
}
