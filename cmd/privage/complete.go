package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/revelaction/privage/header"
)

// completeAction handles the autocompletion requests.
// args[0] is usually the command being completed (e.g., "privage")
// args[1] is the sub-command (e.g., "show")
// args[2...] are the arguments typed so far
func completeAction(args []string) error {

	if len(args) < 2 {
		return nil
	}

	var cmdToComplete string
	var lastWord string

	knownCmds := map[string]bool{
		"init": true, "key": true, "status": true, "add": true,
		"delete": true, "list": true, "show": true, "cat": true,
		"clipboard": true, "decrypt": true, "reencrypt": true,
		"rotate": true, "bash": true, "help": true,
	}

	// Advanced parsing to skip flags
	for i := 1; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			// Check for global flags that take arguments
			// -k, -key, -c, -conf, -p, -piv-slot, -r, -repository
			// Note: flag package allows -flag or --flag
			trimmed := strings.TrimLeft(arg, "-")
			switch trimmed {
			case "k", "key", "c", "conf", "p", "piv-slot", "r", "repository":
				// These take an argument, so skip the next one if available
				i++
			}
			continue
		}

		if knownCmds[arg] {
			cmdToComplete = arg
			break
		}
	}
	
	lastWord = args[len(args)-1]

	// If we haven't typed a subcommand yet, or we are currently typing it
	// (cmdToComplete == lastWord means we just found it at the end)
	// We check if we are truly at the end (len check might need adjustment if we had flags)
	// Actually, if cmdToComplete == lastWord, we are completing the command name itself.
	if cmdToComplete == "" || cmdToComplete == lastWord {
		// double check: if we typed "privage show <TAB>", args is ["privage", "show", ""]
		// cmdToComplete is "show". lastWord is "". They are different.
		// if "privage show", args is ["privage", "show"]. lastWord is "show". Same.
		
		cmds := []string{"init", "key", "status", "add", "delete", "list", "show", "cat", "clipboard", "decrypt", "reencrypt", "rotate", "bash", "help"}
		for _, c := range cmds {
			if strings.HasPrefix(c, lastWord) {
				fmt.Println(c)
			}
		}
		return nil
	}

	switch cmdToComplete {
	case "show", "cat", "delete", "clipboard", "decrypt":
		return completeLabels(lastWord)
	case "list":
		return completeCategoriesAndLabels(lastWord)
	case "add":
		return completeAdd(args, lastWord)
	}

	return nil
}

func completeLabels(prefix string) error {
	s, err := setupEnv(global.KeyFile, global.ConfigFile, global.RepoPath, global.PivSlot)
	if err != nil {
		return nil
	}
	if s.Id.Id == nil {
		return nil
	}

	for h := range headerGenerator(s.Repository, s.Id) {
		if strings.HasPrefix(h.Label, prefix) {
			fmt.Println(h.Label)
		}
	}
	return nil
}

func completeCategoriesAndLabels(prefix string) error {
	s, err := setupEnv(global.KeyFile, global.ConfigFile, global.RepoPath, global.PivSlot)
	if err != nil {
		return nil
	}
	if s.Id.Id == nil {
		return nil
	}

	categories := map[string]struct{}{}
	
	for h := range headerGenerator(s.Repository, s.Id) {
		if strings.HasPrefix(h.Label, prefix) {
			fmt.Println(h.Label)
		}
		categories[h.Category] = struct{}{}
	}
	
	for cat := range categories {
		if strings.HasPrefix(cat, prefix) {
			fmt.Println(cat)
		}
	}
	
	return nil
}

func completeAdd(args []string, prefix string) error {
	if len(args) <= 3 {
		s, err := setupEnv(global.KeyFile, global.ConfigFile, global.RepoPath, global.PivSlot)
		if err == nil && s.Id.Id != nil {
			categories := map[string]struct{}{}
			for h := range headerGenerator(s.Repository, s.Id) {
				categories[h.Category] = struct{}{}
			}
			for cat := range categories {
				if strings.HasPrefix(cat, prefix) {
					fmt.Println(cat)
				}
			}
		}
		if strings.HasPrefix(header.CategoryCredential, prefix) {
			fmt.Println(header.CategoryCredential)
		}
		return nil
	}
	
	if len(args) == 4 {
		for _, f := range filesForAddCmd(".") {
			if strings.HasPrefix(f, prefix) {
				fmt.Println(f)
			}
		}
	}
	
	return nil
}


// filesForAddCmd returns a slice of files for the add command.
// Copied/Adapted from bash.go (which will be deleted/refactored)
func filesForAddCmd(root string) []string {
	var a []string
	
	// This is a simplified version for the example. 
	// In reality we should use filepath.WalkDir like before.
	entries, err := filepath.Glob(root + "/*")
	if err != nil {
		return nil
	}
	
	for _, e := range entries {
		// primitive filtering
		if strings.HasSuffix(e, ".age") {
			continue
		}
		a = append(a, e)
	}

	return a
}
