package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"filippo.io/age"
	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
)

// createTestAgeFile creates an encrypted age file for testing.
func createTestAgeFile(t *testing.T, dir, filename string, h *header.Header, identity *age.X25519Identity) string {
	t.Helper()

	// 1. Serialize and Encrypt Header
	buf := new(bytes.Buffer)
	ageWr, err := age.Encrypt(buf, identity.Recipient())
	if err != nil {
		t.Fatalf("failed to create age encryptor: %v", err)
	}
	if _, err := ageWr.Write(h.Pad()); err != nil {
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
	path := filepath.Join(dir, filename+AgeExtension)
	if err := os.WriteFile(path, padded, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	return path
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
			createTestAgeFile(t, tmpDir, fmt.Sprintf("file%d", i), h, identity)
		}

		gen := headerGenerator(tmpDir, privageId)
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

	t.Run("ResourceLeak_ManyFiles", func(t *testing.T) {
		// Create many files to test file descriptor management.
		// On many systems, default ulimit is 1024. 200 is enough to verify we aren't leaking.
		tmpDir := t.TempDir()
		numFiles := 200
		h := &header.Header{Category: "test", Label: "test"}
		
		for i := 0; i < numFiles; i++ {
			createTestAgeFile(t, tmpDir, fmt.Sprintf("file%d", i), h, identity)
		}

		gen := headerGenerator(tmpDir, privageId)
		count := 0
		for h := range gen {
			if h.Err != nil {
				t.Fatalf("unexpected error at file %d: %v", count, h.Err)
			}
			count++
		}

		if count != numFiles {
			t.Errorf("expected %d files, got %d", numFiles, count)
		}

		// Check open files if on linux (best effort)
		if runtime.GOOS == "linux" {
			fds, err := os.ReadDir("/proc/self/fd")
			if err == nil {
				// We don't know the exact base number, but it should definitely be less than numFiles + base.
				// If we leaked, it would be > 200.
				if len(fds) > 100 { // Normal apps have ~10-20 FDs open
					t.Logf("Warning: suspiciously high number of FDs open: %d", len(fds))
				}
			}
		}
	})

	t.Run("PartialFailure_InvalidFiles", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 1. Valid file
		createTestAgeFile(t, tmpDir, "valid", &header.Header{Label: "valid"}, identity)

		// 2. Malformed file (too short)
		shortPath := filepath.Join(tmpDir, "short"+AgeExtension)
		if err := os.WriteFile(shortPath, []byte("too short"), 0600); err != nil {
			t.Fatalf("failed to write short test file: %v", err)
		}

		// 3. Wrong identity file
		otherIdentity, _ := age.GenerateX25519Identity()
		createTestAgeFile(t, tmpDir, "wrong_key", &header.Header{Label: "wrong"}, otherIdentity)

		gen := headerGenerator(tmpDir, privageId)

		results := make(map[string]*header.Header)
		for h := range gen {
			results[filepath.Base(h.Path)] = h
		}

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		// Verify valid
		if results["valid.age"].Err != nil {
			t.Errorf("valid file should not have error: %v", results["valid.age"].Err)
		}

		// Verify short
		if results["short.age"].Err == nil {
			t.Error("short file should have error")
		} else if !bytes.Contains([]byte(results["short.age"].Err.Error()), []byte("could not read header")) {
			t.Errorf("unexpected error message for short file: %v", results["short.age"].Err)
		}

		// Verify wrong key
		if results["wrong_key.age"].Err == nil {
			t.Error("wrong_key file should have error")
		} else if !bytes.Contains([]byte(results["wrong_key.age"].Err.Error()), []byte("could not Decrypt header")) {
			t.Errorf("unexpected error message for wrong_key file: %v", results["wrong_key.age"].Err)
		}
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "unreadable"+AgeExtension)
		if err := os.WriteFile(path, []byte("data"), 0000); err != nil { // No permissions
			t.Fatalf("failed to write unreadable test file: %v", err)
		}

		gen := headerGenerator(tmpDir, privageId)
		h := <-gen
		if h == nil {
			t.Fatal("expected at least one result")
		}
		if h.Err == nil {
			t.Error("expected error for unreadable file")
		} else if !bytes.Contains([]byte(h.Err.Error()), []byte("could not open file")) {
			t.Errorf("unexpected error message: %v", h.Err)
		}
	})
}
