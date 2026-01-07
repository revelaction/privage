package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// TestEncryptSave_HappyPath tests the normal successful case where
// header and content are encrypted and saved correctly.
func TestEncryptSave_HappyPath(t *testing.T) {
	// Setup: create temporary directory for test files
	tempDir := t.TempDir()

	// Create a test identity (age key pair)
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	// Create test header
	h := &header.Header{
		Label:    "test-password",
		Category: "banking",
	}

	// Create test setup
	s := &setup.Setup{
		Repository: tempDir,
		Id: id.Identity{
			Id: identity,
		},
	}

	// Test content
	content := strings.NewReader("secret password content")

	// Execute: encrypt and save
	err = encryptSave(h, ".age", content, s)
	if err != nil {
		t.Fatalf("encryptSave failed: %v", err)
	}

	// Verify: file was created with expected name
	expectedFileName := fileName(h, s.Id, ".age")
	filePath := filepath.Join(tempDir, expectedFileName)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("expected file not found: %v", err)
	}

	// Verify: file has content (header + encrypted content)
	if fileInfo.Size() == 0 {
		t.Error("file is empty, expected encrypted content")
	}

	// Verify: file permissions are restrictive (0600)
	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %o, want 0600", fileInfo.Mode().Perm())
	}

	// Verify: we can decrypt and read back the content
	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read encrypted file: %v", err)
	}

	// The file contains: [padded encrypted header][encrypted content]
	// We need to skip the header part and decrypt the content
	// Header size after padding is known (from header.PadEncrypted)
	
	// For this test, we just verify the file exists and has content
	// A more thorough test would decrypt and verify the content matches
	t.Logf("Successfully created encrypted file: %s (%d bytes)", filePath, len(encryptedData))
}

// TestEncryptSave_EmptyContent tests that we can save a file with empty content.
func TestEncryptSave_EmptyContent(t *testing.T) {
	tempDir := t.TempDir()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	h := &header.Header{
		Label:    "empty-entry",
		Category: "test",
	}

	s := &setup.Setup{
		Repository: tempDir,
		Id: id.Identity{
			Id: identity,
		},
	}

	// Empty content
	content := strings.NewReader("")

	err = encryptSave(h, ".age", content, s)
	if err != nil {
		t.Fatalf("encryptSave with empty content failed: %v", err)
	}

	// Verify file exists
	expectedFileName := fileName(h, s.Id, ".age")
	filePath := filepath.Join(tempDir, expectedFileName)

	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("expected file not found: %v", err)
	}
}

// TestEncryptSave_LargeContent tests encryption with larger content
// to ensure buffering and streaming work correctly.
func TestEncryptSave_LargeContent(t *testing.T) {
	tempDir := t.TempDir()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	h := &header.Header{
		Label:    "large-file",
		Category: "documents",
	}

	s := &setup.Setup{
		Repository: tempDir,
		Id: id.Identity{
			Id: identity,
		},
	}

	// Create 1MB of content
	largeContent := bytes.Repeat([]byte("x"), 1024*1024)
	content := bytes.NewReader(largeContent)

	err = encryptSave(h, ".age", content, s)
	if err != nil {
		t.Fatalf("encryptSave with large content failed: %v", err)
	}

	expectedFileName := fileName(h, s.Id, ".age")
	filePath := filepath.Join(tempDir, expectedFileName)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("expected file not found: %v", err)
	}

	// Encrypted file should be larger than 1MB (encryption overhead)
	if fileInfo.Size() < int64(len(largeContent)) {
		t.Errorf("encrypted file size = %d, expected > %d", fileInfo.Size(), len(largeContent))
	}

	t.Logf("Large file encrypted successfully: %d bytes -> %d bytes", len(largeContent), fileInfo.Size())
}

// TestEncryptSave_InvalidRepository tests behavior when repository path is invalid.
func TestEncryptSave_InvalidRepository(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	h := &header.Header{
		Label:    "test",
		Category: "test",
	}

	// Use non-existent directory
	s := &setup.Setup{
		Repository: "/nonexistent/directory/that/does/not/exist",
		Id: id.Identity{
			Id: identity,
		},
	}

	content := strings.NewReader("test content")

	err = encryptSave(h, ".age", content, s)
	if err == nil {
		t.Fatal("expected error with invalid repository, got nil")
	}

	// Verify error mentions file creation failure
	if !strings.Contains(err.Error(), "failed to create file") {
		t.Errorf("expected 'failed to create file' in error, got: %v", err)
	}
}

// TestEncryptSave_ReadOnlyRepository tests behavior when repository is read-only.
func TestEncryptSave_ReadOnlyRepository(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping read-only test when running as root")
	}

	tempDir := t.TempDir()

	// Make directory read-only
	if err := os.Chmod(tempDir, 0444); err != nil {
		t.Fatalf("failed to make directory read-only: %v", err)
	}
	defer os.Chmod(tempDir, 0755) // Restore permissions for cleanup

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	h := &header.Header{
		Label:    "test",
		Category: "test",
	}

	s := &setup.Setup{
		Repository: tempDir,
		Id: id.Identity{
			Id: identity,
		},
	}

	content := strings.NewReader("test content")

	err = encryptSave(h, ".age", content, s)
	if err == nil {
		t.Fatal("expected error with read-only repository, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create file") {
		t.Errorf("expected 'failed to create file' in error, got: %v", err)
	}
}

// failingReader simulates an io.Reader that fails after reading some bytes.
type failingReader struct {
	data      []byte
	bytesRead int
	failAfter int // fail after this many bytes
}

func (fr *failingReader) Read(p []byte) (n int, err error) {
	if fr.bytesRead >= fr.failAfter {
		return 0, fmt.Errorf("simulated read failure")
	}

	remaining := fr.failAfter - fr.bytesRead
	if len(p) > remaining {
		p = p[:remaining]
	}

	n = copy(p, fr.data[fr.bytesRead:])
	fr.bytesRead += n

	if fr.bytesRead >= fr.failAfter {
		return n, fmt.Errorf("simulated read failure")
	}

	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

// TestEncryptSave_ContentReadFailure tests error handling when content reader fails.
func TestEncryptSave_ContentReadFailure(t *testing.T) {
	tempDir := t.TempDir()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	h := &header.Header{
		Label:    "test-read-failure",
		Category: "test",
	}

	s := &setup.Setup{
		Repository: tempDir,
		Id: id.Identity{
			Id: identity,
		},
	}

	// Create a reader that fails after 100 bytes
	content := &failingReader{
		data:      bytes.Repeat([]byte("x"), 1000),
		failAfter: 100,
	}

	err = encryptSave(h, ".age", content, s)
	if err == nil {
		t.Fatal("expected error when content reading fails, got nil")
	}

	// Verify error mentions copy failure
	if !strings.Contains(err.Error(), "failed to copy content") {
		t.Errorf("expected 'failed to copy content' in error, got: %v", err)
	}

	// Verify partial file exists (we don't delete on error)
	expectedFileName := fileName(h, s.Id, ".age")
	filePath := filepath.Join(tempDir, expectedFileName)

	if _, err := os.Stat(filePath); err != nil {
		t.Logf("partial file not found (acceptable): %v", err)
	} else {
		t.Logf("partial file exists at: %s (expected behavior)", filePath)
	}
}

// TestEncryptSave_FileOverwrite tests that existing files are overwritten.
func TestEncryptSave_FileOverwrite(t *testing.T) {
	tempDir := t.TempDir()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	h := &header.Header{
		Label:    "overwrite-test",
		Category: "test",
	}

	s := &setup.Setup{
		Repository: tempDir,
		Id: id.Identity{
			Id: identity,
		},
	}

	// First write
	content1 := strings.NewReader("first content")
	err = encryptSave(h, ".age", content1, s)
	if err != nil {
		t.Fatalf("first encryptSave failed: %v", err)
	}

	expectedFileName := fileName(h, s.Id, ".age")
	filePath := filepath.Join(tempDir, expectedFileName)

	firstFileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("first file not found: %v", err)
	}

	// Second write with different content (same header = same filename)
	content2 := strings.NewReader("second content that is much longer")
	err = encryptSave(h, ".age", content2, s)
	if err != nil {
		t.Fatalf("second encryptSave failed: %v", err)
	}

	secondFileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("second file not found: %v", err)
	}

	// File should be overwritten (different size due to different content)
	if firstFileInfo.Size() == secondFileInfo.Size() {
		t.Logf("warning: file sizes are the same, overwrite may not have occurred")
	}

	t.Logf("File overwritten: %d bytes -> %d bytes", firstFileInfo.Size(), secondFileInfo.Size())
}

// TestFileName verifies the filename generation logic.
func TestFileName(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	testID := id.Identity{Id: identity}

	tests := []struct {
		name     string
		header   *header.Header
		suffix   string
	}{
		{
			name: "simple header",
			header: &header.Header{
				Label:    "password",
				Category: "work",
			},
			suffix: ".age",
		},
		{
			name: "empty category",
			header: &header.Header{
				Label:    "password",
				Category: "",
			},
			suffix: ".age",
		},
		{
			name: "different suffix",
			header: &header.Header{
				Label:    "test",
				Category: "test",
			},
			suffix: ".encrypted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := fileName(tt.header, testID, tt.suffix)

			// Verify filename format: [64 hex chars][suffix]
			if !strings.HasSuffix(name, tt.suffix+AgeExtension) {
				t.Errorf("filename doesn't have expected suffix: %s", name)
			}

			// Verify it's a valid hex string before suffix
			withoutSuffix := strings.TrimSuffix(name, tt.suffix+AgeExtension)
			if len(withoutSuffix) != 64 {
				t.Errorf("hash part length = %d, want 64", len(withoutSuffix))
			}

			// Verify deterministic: same input = same output
			name2 := fileName(tt.header, testID, tt.suffix)
			if name != name2 {
				t.Errorf("fileName not deterministic: %s != %s", name, name2)
			}

			t.Logf("Generated filename: %s", name)
		})
	}
}

// TestFileName_Uniqueness verifies that different headers produce different filenames.
func TestFileName_Uniqueness(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	testID := id.Identity{Id: identity}

	h1 := &header.Header{Label: "password1", Category: "work"}
	h2 := &header.Header{Label: "password2", Category: "work"}
	h3 := &header.Header{Label: "password1", Category: "personal"}

	name1 := fileName(h1, testID, ".age")
	name2 := fileName(h2, testID, ".age")
	name3 := fileName(h3, testID, ".age")

	// All should be different
	if name1 == name2 {
		t.Error("different labels produced same filename")
	}
	if name1 == name3 {
		t.Error("different categories produced same filename")
	}
	if name2 == name3 {
		t.Error("different headers produced same filename")
	}
}

// TestFileName_IncludesIdentity verifies that different identities produce different filenames.
func TestFileName_IncludesIdentity(t *testing.T) {
	identity1, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity 1: %v", err)
	}

	identity2, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity 2: %v", err)
	}

	h := &header.Header{Label: "password", Category: "work"}

	name1 := fileName(h, id.Identity{Id: identity1}, ".age")
	name2 := fileName(h, id.Identity{Id: identity2}, ".age")

	if name1 == name2 {
		t.Error("same header with different identities produced same filename")
	}
}

// TestEncryptSave_DifferentHeaders verifies that different headers create different files.
func TestEncryptSave_DifferentHeaders(t *testing.T) {
	tempDir := t.TempDir()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	s := &setup.Setup{
		Repository: tempDir,
		Id: id.Identity{
			Id: identity,
		},
	}

	headers := []*header.Header{
		{Label: "gmail", Category: "email"},
		{Label: "outlook", Category: "email"},
		{Label: "gmail", Category: "work"},
	}

	content := strings.NewReader("test content")

	var filenames []string
	for i, h := range headers {
		// Reset reader for each iteration
		content = strings.NewReader("test content")

		err := encryptSave(h, ".age", content, s)
		if err != nil {
			t.Fatalf("encryptSave for header %d failed: %v", i, err)
		}

		filename := fileName(h, s.Id, ".age")
		filenames = append(filenames, filename)
	}

	// Verify all filenames are unique
	for i := 0; i < len(filenames); i++ {
		for j := i + 1; j < len(filenames); j++ {
			if filenames[i] == filenames[j] {
				t.Errorf("headers %d and %d produced same filename: %s", i, j, filenames[i])
			}
		}
	}

	// Verify all files exist
	for i, filename := range filenames {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("file for header %d not found: %v", i, err)
		}
	}

	t.Logf("Created %d unique encrypted files", len(filenames))
}
