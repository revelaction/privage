package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestList_All(t *testing.T) {
	tests := []struct {
		name      string
		setupData func(th *TestHelper)
		contains  []string
		countMsg  string
	}{
		{
			name:      "Empty Repository",
			setupData: func(th *TestHelper) {},
			contains:  []string{},
			countMsg:  "Found 0 total encrypted tracked files",
		},
		{
			name: "Multiple Files",
			setupData: func(th *TestHelper) {
				th.AddEncryptedFile("report.pdf", "work", "content")
				th.AddEncryptedFile("photo.jpg", "personal", "content")
			},
			contains: []string{
				"report.pdf",
				"photo.jpg",
				"work",
				"personal",
			},
			countMsg: "Found 2 total encrypted tracked files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelper(t)
			tt.setupData(th)

			var outBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &bytes.Buffer{}}

			err := listCommand(th.Setup, "", ui)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := outBuf.String()
			if !strings.Contains(output, tt.countMsg) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.countMsg, output)
			}

			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestList_Filter(t *testing.T) {
	// Standard setup for filter tests
	setupData := func(th *TestHelper) {
		th.AddEncryptedFile("invoice.pdf", "finance", "content")
		th.AddEncryptedFile("salary.xls", "finance", "content")
		th.AddEncryptedFile("todo.txt", "personal", "content")
	}

	tests := []struct {
		name     string
		filter   string
		contains []string
		excludes []string
	}{
		{
			name:     "Match Category",
			filter:   "finance",
			contains: []string{"invoice.pdf", "salary.xls", "Found 2 files with category matching 'finance'"},
			excludes: []string{"todo.txt"},
		},
		{
			name:     "Match Label",
			filter:   "todo",
			contains: []string{"todo.txt", "Found 1 files with name matching 'todo'"},
			excludes: []string{"invoice.pdf", "salary.xls"},
		},
		{
			name:     "Match Both (Partial)",
			filter:   "inv",
			contains: []string{"invoice.pdf"}, // Matches label 'invoice'
			excludes: []string{"salary.xls", "todo.txt"},
		},
		{
			name:     "No Match",
			filter:   "missing",
			contains: []string{"Found no encrypted tracked files matching 'missing'"},
			excludes: []string{"invoice.pdf", "salary.xls", "todo.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelper(t)
			setupData(th)

			var outBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &bytes.Buffer{}}

			err := listCommand(th.Setup, tt.filter, ui)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := outBuf.String()

			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, output)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(output, s) {
					t.Errorf("expected output NOT to contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestList_Error(t *testing.T) {
	th := NewTestHelper(t)
	// Corrupt identity
	th.Id.Id = nil
	th.Id.Err = errors.New("simulated error")

	ui := UI{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

	err := listCommand(th.Setup, "", ui)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNoIdentity) {
		t.Errorf("expected ErrNoIdentity, got %v", err)
	}
}
