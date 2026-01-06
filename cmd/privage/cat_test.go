package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

func TestCatCommand(t *testing.T) {
	// setupTestEnv creates a standard environment with a valid age key
	setupTestEnv := func(t *testing.T) (*setup.Setup, string) {
		tmpDir := t.TempDir()
		idPath := filepath.Join(tmpDir, "key.age")
		f, err := os.Create(idPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := identity.GenerateAge(f); err != nil {
			t.Fatal(err)
		}
		f.Close()

		f, _ = os.Open(idPath)
		ident := identity.LoadAge(f, idPath)
		f.Close()

		return &setup.Setup{Id: ident, Repository: tmpDir}, tmpDir
	}

	tests := []struct {
		name           string
		setupData      func(t *testing.T, s *setup.Setup)
		label          string
		expectedOutput string
		expectedErr    string
	}{
		{
			name: "Success",
			setupData: func(t *testing.T, s *setup.Setup) {
				label := "secret.txt"
				content := "real secret content"
				h := &header.Header{Label: label}
				if err := encryptSave(h, "", strings.NewReader(content), s); err != nil {
					t.Fatalf("failed to encrypt: %v", err)
				}
			},
			label:          "secret.txt",
			expectedOutput: "real secret content",
		},
		{
			name: "Label Not Found",
			setupData: func(t *testing.T, s *setup.Setup) {
				// No files encrypted
			},
			label:       "missing.txt",
			expectedErr: "file \"missing.txt\" not found in repository",
		},
		{
			name: "Identity Error",
			setupData: func(t *testing.T, s *setup.Setup) {
				s.Id.Id = nil
				s.Id.Err = errors.New("key failed")
			},
			label:       "any.txt",
			expectedErr: "found no privage key file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := setupTestEnv(t)
			tt.setupData(t, s)

			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}

			err := catCommand(s, tt.label, ui)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if outBuf.String() != tt.expectedOutput {
					t.Errorf("expected output %q, got %q", tt.expectedOutput, outBuf.String())
				}
			}
		})
	}
}
