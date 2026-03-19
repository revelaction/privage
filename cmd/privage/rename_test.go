package main

import (
	"bytes"
	"testing"
)

func TestRenameCommand(t *testing.T) {
	th := NewTestHelper(t)
	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	// Setup: Create a file to rename
	th.AddEncryptedFile("old_label", "category", "some content")

	t.Run("Success", func(t *testing.T) {
		// Execute rename
		if err := renameCommand(th.Setup, "old_label", "new_label", ui); err != nil {
			t.Fatalf("renameCommand failed: %v", err)
		}

		// Verify old file is gone and new one exists
		// We use headerGenerator to scan the repository as filenames are hashed
		foundOld := false
		foundNew := false
		
		ch, err := headerGenerator(th.Root, th.Setup.Id)
		if err != nil {
			t.Fatal(err)
		}
		for h := range ch {
			if h.Label == "old_label" {
				foundOld = true
			}
			if h.Label == "new_label" {
				foundNew = true
			}
		}
		
		if foundOld {
			t.Error("old_label still exists after rename")
		}
		if !foundNew {
			t.Error("new_label not found after rename")
		}
	})

	t.Run("SourceNotFound", func(t *testing.T) {
		err := renameCommand(th.Setup, "non_existent", "target", ui)
		if err == nil {
			t.Fatal("expected error for missing source")
		}
	})

	t.Run("DestinationExists", func(t *testing.T) {
		th.AddEncryptedFile("src_label", "cat", "content")
		th.AddEncryptedFile("dst_label", "cat", "content")

		err := renameCommand(th.Setup, "src_label", "dst_label", ui)
		if err == nil {
			t.Fatal("expected error when destination exists")
		}
	})

	t.Run("SameLabel", func(t *testing.T) {
		err := renameCommand(th.Setup, "same", "same", ui)
		if err == nil {
			t.Fatal("expected error when source and destination are identical")
		}
	})
}
