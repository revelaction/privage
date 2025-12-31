package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/revelaction/privage/setup"
)

// captureStdout captures the output of a function that writes to stdout
func captureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}

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
			output, err := captureStdout(func() error {
				return completeAction(setup.Options{}, tt.args)
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, c := range tt.contains {
				if !strings.Contains(output, c) {
					t.Errorf("expected output to contain '%s', got:\n%s", c, output)
				}
			}
		})
	}
}
