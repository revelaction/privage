package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/revelaction/privage/config"
)

func setupAddTest(t *testing.T) (*TestHelper, func()) {
	t.Helper()
	th := NewTestHelper(t)

	// When key and repository are resolved (as in NewTestHelper), 
	// C can be an empty struct. This prevents nil pointer dereference 
	// in credential.New while avoiding redundant path definitions.
	th.C = &config.Config{}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(th.Repository); err != nil {
		t.Fatal(err)
	}

	return th, func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Logf("failed to restore working directory: %v", err)
		}
	}
}

func TestAddCommand_CredentialSuccess(t *testing.T) {
	th, cleanup := setupAddTest(t)
	defer cleanup()

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := addCommand(th.Setup, "credential", "my-cred", ui)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it was created
	if !labelExists("my-cred", th.Id) {
		t.Error("expected label 'my-cred' to exist")
	}

	// Verify showCommand was called (it should output the new credential)
	if !strings.Contains(outBuf.String(), "Password:🔑") {
		t.Errorf("expected credential output, got: %s", outBuf.String())
	}
}

func TestAddCommand_InstructionOutput(t *testing.T) {
	th, cleanup := setupAddTest(t)
	defer cleanup()

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := addCommand(th.Setup, "credential", "my-cred", ui)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify instructions in Err
	expectedInstructions := []string{
		"You can edit the credentials file",
		"privage decrypt my-cred",
		"vim my-cred",
		"privage reencrypt",
	}

	for _, s := range expectedInstructions {
		if !strings.Contains(errBuf.String(), s) {
			t.Errorf("expected instruction %q in errBuf, got: %s", s, errBuf.String())
		}
	}
}

func TestAddCommand_CustomCategorySuccess(t *testing.T) {
	th, cleanup := setupAddTest(t)
	defer cleanup()

	// Create a file to add
	fileName := "my-file.txt"
	content := "secret content"
	if err := os.WriteFile(fileName, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := addCommand(th.Setup, "my-cat", fileName, ui)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it was created
	if !labelExists(fileName, th.Id) {
		t.Error("expected label to exist")
	}

	if !strings.Contains(errBuf.String(), "Added file 'my-file.txt' to category 'my-cat'") {
		t.Errorf("expected success message, got: %s", errBuf.String())
	}
}

func TestAddCommand_LabelAlreadyExists(t *testing.T) {
	th, cleanup := setupAddTest(t)
	defer cleanup()

	th.AddEncryptedFile("existing", "credential", "content")

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := addCommand(th.Setup, "credential", "existing", ui)
	if err == nil {
		t.Fatal("expected error when adding existing label")
	}

	if !strings.Contains(err.Error(), "already exist") {
		t.Errorf("expected 'already exist' error, got: %v", err)
	}
}

func TestAddCommand_CustomCategoryFileNotFound(t *testing.T) {
	th, cleanup := setupAddTest(t)
	defer cleanup()

	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := addCommand(th.Setup, "my-cat", "non-existent", ui)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if !os.IsNotExist(err) {
		t.Errorf("expected not exist error, got: %v", err)
	}
}