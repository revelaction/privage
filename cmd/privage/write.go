package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"filippo.io/age"

	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

// encryptSave encrypts the Header h and the content separately and
// concatenates both encrypted payloads.
//
// It saves the concatenated encrypted payloads on an age file. The name of the
// file is a hash of the header (label and category) and the public age key.
//
// If an error occurs during content writing, a partial file may remain on disk.
func encryptSave(h *header.Header, suffix string, content io.Reader, s *setup.Setup) (err error) {

	// Step 1: Encrypt header to memory buffer
	// This is done first because it's independent and writes to memory only.
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

	// Step 3: Create the output file
	fname, err := fileName(h, s.Id, suffix)
	if err != nil {
		return fmt.Errorf("failed to generate filename: %w", err)
	}
	filePath := s.Repository + "/" + fname

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	// Defer file close - this will run last (defers execute in LIFO order)
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	// Step 4: Write the encrypted header to file
	_, err = f.Write(headerPadded)
	if err != nil {
		return fmt.Errorf("failed to write encrypted header to file: %w", err)
	}

	// Step 5: Set up the content encryption writer stack
	// Stack: file ← bufFile (buffered) ← ageContentWr (encrypted)
	//
	// Variables declared here so they're in scope for the cleanup defer below.
	var ageContentWr io.WriteCloser
	var bufFile *bufio.Writer

	// CLEANUP DEFER: Handles proper shutdown of the writer stack.
	//
	// This defer will execute BEFORE the file close defer above (LIFO order).
	// It ensures that even if errors occur during writing, we attempt to:
	// 1. Finalize encryption (close ageContentWr)
	// 2. Flush buffered data (flush bufFile)
	//
	// Cleanup order is critical:
	//   - ageContentWr.Close() must happen first to finalize encryption and
	//     write authentication tags to bufFile
	//   - bufFile.Flush() must happen second to push buffered data to the file
	//   - f.Close() happens last (in the defer above) to sync to disk
	//
	// Error handling strategy: preserve the first error (the root cause),
	// but still attempt all cleanup steps. This means if writing fails,
	// we still try to close/flush everything to leave the file in the most
	// consistent state possible.
	defer func() {
		// Close the age encryption writer to finalize encryption
		if ageContentWr != nil {
			if cerr := ageContentWr.Close(); cerr != nil && err == nil {
				// Only capture this error if no previous error occurred
				err = fmt.Errorf("failed to close content encryptor: %w", cerr)
			}
		}

		// Flush the buffered writer to push remaining data to file
		if bufFile != nil {
			if ferr := bufFile.Flush(); ferr != nil && err == nil {
				// Only capture this error if no previous error occurred
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

	// Step 6: Stream content through the encryption stack
	bufContent := bufio.NewReader(content)

	if _, err := io.Copy(ageContentWr, bufContent); err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	// Cleanup happens in defer above - no need to explicitly close/flush here
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
	return hashStr + suffix + AgeExtension, nil
}
