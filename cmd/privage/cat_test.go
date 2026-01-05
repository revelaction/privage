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
	// 1. Setup real environment in TempDir
	tmpDir := t.TempDir()
	
	// Create identity
	idPath := filepath.Join(tmpDir, "key.age")
	f, err := os.Create(idPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := identity.GenerateAge(f); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Load identity wrapper
	f, _ = os.Open(idPath)
	ident := identity.LoadAge(f, idPath)
	f.Close()

	s := &setup.Setup{
		Id:         ident,
		Repository: tmpDir,
	}

	// 2. Encrypt a real file
	label := "secret.txt"
	content := "real secret content"
	h := &header.Header{Label: label}
	if err := encryptSave(h, "", strings.NewReader(content), s); err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	// 3. Run Command
	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}
	
	err = catCommand(s, []string{label}, ui)

	// 4. Assert
	if err != nil {
		t.Fatalf("catCommand failed: %v", err)
	}

	if outBuf.String() != content {
		t.Errorf("expected output %q, got %q", content, outBuf.String())
	}
}

func TestCatCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	s := &setup.Setup{
		Id:         identity.Identity{Path: "key.age"}, // Path is set but Id is nil
		Repository: tmpDir,
	}
	// Note: In production s.Id.Err would be set if Id is nil.
	s.Id.Err = errors.New("key not found")

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := catCommand(s, []string{"missing"}, ui)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "found no privage key file") {
		t.Errorf("expected key error, got %v", err)
	}
}