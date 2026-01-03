package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/revelaction/privage/header"
)

// noHeaders is a helper list func returning empty headers
func noHeaders() ([]*header.Header, error) {
	return nil, nil
}

// noFiles is a helper list func returning empty files
func noFiles() ([]string, error) {
	return nil, nil
}

// errorHeaders is a helper list func returning an error
func errorHeaders() ([]*header.Header, error) {
	return nil, errors.New("simulated error")
}

func TestCompleteAction_Subcommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string // Strings that should be present in output
	}{
		{
			name:     "Empty args",
			args:     []string{"--", "privage"},
			contains: []string{},
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
			completions, err := getCompletions(tt.args, noHeaders, noFiles)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, c := range tt.contains {
				assertContains(t, completions, c)
			}
		})
	}
}

func TestComplete_Values(t *testing.T) {
	// Setup static data
	headers := []*header.Header{
		{Label: "mycred", Category: "credential"},
		{Label: "work_stuff", Category: "work"},
		{Label: "other_thing", Category: "other"},
	}
	files := []string{"local.txt", "image.png"}

	// ListFuncs
	listHeaders := func() ([]*header.Header, error) { return headers, nil }
	listFiles := func() ([]string, error) { return files, nil }

	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "Show Label",
			args:     []string{"--", "privage", "show", "my"},
			contains: []string{"mycred"},
		},
		{
			name:     "Show Label (All)",
			args:     []string{"--", "privage", "show", ""},
			contains: []string{"mycred", "work_stuff", "other_thing"},
		},
		{
			name:     "List Category/Label",
			args:     []string{"--", "privage", "list", "wo"},
			contains: []string{"work"}, // category matches
		},
		{
			name:     "Add Category",
			args:     []string{"--", "privage", "add", "wo"},
			contains: []string{"work"},
		},
		{
			name:     "Add Category (Credential)",
			args:     []string{"--", "privage", "add", "cred"},
			contains: []string{"credential"}, // from default
		},
		{
			name:     "Add File (Local)",
			args:     []string{"--", "privage", "add", "work", "loc"},
			contains: []string{"local.txt"},
		},
		{
			name:     "Add File (All)",
			args:     []string{"--", "privage", "add", "work", ""},
			contains: []string{"local.txt", "image.png"},
		},
		{
			name:     "Show Field (Credential)",
			args:     []string{"--", "privage", "show", "mycred", "pas"},
			contains: []string{"password"},
		},
		{
			name:     "Show Field (All Credential Fields)",
			args:     []string{"--", "privage", "show", "mycred", ""},
			contains: []string{"login", "password", "email", "url", "api_key", "remarks"},
		},
		{
			name:     "Show Field (Non-Credential)",
			args:     []string{"--", "privage", "show", "work_stuff", ""},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, err := getCompletions(tt.args, listHeaders, listFiles)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.name == "Show Field (Non-Credential)" && len(completions) != 0 {
				t.Errorf("expected no completions for non-credential field, got %v", completions)
			}
			for _, c := range tt.contains {
				assertContains(t, completions, c)
			}
		})
	}
}

func TestComplete_Errors(t *testing.T) {
	// If getting headers fails, we should handle it gracefully (e.g. return nil or partial)

	// Case 1: Show command - expects headers. If error, returns nil/empty.
	args := []string{"--", "privage", "show", ""}
	completions, err := getCompletions(args, errorHeaders, noFiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(completions) != 0 {
		t.Errorf("expected empty completions on error, got %v", completions)
	}

	// Case 2: Add command - expects headers for categories. If error, should still return "credential"
	args = []string{"--", "privage", "add", "cred"}
	completions, err = getCompletions(args, errorHeaders, noFiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, completions, header.CategoryCredential)
}

func TestFilesForAddCmd(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files
	createFile(t, tmpDir, "file1.txt")
	createFile(t, tmpDir, "file2.log")
	createFile(t, tmpDir, ".hidden")
	createFile(t, tmpDir, "secret.age")
	if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	// Switch to temp dir to simulate real usage
	wd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	files, err := filesForAddCmd(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"file1.txt",
		"file2.log",
	}

	if len(files) != len(expected) {
		t.Errorf("expected %d files, got %d: %v", len(expected), len(files), files)
	}

	for _, exp := range expected {
		assertContains(t, files, exp)
	}

	for _, f := range files {
		if filepath.Base(f) == ".hidden" {
			t.Error("should not contain dot files")
		}
		if filepath.Ext(f) == ".age" {
			t.Error("should not contain .age files")
		}
	}
}

// Helpers

func createFile(t *testing.T, dir, name string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
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
		t.Errorf("expected list to contain '%s', got: %v", item, list)
	}
}
