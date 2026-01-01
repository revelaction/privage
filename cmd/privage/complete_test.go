package main

import (
	"testing"

	"github.com/revelaction/privage/setup"
)

func TestCompleteAction_Subcommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string // Strings that should be present in output
	}{
		{
			name:     "Empty args",
			args:     []string{"--", "privage"}, // Should ideally be ["--", "privage", ""], but let's see logic
			contains: []string{},                // len < 2 returns nil, but now it's len 2
		},
		{
			name:     "Command completion (empty)",
			args:     []string{"--", "privage", ""},
			contains: []string{"show", "add", "list", "init"},
		},
		{
			name:     "Command completion (partial)",
			args:     []string{"--", "privage", "sh"},
			contains: []string{"show"},
		},
		{
			name:     "Global flag skipping",
			args:     []string{"--", "privage", "-k", "key.txt", ""},
			contains: []string{"show", "add"},
		},
		{
			name:     "Global flag skipping (partial)",
			args:     []string{"--", "privage", "-k", "key.txt", "sh"},
			contains: []string{"show"},
		},
		{
			name:     "Global flag skipping (multiple)",
			args:     []string{"--", "privage", "-c", "conf", "-r", "repo", ""},
			contains: []string{"show", "add"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, err := getCompletions(setup.Options{}, tt.args)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, c := range tt.contains {
				found := false
				for _, s := range suggestions {
					if s == c {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected suggestions to contain '%s', got: %v", c, suggestions)
				}
			}
		})
	}
}