package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
)

// validHexName returns a 64-character hex string based on the input seed.
func validHexName(seed string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(seed)))
}

// createTestAgeFile creates an encrypted age file for testing.
func createTestAgeFile(t *testing.T, dir, filename string, h *header.Header, identity *age.X25519Identity) string {
	t.Helper()

	// 1. Serialize and Encrypt Header
	buf := new(bytes.Buffer)
	ageWr, err := age.Encrypt(buf, identity.Recipient())
	if err != nil {
		t.Fatalf("failed to create age encryptor: %v", err)
	}
	headerBytes, err := h.Pad()
	if err != nil {
		t.Fatalf("failed to pad header: %v", err)
	}
	if _, err := ageWr.Write(headerBytes); err != nil {
		t.Fatalf("failed to write padded header to age: %v", err)
	}
	if err := ageWr.Close(); err != nil {
		t.Fatalf("failed to close age encryptor: %v", err)
	}

	// 2. Pad Encrypted Header to BlockSize
	padded, err := header.PadEncrypted(buf.Bytes())
	if err != nil {
		t.Fatalf("failed to pad encrypted header: %v", err)
	}

	// 3. Write to file
	path := filepath.Join(dir, filename+PrivageExtension)
	if err := os.WriteFile(path, padded, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	return path
}

func TestIsPrivageFile(t *testing.T) {
	validHash := validHexName("test")

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Valid cases
		{"Valid hash", validHash + PrivageExtension, true},
		{"Valid hash with suffix", validHash + ".rotate" + PrivageExtension, true},
		{"Valid hash with alphanumeric suffix", validHash + ".v1-backup" + PrivageExtension, true},

		// Invalid lengths / extensions
		{"Too short", "abc.privage", false},
		{"Wrong extension", validHash + ".age", false},
		{"No extension", validHash, false},
		{"Empty string", "", false},

		// Invalid hex prefix
		{"Invalid hex (non-hex char)", "g" + validHash[1:] + PrivageExtension, false},
		{"Invalid hex (uppercase)", "A" + validHash[1:] + PrivageExtension, false},
		{"Invalid hex (too short prefix)", validHash[:63] + ".privage", false},

		// Path traversal and security
		{"Path separator /", validHash + "/suffix" + PrivageExtension, false},
		{"Path separator \\", validHash + "\\suffix" + PrivageExtension, false},
		{"Path traversal ..", validHash + "..suffix" + PrivageExtension, true}, // Dots are OK
		{"Root path", "/" + validHash + PrivageExtension, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPrivageFile(tt.filename); got != tt.want {
				t.Errorf("isPrivageFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestHeaderGenerator(t *testing.T) {
	// Create test identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}
	privageId := id.Identity{Id: identity, Path: "test-key"}

	t.Run("Success_MultipleFiles", func(t *testing.T) {
		tmpDir := t.TempDir()

		headers := []*header.Header{
			{Category: "cat1", Label: "label1"},
			{Category: "cat2", Label: "label2"},
		}

		for i, h := range headers {
			filename := validHexName(fmt.Sprintf("file%d", i))
			createTestAgeFile(t, tmpDir, filename, h, identity)
		}

		gen, err := headerGenerator(tmpDir, privageId)
		if err != nil {
			t.Fatalf("headerGenerator failed: %v", err)
		}
		count := 0
		for h := range gen {
			if h.Err != nil {
				t.Errorf("unexpected error for %s: %v", h.Path, h.Err)
			}
			found := false
			for _, expected := range headers {
				if h.Category == expected.Category && h.Label == expected.Label {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("found unexpected header: category=%s, label=%s", h.Category, h.Label)
			}
			count++
		}

		if count != len(headers) {
			t.Errorf("expected %d headers, got %d", len(headers), count)
		}
	})

	t.Run("FlatRepository_IgnoreSubdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 1. Valid file in root
		rootName := validHexName("root")
		createTestAgeFile(t, tmpDir, rootName, &header.Header{Label: "root"}, identity)

		// 2. File in subdirectory (should be ignored)
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.Mkdir(subDir, 0700); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}
		subName := validHexName("sub")
		createTestAgeFile(t, subDir, subName, &header.Header{Label: "sub"}, identity)

		gen, err := headerGenerator(tmpDir, privageId)
		if err != nil {
			t.Fatalf("headerGenerator failed: %v", err)
		}
		count := 0
		for h := range gen {
			if h.Label != "root" {
				t.Errorf("expected only root file, got: %s", h.Label)
			}
			count++
		}

		if count != 1 {
			t.Errorf("expected 1 file, got %d", count)
		}
	})

	t.Run("PartialFailure_InvalidFiles", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 1. Valid file
		validName := validHexName("valid")
		createTestAgeFile(t, tmpDir, validName, &header.Header{Label: "valid"}, identity)

		// 2. Malformed file (too short)
		shortName := validHexName("short")
		shortPath := filepath.Join(tmpDir, shortName+PrivageExtension)
		if err := os.WriteFile(shortPath, []byte("too short"), 0600); err != nil {
			t.Fatalf("failed to write short test file: %v", err)
		}

		// 3. Wrong identity file
		wrongKeyName := validHexName("wrong_key")
		otherIdentity, _ := age.GenerateX25519Identity()
		createTestAgeFile(t, tmpDir, wrongKeyName, &header.Header{Label: "wrong"}, otherIdentity)

		gen, err := headerGenerator(tmpDir, privageId)
		if err != nil {
			t.Fatalf("headerGenerator failed: %v", err)
		}

		results := make(map[string]*header.Header)
		for h := range gen {
			results[filepath.Base(h.Path)] = h
		}

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		// Verify valid
		if results[validName+PrivageExtension].Err != nil {
			t.Errorf("valid file should not have error: %v", results[validName+PrivageExtension].Err)
		}

		// Verify short
		if results[shortName+PrivageExtension].Err == nil {
			t.Error("short file should have error")
		}

		// Verify wrong key
		if results[wrongKeyName+PrivageExtension].Err == nil {
			t.Error("wrong_key file should have error")
		}
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		tmpDir := t.TempDir()
		name := validHexName("unreadable")
		path := filepath.Join(tmpDir, name+PrivageExtension)
		if err := os.WriteFile(path, []byte("data"), 0000); err != nil { // No permissions
			t.Fatalf("failed to write unreadable test file: %v", err)
		}

		gen, err := headerGenerator(tmpDir, privageId)
		if err != nil {
			t.Fatalf("headerGenerator failed: %v", err)
		}
		h := <-gen
		if h == nil {
			t.Fatal("expected at least one result")
		}
		if h.Err == nil {
			t.Error("expected error for unreadable file")
		}
	})

	t.Run("StandardAgeFile_Collision", func(t *testing.T) {
		tmpDir := t.TempDir()
		name := validHexName("standard")
		path := filepath.Join(tmpDir, name+PrivageExtension)
		
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("failed to create standard age file: %v", err)
		}
		
		aw, err := age.Encrypt(f, identity.Recipient())
		if err != nil {
			_ = f.Close()
			t.Fatalf("failed to create age writer: %v", err)
		}
		if _, err := aw.Write([]byte("some content")); err != nil {
			t.Errorf("failed to write content: %v", err)
		}
		if err := aw.Close(); err != nil {
			t.Errorf("failed to close age writer: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Errorf("failed to close file: %v", err)
		}

		gen, err := headerGenerator(tmpDir, privageId)
		if err != nil {
			t.Fatalf("headerGenerator failed: %v", err)
		}
		h := <-gen
		
		if h == nil {
			t.Fatal("expected result for standard age file")
		}
		
		// We expect an error because it's not a valid privage file
		if h.Err == nil {
			t.Errorf("expected error for standard age file, got success. Parsed header: %+v", h)
		} else {
			t.Logf("Got expected error for standard age file: %v", h.Err)
		}
	})

	t.Run("StandardAgeFile_Large_Collision", func(t *testing.T) {
		tmpDir := t.TempDir()
		name := validHexName("large")
		path := filepath.Join(tmpDir, name+PrivageExtension)
		
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("failed to create standard age file: %v", err)
		}
		
		aw, err := age.Encrypt(f, identity.Recipient())
		if err != nil {
			_ = f.Close()
			t.Fatalf("failed to create age writer: %v", err)
		}
		// Write > 512 bytes
		largeData := bytes.Repeat([]byte("A"), 1024)
		if _, err := aw.Write(largeData); err != nil {
			t.Errorf("failed to write content: %v", err)
		}
		if err := aw.Close(); err != nil {
			t.Errorf("failed to close age writer: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Errorf("failed to close file: %v", err)
		}

		gen, err := headerGenerator(tmpDir, privageId)
		if err != nil {
			t.Fatalf("headerGenerator failed: %v", err)
		}
		h := <-gen
		
		if h == nil {
			t.Fatal("expected result for large standard age file")
		}
		
		if h.Err == nil {
			t.Errorf("expected error for large standard age file, got success")
		} else {
			t.Logf("Got expected error for large standard age file: %v", h.Err)
		}
	})
}
