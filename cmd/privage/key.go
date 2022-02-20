package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"

	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

func keyAction(ctx *cli.Context) error {

	s, ok := ctx.App.Metadata["setup"].(*setup.Setup)
	if !ok {
		return errors.New("Can not cast to Type Setup")
	}

	// piv functionality requires conf piv slot
	if len(s.C.IdentityPivSlot) == 0 {
		return fmt.Errorf("Found no piv slot in conf: %w", s.Id.Err)
	}

	ps, err := strconv.ParseUint(s.C.IdentityPivSlot, 16, 32)
	if err != nil {
		return fmt.Errorf("Could not convert slot %s to hex: %v", s.C.IdentityPivSlot, err)
	}

	ageKey, err := id.LoadRaw(s.Id.Path, uint32(ps), "")
	if err != nil {
		return fmt.Errorf("Could not decrypt age key: %w", err)
	}

	_, err = fmt.Fprintf(os.Stdout, "%s\n", ageKey)
	if err != nil {
		return fmt.Errorf("Could not copy to the console: %w", err)
	}

	return nil
}
