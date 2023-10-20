package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"io/fs"
	"path/filepath"

	"github.com/revelaction/privage/header"
)

const complete = `#! /bin/bash

: ${PROG:=$(basename ${BASH_SOURCE})}

# Macs have bash3 for which the bash-completion package doesn't include
# _init_completion. This is a minimal version of that function.
_cli_init_completion() {
  COMPREPLY=()
  _get_comp_words_by_ref "$@" cur prev words cword
}

_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base words
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if declare -F _init_completion >/dev/null 2>&1; then
      _init_completion -n "=:" || return
    else
      _cli_init_completion -n "=:" || return
    fi
    words=("${words[@]:0:$cword}")
    if [[ "$cur" == "-"* ]]; then
      requestComp="${words[*]} ${cur} --generate-shell-completion"
    else
      requestComp="${words[*]} --generate-shell-completion"
    fi
    opts=$(eval "${requestComp}" 2>/dev/null)
    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
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

	s, err := setupEnv(ctx)
	if err != nil {
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

	s, err := setupEnv(ctx)
	if err != nil {
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

	for k := range categories {
		fmt.Println(k)
	}
}

func bashCompleteForAdd(ctx *cli.Context) {

	if ctx.NArg() == 0 {
		s, err := setupEnv(ctx)
		if err != nil {
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
		if s[0:1] == "." {
			return nil
		}

		a = append(a, s)

		return nil
	})

	return a
}
