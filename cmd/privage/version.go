package main

import (
	"fmt"
)

func versionCommand(ui UI) error {
	fmt.Fprintf(ui.Out, "privage version %s (commit: %s, yubikey: %s)\n", BuildTag, BuildCommit, YubikeySupport)
	return nil
}