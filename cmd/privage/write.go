package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"

	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// encryptSave encrypts the Header h and the content separately and
// concatenates both encrypted payloads.
//
// It saves the concatenated encrypted payloads on an age file atomically.
// The name of the file is a hash of the header (label and category) and the 
// public age key.
//
// Uses atomic write pattern: writes to temp file, then renames on success.
func encryptSave(h *header.Header, suffix string, content io.Reader, s *setup.Setup) (err error) {

	// Step 1: Encrypt header to memory buffer
	buf := new(bytes.Buffer)
	ageWr, err := age.Encrypt(buf, s.Id.Id.Recipient())
	if err != nil {
		return fmt.Errorf("failed to create age encryptor for header: %w", err)
	}

	headerBytes, err := h.Pad()
	if err != nil {
		return fmt.Errorf("failed to pad header: %w", err)
	}

	_, err = ageWr.Write(headerBytes)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if err = ageWr.Close(); err != nil {
		return fmt.Errorf("failed to close header encryptor: %w", err)
	}

	// Step 2: Pad the encrypted header to fixed size
	headerPadded, err := header.PadEncrypted(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to pad encrypted header: %w", err)
	}

	// Step 3: Generate final and temporary file paths
	fname, err := fileName(h, s.Id, suffix)
	if err != nil {
		return fmt.Errorf("failed to generate filename: %w", err)
	}
	finalPath := filepath.Join(s.Repository, fname)
	tmpPath := finalPath + ".tmp"

	// Step 4: Create temporary file
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temp file %s: %w", tmpPath, err)
	}

	// DEFER 1 (executes LAST): Atomic rename if successful
	// This runs after everything is closed and flushed
	defer func() {
		if err == nil {
			if rerr := os.Rename(tmpPath, finalPath); rerr != nil {
				err = fmt.Errorf("failed to rename temp file: %w", rerr)
			}
		}
	}()

	// DEFER 2 (executes THIRD): Cleanup temp file on error
	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	// DEFER 3 (executes SECOND): Close file
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	// Step 5: Write the encrypted header to temp file
	_, err = f.Write(headerPadded)
	if err != nil {
		return fmt.Errorf("failed to write encrypted header to file: %w", err)
	}

	// Step 6: Set up the content encryption writer stack
	var ageContentWr io.WriteCloser
	var bufFile *bufio.Writer

	// DEFER 4 (executes FIRST): Close age writer and flush buffer
	// This ensures cleanup happens on ANY error path
	defer func() {
		if ageContentWr != nil {
			if cerr := ageContentWr.Close(); cerr != nil && err == nil {
				err = fmt.Errorf("failed to close content encryptor: %w", cerr)
			}
		}

		if bufFile != nil {
			if ferr := bufFile.Flush(); ferr != nil && err == nil {
				err = fmt.Errorf("failed to flush buffered writer: %w", ferr)
			}
		}
	}()

	// Create the writer stack for content
	bufFile = bufio.NewWriter(f)

	ageContentWr, err = age.Encrypt(bufFile, s.Id.Id.Recipient())
	if err != nil {
		return fmt.Errorf("failed to create age encryptor for content: %w", err)
	}

	// Step 7: Stream content through the encryption stack
	bufContent := bufio.NewReader(content)

	if _, err := io.Copy(ageContentWr, bufContent); err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	// Step 8: Return and let defers handle everything
	// Execution order when returning:
	//   1. Close age writer (finalize encryption)
	//   2. Flush buffer (push to file)
	//   3. Close file (sync to disk)
	//   4. Cleanup temp (if err != nil) OR Rename (if err == nil)
	return nil
}

// fileName generates the file name of a privage encrypted file.
// The hash is a function of the header and the age public key.
func fileName(h *header.Header, identity id.Identity, suffix string) (string, error) {
	padded, err := h.Pad()
	if err != nil {
		return "", fmt.Errorf("failed to pad header: %w", err)
	}
	hash := append(padded, identity.Id.Recipient().String()...)
	hashStr := fmt.Sprintf("%x", sha256.Sum256(hash))
	return hashStr + suffix + PrivageExtension, nil
}
