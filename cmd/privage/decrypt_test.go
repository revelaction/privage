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

func TestDecryptCommand(t *testing.T) {
	// 1. Setup real environment
	tmpDir := t.TempDir()
	idPath := filepath.Join(tmpDir, "key.age")
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

	s := &setup.Setup{
		Id:         ident,
		Repository: tmpDir,
	}

	// 2. Encrypt a file
	label := "target.txt"
	content := "decrypted payload"
	h := &header.Header{Label: label}
	if err := encryptSave(h, "", strings.NewReader(content), s); err != nil {
		t.Fatal(err)
	}

	// 3. Run Command
	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}
	
	err = decryptCommand(s, label, ui)

	// 4. Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created on disk
	decryptedPath := filepath.Join(tmpDir, label)
	got, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("expected %q, got %q", content, string(got))
	}
}

func TestDecryptCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	s := &setup.Setup{
		Id:         identity.Identity{Path: "key.age"},
		Repository: tmpDir,
	}
	s.Id.Err = errors.New("key missing")

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := decryptCommand(s, "missing", ui)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "found no privage key file") {
		t.Errorf("expected key error, got %v", err)
	}
}
