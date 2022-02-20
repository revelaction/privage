package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"io/fs"
	"path/filepath"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

const complete = `#! /bin/bash

: ${PROG:=$(basename ${BASH_SOURCE})}

_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if [[ "$cur" == "-"* ]]; then
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} ${cur} --generate-bash-completion )
    else
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    fi
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
  fi
}

complete -o bashdefault -o default -o nospace -F _cli_bash_autocomplete $PROG
unset PROG`

func bashAction(ctx *cli.Context) error {

	fmt.Println(complete)
	return nil
}

func bashCompleteLabel(ctx *cli.Context) {

	if ctx.NArg() > 0 {
		return
	}

	s, ok := ctx.App.Metadata["setup"].(*setup.Setup)

	if !ok {
		return
	}

	if s.Id.Id == nil {
		return
	}

	for h := range headerGenerator(s.Repository, s.Id) {
		fmt.Println(h.Label)
	}
}

func bashCompleteCategory(ctx *cli.Context) {

	if ctx.NArg() > 0 {
		return
	}

	s, ok := ctx.App.Metadata["setup"].(*setup.Setup)

	if !ok {
		return
	}

	if s.Id.Id == nil {
		return
	}

	categories := map[string]struct{}{}

	for h := range headerGenerator(s.Repository, s.Id) {
		if _, ok := categories[h.Category]; !ok {
			categories[h.Category] = struct{}{}
		}

	}

	for k, _ := range categories {
		fmt.Println(k)
	}
}

func bashCompleteForAdd(ctx *cli.Context) {

	if ctx.NArg() == 0 {
		s, ok := ctx.App.Metadata["setup"].(*setup.Setup)

		if !ok {
			return
		}

		if s.Id.Id == nil {
			return
		}

		for h := range headerGenerator(".", s.Id) {
			fmt.Println(h.Category)
		}

		// always credential
		fmt.Println(header.CategoryCredential)

		return
	}

	// For the second parameter label
	if ctx.NArg() == 1 {
		cat := ctx.Args().First()
		if header.CategoryCredential == cat {
			return
		}

		for _, f := range filesForAddCmd(".") {
			fmt.Println(f)
		}

		return
	}

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

		if filepath.Ext(d.Name()) == AgeExtension {
			return nil
		}

		// no dir, no symlink
		if !d.Type().IsRegular() {
			return nil
		}

		// no dot
		if "." == s[0:1] {
			return nil
		}

		a = append(a, s)

		return nil
	})

	return a
}
