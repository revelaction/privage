package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/revelaction/privage/header"
)

// completeAction handles the autocompletion requests triggered by the bash completion script.
//
// Understanding the Argument Flow:
// 1. Bash triggers the completion function and executes:
//    privage complete -- "${COMP_WORDS[@]}"
//
// 2. In main.go, flag.Parse() is called. Since "complete" is the first positional 
//    argument, flag.Parse() stops there. It does NOT consume the "--" separator 
//    because the separator appears after the "complete" command token.
//
// 3. main.go passes flag.Args()[1:] to this function.
//    - args[0]: "--" (The separator inserted by bash.go)
//    - args[1]: "privage" (The first element of COMP_WORDS, the binary name)
//    - args[2...]: The actual command line arguments typed by the user.
//
// Example: User types 'privage -k key.txt show [TAB]'
// - COMP_WORDS: ["privage", "-k", "key.txt", "show", ""]
// - args received here: ["--", "privage", "-k", "key.txt", "show", ""]
// - commandIndex starts at 2, skips "-k" and "key.txt", and identifies "show" at index 4.
func completeAction(args []string) error {

	if len(args) < 2 {
		return nil
	}

	// 1. Find the index where the subcommand should be
	// args[0] is "--", args[1] is "privage"
	commandIndex := 2
	for commandIndex < len(args) {
		arg := args[commandIndex]
		if strings.HasPrefix(arg, "-") {
			commandIndex++ // Skip the flag
			// Check if this global flag takes an argument
			trimmed := strings.TrimLeft(arg, "-")
			switch trimmed {
			case "k", "key", "c", "conf", "p", "piv-slot", "r", "repository":
				commandIndex++ // Skip the flag value
			}
			continue
		}
		break
	}

	cursorIndex := len(args) - 1

	// 2. Decide what to complete based on cursor position relative to command position
	if cursorIndex == commandIndex {
		// User is typing the command itself
		lastWord := args[cursorIndex]
		cmds := []string{"init", "key", "status", "add", "delete", "list", "show", "cat", "clipboard", "decrypt", "reencrypt", "rotate", "bash", "help"}
		for _, c := range cmds {
			if strings.HasPrefix(c, lastWord) {
				fmt.Println(c)
			}
		}
		return nil
	}

	if cursorIndex > commandIndex {
		// We have a subcommand, delegate to specific completion logic
		cmd := args[commandIndex]
		lastWord := args[cursorIndex]

		switch cmd {
		case "show", "cat", "delete", "clipboard", "decrypt":
			return completeLabels(lastWord)
		case "list":
			return completeCategoriesAndLabels(lastWord)
		case "add":
			return completeAdd(args, commandIndex, lastWord)
		}
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

func completeAdd(args []string, commandIndex int, prefix string) error {
	// args[commandIndex] is "add"
	// args[commandIndex+1] is category
	// args[commandIndex+2] is label
	
	relativeIndex := len(args) - 1 - commandIndex

	if relativeIndex == 1 {
		// complete categories
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
		// Always suggest credential
		if strings.HasPrefix(header.CategoryCredential, prefix) {
			fmt.Println(header.CategoryCredential)
		}
		return nil
	}

	if relativeIndex == 2 {
		// suggest files in current directory
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
