package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// setupTestEnv creates a standard environment with a valid age key
func setupTestEnv(t *testing.T) (*setup.Setup, string) {
	t.Helper()
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

// createEncryptedFile creates a valid encrypted file in the setup repository
func createEncryptedFile(t *testing.T, s *setup.Setup, label, category string, content string) {
	t.Helper()
	h := &header.Header{Label: label, Category: category}
	if err := encryptSave(h, "", strings.NewReader(content), s); err != nil {
		t.Fatalf("failed to encrypt %s: %v", label, err)
	}
}

// createFile creates a plain file (for testing non-encrypted file listing)
func createFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}
