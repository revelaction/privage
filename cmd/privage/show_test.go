package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

func TestShowCommand(t *testing.T) {
	// 1. Setup real environment
	tmpDir := t.TempDir()
	idPath := filepath.Join(tmpDir, "key.age")
	f, _ := os.Create(idPath)
	identity.GenerateAge(f)
	f.Close()

	f, _ = os.Open(idPath)
	ident := identity.LoadAge(f, idPath)
	f.Close()

	s := &setup.Setup{
		Id:         ident,
		Repository: tmpDir,
	}

	// 2. Encrypt a credential
	label := "mycred"
	secretContent := `login = "user123"
password = "supersecret"
`
	h := &header.Header{Label: label, Category: header.CategoryCredential}
	if err := encryptSave(h, "", strings.NewReader(secretContent), s); err != nil {
		t.Fatal(err)
	}

	// 3. Run Command - Full Show
	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}
	
	err := showCommand(s, []string{label}, ui)
	if err != nil {
		t.Fatalf("showCommand failed: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "user123") {
		t.Errorf("expected output to contain login 'user123', got %q", output)
	}

	// 4. Run Command - Specific field
	outBuf.Reset()
	err = showCommand(s, []string{label, "password"}, ui)
	if err != nil {
		t.Fatalf("showCommand failed: %v", err)
	}
	if outBuf.String() != "supersecret" {
		t.Errorf("expected %q, got %q", "supersecret", outBuf.String())
	}
}

func TestShowCommand_WrongCategory(t *testing.T) {
	tmpDir := t.TempDir()
	idPath := filepath.Join(tmpDir, "key.age")
	f, _ := os.Create(idPath)
	identity.GenerateAge(f)
	f.Close()

	f, _ = os.Open(idPath)
	ident := identity.LoadAge(f, idPath)
	f.Close()

	s := &setup.Setup{
		Id:         ident,
		Repository: tmpDir,
	}

	label := "not-a-cred"
	h := &header.Header{Label: label, Category: "other"}
	if err := encryptSave(h, "", strings.NewReader("some content"), s); err != nil {
		t.Fatal(err)
	}

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := showCommand(s, []string{label}, ui)
	if err == nil {
		t.Fatal("expected error for non-credential category, got nil")
	}
	if !strings.Contains(err.Error(), "is not a credential") {
		t.Errorf("expected 'is not a credential' error, got: %v", err)
	}
}