package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

var commands = []string{
	"init",
	"key",
	"status",
	"add",
	"delete",
	"list",
	"show",
	"cat",
	"clipboard",
	"decrypt",
	"reencrypt",
	"rotate",
	"bash",
	"help",
}

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
func completeCommand(opts setup.Options, args []string) error {

	listHeaders := func() ([]*header.Header, error) {
		s, err := setupEnv(opts)
		if err != nil {
			return nil, err
		}
		if s.Id.Id == nil {
			return nil, nil
		}

		var headers []*header.Header
		for h := range headerGenerator(s.Repository, s.Id) {
			headers = append(headers, h)
		}
		return headers, nil
	}

	listFiles := func() ([]string, error) {
		return filesForAddCmd("."), nil
	}

	completions, err := getCompletions(args, listHeaders, listFiles)
	if err != nil {
		return err
	}
	for _, c := range completions {
		fmt.Println(c)
	}
	return nil
}

type HeaderListFunc func() ([]*header.Header, error)
type FileListFunc func() ([]string, error)

func getCompletions(args []string, listHeaders HeaderListFunc, listFiles FileListFunc) ([]string, error) {
	if len(args) < 2 {
		return nil, nil
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
		var completions []string
		for _, c := range commands {
			if strings.HasPrefix(c, lastWord) {
				completions = append(completions, c)
			}
		}
		return completions, nil
	}

	if cursorIndex > commandIndex {
		// We have a subcommand, delegate to specific completion logic
		cmd := args[commandIndex]
		lastWord := args[cursorIndex]

		switch cmd {
		case "show", "cat", "delete", "clipboard", "decrypt":
			headers, err := listHeaders()
			if err != nil {
				return nil, nil
			}
			return completeLabels(headers, lastWord), nil
		case "list":
			headers, err := listHeaders()
			if err != nil {
				return nil, nil
			}
			return completeCategoriesAndLabels(headers, lastWord), nil
		case "add":
			// We ignore header errors to allow at least "credential" completion
			headers, _ := listHeaders()
			// We ignore file errors to allow other completions
			files, _ := listFiles()
			return completeAdd(headers, files, args, commandIndex, lastWord), nil
		}
	}

	return nil, nil
}

func completeLabels(headers []*header.Header, prefix string) []string {
	var completions []string
	for _, h := range headers {
		if strings.HasPrefix(h.Label, prefix) {
			completions = append(completions, h.Label)
		}
	}
	return completions
}

func completeCategoriesAndLabels(headers []*header.Header, prefix string) []string {
	categories := map[string]struct{}{}
	var completions []string

	for _, h := range headers {
		if strings.HasPrefix(h.Label, prefix) {
			completions = append(completions, h.Label)
		}
		categories[h.Category] = struct{}{}
	}

	for cat := range categories {
		if strings.HasPrefix(cat, prefix) {
			completions = append(completions, cat)
		}
	}

	return completions
}

func completeAdd(headers []*header.Header, files []string, args []string, commandIndex int, prefix string) []string {
	// args[commandIndex] is "add"
	// args[commandIndex+1] is category
	// args[commandIndex+2] is label

	relativeIndex := len(args) - 1 - commandIndex

	if relativeIndex == 1 {
		var completions []string
		// complete categories
		if headers != nil {
			categories := map[string]struct{}{}
			for _, h := range headers {
				categories[h.Category] = struct{}{}
			}
			for cat := range categories {
				if strings.HasPrefix(cat, prefix) {
					completions = append(completions, cat)
				}
			}
		}
		// Always suggest credential
		if strings.HasPrefix(header.CategoryCredential, prefix) {
			completions = append(completions, header.CategoryCredential)
		}
		return completions
	}

	if relativeIndex == 2 {
		var completions []string
		// suggest files in current directory
		for _, f := range files {
			if strings.HasPrefix(f, prefix) {
				completions = append(completions, f)
			}
		}
		return completions
	}

	return nil
}

// filesForAddCmd returns a slice of files for the add command.
//
// It excludes dot and age files, symlinks and directory paths.
func filesForAddCmd(root string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		// 1. Identify hidden files/dirs by name (works for abs paths and subdirs)
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				// We do not want to walk into hidden directories (like .git)
				// But we must be careful: if root IS a hidden dir (like .), we shouldn't skip it immediately
				// However, standard usage is root="." which has Name=".".
				if s != root {
					return filepath.SkipDir
				}
				// If s == root and it starts with ., we proceed (but we don't append it because it's a dir)
			}
			// If it's a file starting with ., skip it
			if !d.IsDir() {
				return nil
			}
		}

		if filepath.Ext(d.Name()) == AgeExtension {
			return nil
		}

		// no dir, no symlink
		if !d.Type().IsRegular() {
			return nil
		}

		a = append(a, s)

		return nil
	})

	return a
}
