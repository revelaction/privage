package main

import (
	"bufio"
	"bytes"
	"errors"
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
		// Join the error from Close (if any) with the padding error
		if closeErr := ageWr.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close header encryptor: %w", closeErr))
		}
		return fmt.Errorf("failed to pad header: %w", err)
	}

	_, err = ageWr.Write(headerBytes)
	if err != nil {
		// Join the error from Close (if any) with the write error
		if closeErr := ageWr.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close header encryptor: %w", closeErr))
		}
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

	// DEFER 1 (executes LAST): Cleanup temp file on error
	// Fix: Defined BEFORE Rename, so it executes AFTER Rename.
	// If defer Rename fails, it updates 'err', and this defer will see the error and remove the file.
	// If Rename success, it sees no error and does nothing. This is correct because the temp file is already gone.
	defer func() {
		if err != nil {
			if remErr := os.Remove(tmpPath); remErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to remove temp file: %w", remErr))
			}
		}
	}()

	// DEFER 2 (executes THIRD): Atomic rename if successful
	// Only runs if no errors yet.
	// This ensures that if os.Rename fails, the 'err' is captured, and Defer 1 cleans up.
	defer func() {
		if err == nil {
			if rerr := os.Rename(tmpPath, finalPath); rerr != nil {
				err = fmt.Errorf("failed to rename temp file: %w", rerr)
			}
		}
	}()

	// DEFER 3 (executes SECOND): Close file
	// If an error already exists, we keep it AND add the close error.
	// If err is nil, Join(nil, cerr) sets err to cerr.
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close file: %w", cerr))
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
	// Using Join to capture flush/close errors without losing main errors.
	defer func() {
		if ageContentWr != nil {
			if cerr := ageContentWr.Close(); cerr != nil {
				err = errors.Join(err, fmt.Errorf("failed to close content encryptor: %w", cerr))
			}
		}

		if bufFile != nil {
			if ferr := bufFile.Flush(); ferr != nil {
				err = errors.Join(err, fmt.Errorf("failed to flush buffered writer: %w", ferr))
			}
		}
	}()

	// Create the writer stack for content
	bufFile = bufio.NewWriter(f)

	ageContentWr, err = age.Encrypt(bufFile, s.Id.Id.Recipient())
	if err != nil {
		return fmt.Errorf("failed to create age encryptor for content: %w", err)
	}

	// Step 7: Stream content
	bufContent := bufio.NewReader(content)
	if _, err := io.Copy(ageContentWr, bufContent); err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	return nil
}

// fileName generates the file name of a privage encrypted file.
// The hash is a function of the header and the age public key.
func fileName(h *header.Header, identity id.Identity, suffix string) (string, error) {
	hashStr, err := h.Hash(identity.Id.Recipient().String())
	if err != nil {
		return "", fmt.Errorf("failed to generate header hash: %w", err)
	}
	return hashStr + suffix + PrivageExtension, nil
}
