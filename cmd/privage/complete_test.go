package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/revelaction/privage/setup"
)

func TestCompleteCommand(t *testing.T) {
	tests := []struct {
		name      string
		setupData func(th *TestHelper)
		args      []string
		contains  []string
	}{
		{
			name:      "Command completion (empty)",
			setupData: func(th *TestHelper) {},
			args:      []string{"--", "privage", ""},
			contains:  []string{"show", "add", "list", "init", "help"},
		},
		{
			name:      "Command completion (partial)",
			setupData: func(th *TestHelper) {},
			args:      []string{"--", "privage", "ve"},
			contains:  []string{"version"},
		},
		{
			name: "Show Label",
			setupData: func(th *TestHelper) {
				th.AddEncryptedFile("mycred", "credential", "pass")
				th.AddEncryptedFile("work_stuff", "work", "doc")
			},
			args:     []string{"--", "privage", "show", "my"},
			contains: []string{"mycred"},
		},
		{
			name: "Show Label (All)",
			setupData: func(th *TestHelper) {
				th.AddEncryptedFile("mycred", "credential", "pass")
				th.AddEncryptedFile("work_stuff", "work", "doc")
			},
			args:     []string{"--", "privage", "show", ""},
			contains: []string{"mycred", "work_stuff"},
		},
		{
			name: "Add Category (Credential)",
			setupData: func(th *TestHelper) {
				// No existing files needed for this
			},
			args:     []string{"--", "privage", "add", "cred"},
			contains: []string{"credential"},
		},
		{
			name: "Add File (Local)",
			setupData: func(th *TestHelper) {
				// We need to create a plain file in the repo
				th.AddFile("local.txt")
			},
			args:     []string{"--", "privage", "add", "work", "loc"},
			contains: []string{"local.txt"},
		},
		{
			name: "Show Field (Credential)",
			setupData: func(th *TestHelper) {
				th.AddEncryptedFile("mycred", "credential", "pass")
			},
			args:     []string{"--", "privage", "show", "mycred", "pas"},
			contains: []string{"password"},
		},
		{
			name: "Show Field (Non-Credential)",
			setupData: func(th *TestHelper) {
				// "work" category implies non-credential unless specialized
				th.AddEncryptedFile("work_stuff", "work", "doc")
			},
			args:     []string{"--", "privage", "show", "work_stuff", ""},
			contains: []string{}, // Should be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelper(t)
			tt.setupData(th)

			// We need to switch to the tmpDir because filesForAddCmd uses "."
			// This simulates the user being in the repo directory
			oldWd, _ := os.Getwd()
			if err := os.Chdir(th.Root); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(oldWd)
			}()

			// Construct options that point to our test environment
			opts := setup.Options{
				KeyFile:  th.Id.Path,
				RepoPath: th.Root,
			}

			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}

			err := completeCommand(opts, tt.args, ui)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := outBuf.String()
			completions := strings.Split(strings.TrimSpace(output), "\n")
			// Handle empty case where Split returns [""]
			if len(completions) == 1 && completions[0] == "" {
				completions = []string{}
			}

			if tt.name == "Show Field (Non-Credential)" && len(completions) != 0 {
				t.Errorf("expected no completions, got %v", completions)
			}

			for _, c := range tt.contains {
				assertContains(t, completions, c)
			}
		})
	}
}

func assertContains(t *testing.T, list []string, item string) {
	t.Helper()
	found := false
	for _, s := range list {
		if s == item {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected list to contain %q, got: %v", item, list)
	}
}