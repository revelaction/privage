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

// TestHelper encapsulates the test environment and helper methods.
type TestHelper struct {
	*setup.Setup
	t    *testing.T
	Root string
}

// NewTestHelper creates a standard environment with a valid age key.
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()
	tmpDir := t.TempDir()

	// Isolation: Ensure home directory is redirected to temp dir
	t.Setenv("HOME", tmpDir)

	idPath := filepath.Join(tmpDir, "privage-key.txt")
	f, err := os.Create(idPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := identity.GenerateAge(f); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	f, err = os.Open(idPath)
	if err != nil {
		t.Fatal(err)
	}
	ident := identity.LoadAge(f, idPath)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	s := &setup.Setup{Id: ident, Repository: tmpDir}
	return &TestHelper{Setup: s, t: t, Root: tmpDir}
}

// AddEncryptedFile creates a valid encrypted file in the setup repository.
func (th *TestHelper) AddEncryptedFile(label, category, content string) {
	th.t.Helper()
	h := &header.Header{Label: label, Category: category}
	if err := encryptSave(h, "", strings.NewReader(content), th.Setup); err != nil {
		th.t.Fatalf("failed to encrypt %s: %v", label, err)
	}
}

// AddFile creates a plain file (for testing non-encrypted file listing).
func (th *TestHelper) AddFile(name string) {
	th.t.Helper()
	path := filepath.Join(th.Root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		th.t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		th.t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		th.t.Fatal(err)
	}
}