package main

import (
	"fmt"
	"os"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// renameCommand renames an encrypted file by replacing its label.
//
// Because the filename is a hash of the header (label + category + public key),
// renaming requires decrypting the source file and re-encrypting it under a new
// header. The operation follows a write-then-delete order to ensure that a
// failure at any point never results in data loss:
//
//  1. Write the new encrypted file (encryptSave is itself atomic via .tmp + rename).
//  2. Remove the source file.
func renameCommand(s *setup.Setup, srcLabel string, destLabel string, ui UI) error {
	if s.Id.Id == nil {
		return fmt.Errorf("found no privage key file: %w", s.Id.Err)
	}

	if srcLabel == destLabel {
		return fmt.Errorf("source and destination labels are identical")
	}

	// Single pass: locate source and check destination does not already exist.
	var srcHeader *header.Header
	ch, err := headerGenerator(s.Repository, s.Id)
	if err != nil {
		return err
	}
	for h := range ch {
		if h.Label == srcLabel {
			srcHeader = h
		}
		if h.Label == destLabel {
			return fmt.Errorf("label %q already exists", destLabel)
		}
	}

	if srcHeader == nil {
		return fmt.Errorf("label %q not found", srcLabel)
	}

	// Open the source file; contentReader skips the encrypted header block
	// and returns a reader over the decrypted content.
	srcFile, err := os.Open(srcHeader.Path)
	if err != nil {
		return fmt.Errorf("could not open source file: %w", err)
	}
	defer srcFile.Close()

	contentR, err := contentReader(srcFile, s.Id)
	if err != nil {
		return fmt.Errorf("could not decrypt source content: %w", err)
	}

	// Build the new header: same category, new label.
	newH := &header.Header{
		Label:    destLabel,
		Category: srcHeader.Category,
	}

	// Step 1: Write new encrypted file (atomic: .tmp then rename).
	// From this point forward the data is safe regardless of what happens next.
	if err := encryptSave(newH, "", contentR, s); err != nil {
		return fmt.Errorf("could not write destination file: %w", err)
	}

	// Step 2: Remove the source file.
	if err := os.Remove(srcHeader.Path); err != nil {
		return fmt.Errorf("could not remove source file %s: %w", srcHeader.Path, err)
	}

	_, _ = fmt.Fprintf(ui.Err, "Renamed '%s' to '%s' ✔️\n", srcLabel, destLabel)
	return nil
}
