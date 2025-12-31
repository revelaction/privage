#!/bin/bash
set -e
set -x

# Build
go build -o privage_bin ./cmd/privage

# Mock privage command
privage() {
    ./privage_bin "$@"
}
export -f privage

# Capture completion script
SCRIPT=$(./privage_bin bash)

# Function to run completion test
run_completion() {
    local cmdline="$1"
    local expected="$2"
    
    # Reset completion variables
    COMP_WORDS=($cmdline)
    # Calculate COMP_CWORD (index of last word)
    COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
    
    # If the last character of cmdline is a space, we are completing a new empty word
    if [[ "$cmdline" == *" " ]]; then
        COMP_WORDS+=("")
        COMP_CWORD=$(($COMP_CWORD + 1))
    fi
    
    COMPREPLY=()
    
    # Source script and run function
    eval "$SCRIPT"
    _privage_autocomplete
    
    # Check result
    local result="${COMPREPLY[*]}"
    
    for word in $expected; do
        if [[ "$result" != *"$word"* ]]; then
            echo "FAIL: '$cmdline'"
            echo "  Expected to find: $word"
            echo "  Got: $result"
            exit 1
        fi
    done
    
    echo "PASS: '$cmdline' -> contains all expected words"
}

echo "Running Integration Tests for Autocompletion..."

# 1. Command completion
run_completion "privage " "show init add"

# 2. Command completion partial
run_completion "privage sh" "show"

# 3. Flag skipping
run_completion "privage -k key.txt sh" "show"

# 4. Empty (should be empty if we are just "privage")
# Note: if we type "privage" without space, bash usually handles command completion.
# But if we force it, our script returns nothing if len args < 2
# But run_completion helper mimics "privage" as valid COMP_WORDS=("privage")
# If we want to simulate "privage [TAB]", we pass "privage".
# If we want "privage [SPACE] [TAB]", we pass "privage "

echo "Tests Passed!"
rm privage_bin
