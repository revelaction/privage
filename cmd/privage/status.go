package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/config"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// statusAction prints on the terminal a status of the privage command
// configuration
func statusAction(ctx *cli.Context) error {

	s, ok := ctx.App.Metadata["setup"].(*setup.Setup)
	if !ok {
		return errors.New("Can not cast to Type Setup")
	}

	fmt.Println()

	if s.Id.Id != nil {
		fmt.Printf("🔑 Found age key file %s in %s ✔️\n", id.FileName, s.Id.Path)
	} else {
		fmt.Println("🔑 🚫 Could not find an age key\n")
	}

	fmt.Printf("📂 The directory of the encrypted files is %s ✔️\n", s.Repository)

	if len(s.C.Path) > 0 {
		fmt.Printf("📑 Found config file %s in %s ✔️\n", config.FileName, s.C.Path)

		showUpdateMessage := false
		if s.Id.Path != s.C.IdentityPath {

			fmt.Printf("%4s ⚠ The identity path does not match the identity path in the config file: %s.\n", "", s.C.IdentityPath)
			showUpdateMessage = true
		}

		fmt.Println()
		if showUpdateMessage {

			fmt.Printf("%4s You may want to edit the config file %s\n", "", s.C.Path)
		} else {
			fmt.Printf("%4s The configuration file %s is up to date\n", "", s.C.Path)
		}

		fmt.Println()
	} else {
		fmt.Printf("📑 A config file %s does not exists\n", config.FileName)
		fmt.Println()

	}

	cnt := 0
	if s.Id.Id != nil {
		for _ = range headerGenerator(s.Repository, s.Id) {
			cnt++
		}

		fmt.Printf("🔐  Found %d encrypted files for the age key %s\n", cnt, s.Id.Path)
	}

	return nil
}
