package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDeleteCommand_Success(t *testing.T) {
	th := NewTestHelper(t)
	label := "my-secret"
	th.AddEncryptedFile(label, "credential", "some content")

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := deleteCommand(th.Setup, label, ui)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify success message
	if !strings.Contains(errBuf.String(), "deleted encrypted file for "+label) {
		t.Errorf("expected success message, got: %s", errBuf.String())
	}

	// Verify file is actually gone
	found := false
	for h := range headerGenerator(th.Repository, th.Id) {
		if h.Label == label {
			found = true
			break
		}
	}
	if found {
		t.Error("expected file to be deleted, but it still exists")
	}
}

func TestDeleteCommand_NotFound(t *testing.T) {
	th := NewTestHelper(t)
	label := "non-existent"

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := deleteCommand(th.Setup, label, ui)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify error message in Err (the command doesn't return error when not found, just prints)
	if !strings.Contains(errBuf.String(), "could not find the encrypted file for "+label) {
		t.Errorf("expected not found message, got: %s", errBuf.String())
	}
}

func TestDeleteCommand_NoIdentity(t *testing.T) {
	th := NewTestHelper(t)
	th.Id.Id = nil // Simulate missing identity

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := deleteCommand(th.Setup, "any", ui)
	if err == nil {
		t.Fatal("expected error for missing identity")
	}

	if !strings.Contains(err.Error(), "found no privage key file") {
		t.Errorf("expected key file error, got: %v", err)
	}
}
