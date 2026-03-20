package main

import (
	"bytes"
	"testing"
)

func TestRenameCategoryCommand(t *testing.T) {
	th := NewTestHelper(t)
	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	// Setup: Create a file to rename category
	th.AddEncryptedFile("my_label", "old_cat", "some content")

	t.Run("Success", func(t *testing.T) {
		// Execute rename-cat
		if err := renameCategoryCommand(th.Setup, "my_label", "new_cat", ui); err != nil {
			t.Fatalf("renameCategoryCommand failed: %v", err)
		}

		// Verify old category is gone and new one exists for the label
		foundOld := false
		foundNew := false
		
		ch, err := headerGenerator(th.Root, th.Id)
		if err != nil {
			t.Fatal(err)
		}
		for h := range ch {
			if h.Label == "my_label" {
				if h.Category == "old_cat" {
					foundOld = true
				}
				if h.Category == "new_cat" {
					foundNew = true
				}
			}
		}
		
		if foundOld {
			t.Error("old category still exists after rename-cat")
		}
		if !foundNew {
			t.Error("new category not found after rename-cat")
		}
	})

	t.Run("SourceNotFound", func(t *testing.T) {
		err := renameCategoryCommand(th.Setup, "non_existent", "target_cat", ui)
		if err == nil {
			t.Fatal("expected error for missing source")
		}
	})

	t.Run("SameCategory", func(t *testing.T) {
		th.AddEncryptedFile("another_label", "same_cat", "content")
		err := renameCategoryCommand(th.Setup, "another_label", "same_cat", ui)
		if err == nil {
			t.Fatal("expected error when source and destination categories are identical")
		}
	})
}
