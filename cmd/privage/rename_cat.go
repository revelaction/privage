package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// renameCategoryCommand renames the category of an encrypted file.
func renameCategoryCommand(s *setup.Setup, label string, newCategory string, ui UI) (err error) {
	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	// Single pass: locate source.
	var srcHeader *header.Header
	ch, genErr := headerGenerator(s.Repository, s.Id)
	if genErr != nil {
		return genErr
	}
	for h := range ch {
		if h.Label == label {
			srcHeader = h
			break
		}
	}

	if srcHeader == nil {
		return fmt.Errorf("label %q not found", label)
	}

	if srcHeader.Category == newCategory {
		return fmt.Errorf("source and destination categories are identical")
	}

	// Open the source file; contentReader skips the encrypted header block
	// and returns a reader over the decrypted content.
	srcFile, openErr := os.Open(srcHeader.Path)
	if openErr != nil {
		return fmt.Errorf("could not open source file: %w", openErr)
	}
	defer func() {
		if cerr := srcFile.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("could not close source file: %w", cerr))
		}
	}()

	contentR, crErr := contentReader(srcFile, s.Id)
	if crErr != nil {
		return fmt.Errorf("could not decrypt source content: %w", crErr)
	}

	// Build the new header: same label, new category.
	newH := &header.Header{
		Label:    srcHeader.Label,
		Category: newCategory,
	}

	// Step 1: Write new encrypted file (atomic: .tmp then rename).
	if saveErr := encryptSave(newH, "", contentR, s); saveErr != nil {
		return fmt.Errorf("could not write destination file: %w", saveErr)
	}

	// Step 2: Remove the source file.
	if remErr := os.Remove(srcHeader.Path); remErr != nil {
		return fmt.Errorf("could not remove source file %s: %w", srcHeader.Path, remErr)
	}

	_, _ = fmt.Fprintf(ui.Err, "Renamed category of '%s' to '%s' ✔️\n", label, newCategory)
	return nil
}
