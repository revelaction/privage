package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"filippo.io/age"

	"github.com/revelaction/privage/header"
	id "github.com/revelaction/privage/identity"
)

const (
	PrivageExtension = ".privage"
)

// headerGenerator iterates all .privage files in the repository directory and
// yields the decrypted header.
//
// The repository is expected to be flat; subdirectories are ignored.
func headerGenerator(repoDir string, identity id.Identity) <-chan *header.Header {

	ch := make(chan *header.Header)

	go func() {

		err := filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {

			if err != nil {
				// If WalkDir encounters an error accessing a path (e.g. permission denied),
				// returning the error here will abort the entire walk.
				return err
			}

			// Flat repository: skip subdirectories
			if d.IsDir() {
				if path != repoDir {
					return filepath.SkipDir
				}
				// In a WalkDir callback, 'return nil' is the equivalent of 'continue'
				// in a standard for loop, moving to the next entry.
				return nil
			}

			// Only process .privage files
			ext := filepath.Ext(d.Name())
			if ext != PrivageExtension {
				return nil
			}

			h := &header.Header{Path: path}

			f, err := os.Open(path)
			if err != nil {
				h.Err = fmt.Errorf("could not open file %s: %w", path, err)
				ch <- h
				return nil
			}

			// 1. Read the header
			headerBlock := make([]byte, header.BlockSize)
			_, readErr := io.ReadFull(f, headerBlock)

			// 2. Always capture the close error
			closeErr := f.Close()

			// 3. Prioritize the read error if it exists
			if readErr != nil {
				h.Err = fmt.Errorf("could not read header in file %s: %w", path, readErr)
				ch <- h
				return nil
			}

			// 4. If read succeeded, check if the close failed
			if closeErr != nil {
				h.Err = fmt.Errorf("could not close file %s: %w", path, closeErr)
				ch <- h
				return nil
			}

			// first remove the pad
			unpadded, err := header.Unpad(headerBlock)
			if err != nil {
				h.Err = fmt.Errorf("could not unpad header in file %s: %w", path, err)
				ch <- h
				return nil
			}

			uReader := bytes.NewReader(unpadded)
			r, err := age.Decrypt(uReader, identity.Id)
			if err != nil {
				h.Err = fmt.Errorf("could not Decrypt header in file %s with identity %s: %w", path, identity.Path, err)
				ch <- h
				return nil
			}

			out := &bytes.Buffer{}
			if _, err := io.Copy(out, r); err != nil {
				h.Err = fmt.Errorf("could not copy to buffer the header in file %s: %w", path, err)
				ch <- h
				return nil
			}

			h = header.Parse(out.Bytes())
			h.Path = path

			ch <- h
			return nil
		})

		if err != nil {
			// To maintain original semantics where directory-level errors were silent,
			// we acknowledge the WalkDir error but do not propagate it to the channel.
			_ = err
		}

		close(ch)
	}()

	return ch
}

// contentReader returns an `age` reader that provides the decrypted content
// from an existing reader by skipping the privage header.
func contentReader(src io.Reader, identity id.Identity) (io.Reader, error) {

	// skip header
	if _, err := io.CopyN(io.Discard, src, header.BlockSize); err != nil {
		return nil, err
	}

	return age.Decrypt(src, identity.Id)
}
