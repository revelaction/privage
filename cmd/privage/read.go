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

const AgeExtension = ".age"

// headerGenerator iterates all .age file in the repository directory and
// yields the decrypted header.
func headerGenerator(repoDir string, identity id.Identity) <-chan *header.Header {

	ch := make(chan *header.Header)

	go func() {
		for _, path := range filesWithExt(repoDir, AgeExtension) {
			h := &header.Header{}
			h.Path = path

			f, err := os.Open(path)
			if err != nil {
				h.Err = fmt.Errorf("could not open file %s: %w", path, err)

				ch <- h
				continue
			}

			defer f.Close()

			// Header
			headerBlock := make([]byte, header.BlockSize)
			_, err = io.ReadFull(f, headerBlock)
			if err != nil {
				h.Err = fmt.Errorf("could not read header in file %s: %w", path, err)
				ch <- h
				continue
			}

			// first remove the pad
			unpadded, err := header.Unpad(headerBlock)
			if err != nil {
				h.Err = fmt.Errorf("could not unpad header in file %s: %w", path, err)
				ch <- h
				continue
			}

			uReader := bytes.NewReader(unpadded)
			r, err := age.Decrypt(uReader, identity.Id)
			if err != nil {
				h.Err = fmt.Errorf("could not Decrypt header in file %s with identity %s: %w", path, identity.Path, err)
				ch <- h
				continue
			}

			out := &bytes.Buffer{}
			if _, err := io.Copy(out, r); err != nil {
				h.Err = fmt.Errorf("could not copy to buffer the header in file %s: %w", path, err)
				ch <- h
				continue
			}

			h = header.Parse(out.Bytes())
			h.Path = path

			ch <- h
		}

		close(ch)
	}()

	return ch
}

// contentReader returns an `age` reader that provides the decrypted content
func contentReader(h *header.Header, identity id.Identity) (io.Reader, error) {

	f, _ := os.Open(h.Path)

	// skip header
	_, err := f.Seek(header.BlockSize, io.SeekStart)
	if err != nil {
		return nil, err
	}

	r, err := age.Decrypt(f, identity.Id)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// filesWithExt returns all files with extension ext under the directory root.
func filesWithExt(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}
