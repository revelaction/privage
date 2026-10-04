package main

import (
	"fmt"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/setup"
)

// statusCommand prints on the terminal a status of the privage command
// configuration.
//
// When repository is true it prints only the repository directory, one line
// and nothing else.
func statusCommand(s *setup.Setup, repository bool, ui UI) error {
	if repository {
		_, _ = fmt.Fprintln(ui.Out, s.Repository)
		return nil
	}

	_, _ = fmt.Fprintln(ui.Out)

	if s.Id.Id != nil {
		_, _ = fmt.Fprintf(ui.Out, "🔑 Found age key file in %s ✔️\n", s.Id.Path)
	} else {
		_, _ = fmt.Fprintln(ui.Out, "🔑 🚫 Could not find an age key")
	}

	_, _ = fmt.Fprintf(ui.Out, "📂 The directory of the encrypted files is %s ✔️\n", s.Repository)

	if s.C != nil && len(s.C.Path) > 0 {
		_, _ = fmt.Fprintf(ui.Out, "📑 Found config file in %s ✔️\n", s.C.Path)

		showUpdateMessage := false
		if s.Id.Path != s.C.IdentityPath {

			_, _ = fmt.Fprintf(ui.Out, "%4s ⚠ The identity path does not match the identity path in the config file: %s.\n", "", s.C.IdentityPath)
			showUpdateMessage = true
		}

		_, _ = fmt.Fprintln(ui.Out)
		if showUpdateMessage {

			_, _ = fmt.Fprintf(ui.Out, "%4s You may want to edit the config file %s\n", "", s.C.Path)
		} else {
			_, _ = fmt.Fprintf(ui.Out, "%4s The configuration file %s is up to date\n", "", s.C.Path)
		}

		_, _ = fmt.Fprintln(ui.Out)
	} else {
		_, _ = fmt.Fprintf(ui.Out, "📑 A config file %s does not exists\n", config.DefaultFileName)
		_, _ = fmt.Fprintln(ui.Out)

	}

	cnt := 0
	if s.Id.Id != nil {
		ch, err := headerGenerator(s.Repository, s.Id)
		if err != nil {
			_, _ = fmt.Fprintf(ui.Out, "🔐  Could not count files: %v\n", err)
		} else {
			for range ch {
				cnt++
			}
			_, _ = fmt.Fprintf(ui.Out, "🔐  Found %d encrypted files for the age key %s\n", cnt, s.Id.Path)
		}
	}

	return nil
}
