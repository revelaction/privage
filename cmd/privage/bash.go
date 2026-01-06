package main

import (
	"flag"
	"fmt"
	"os"
)

const complete = `#! /bin/bash

_privage_autocomplete() {
    local cur prev words cword
    
    # Try to initialize using bash-completion if available
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -n "=:" 2>/dev/null
    fi
    
    # Fallback if cur is not set (e.g. _init_completion failed or missing)
    if [[ -z "$cur" ]]; then
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi

    # call privage complete with all words
    # we use -- to separate flags for "complete" command if we added any, 
    # but more importantly to pass the user's command line safely.
    local suggestions=$(privage complete -- "${COMP_WORDS[@]}")
    
    if [ $? -eq 0 ]; then
        COMPREPLY=( $(compgen -W "$suggestions" -- "$cur") )
    fi
}

complete -F _privage_autocomplete privage
`

func bashCommand(ui UI) error {
	fs := flag.NewFlagSet("bash", flag.ContinueOnError)
	fs.SetOutput(ui.Err)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(ui.Err, "Usage: %s bash\n", os.Args[0])
		_, _ = fmt.Fprintf(ui.Err, "\nDescription:\n")
		_, _ = fmt.Fprintf(ui.Err, "  Dump bash complete script.\n")
	}

	// Note: args not passed to bashCommand anymore, but it had no args anyway
	// If it needs them in the future we can adjust main.go

	_, err := fmt.Fprint(ui.Out, complete)
	return err
}