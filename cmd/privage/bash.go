package main

import (
	"fmt"
)

const complete = `#! /bin/bash

_privage_autocomplete() {
    local cur prev words cword
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -n "=:" || return
    else
        COMPREPLY=()
        _get_comp_words_by_ref -n "=:" cur prev words cword
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

func bashAction(args []string) error {
	fmt.Println(complete)
	return nil
}
