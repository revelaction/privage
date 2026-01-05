package main

import (
	"fmt"

	"github.com/revelaction/privage/setup"
)

func versionCommand(_ setup.Options, _ []string) error {
	fmt.Printf("privage version %s (commit: %s, yubikey: %s)\n", BuildTag, BuildCommit, YubikeySupport)
	return nil
}

