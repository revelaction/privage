package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// statusCommand prints on the terminal a status of the privage command
// configuration
func statusCommand(opts setup.Options, args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s status\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDescription:\n")
		fmt.Fprintf(os.Stderr, "  Provide information about the current configuration.\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := setupEnv(opts)
	if err != nil {
		return fmt.Errorf("unable to setup environment configuration: %s", err)
	}

	fmt.Println()

	if s.Id.Id != nil {
		fmt.Printf("🔑 Found age key file in %s ✔️\n", s.Id.Path)
	} else {
		fmt.Println("🔑 🚫 Could not find an age key")
	}

	fmt.Printf("📂 The directory of the encrypted files is %s ✔️\n", s.Repository)

	if s.C != nil && len(s.C.Path) > 0 {
		fmt.Printf("📑 Found config file in %s ✔️\n", s.C.Path)

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
		fmt.Printf("📑 A config file %s does not exists\n", config.DefaultFileName)
		fmt.Println()

	}

	cnt := 0
	if s.Id.Id != nil {
		for range headerGenerator(s.Repository, s.Id) {
			cnt++
		}

		fmt.Printf("🔐  Found %d encrypted files for the age key %s\n", cnt, s.Id.Path)
	}

	return nil
}
