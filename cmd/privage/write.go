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
func encryptSave(h *header.Header, suffix string, content io.Reader, s *setup.Setup) (err error) {

	// age io.Writer
	buf := new(bytes.Buffer)
	ageWr, err := age.Encrypt(buf, s.Id.Id.Recipient())
	if err != nil {
		fatal(err)
	}

	_, err = ageWr.Write(h.Pad())
	if err != nil {
		return err
	}

	if err = ageWr.Close(); err != nil {
		fatal(err)
	}

	headerPadded, _ := header.PadEncrypted(buf.Bytes()) // []byte

	filePath := s.Repository + "/" + fileName(h, s.Id, suffix)

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = f.Write(headerPadded)
	if err != nil {
		return err
	}

	bufContent := bufio.NewReader(content)

	bufFile := bufio.NewWriter(f)

	ageContentWr, err := age.Encrypt(bufFile, s.Id.Id.Recipient())
	if err != nil {
		fatal(err)
	}

	if _, err := io.Copy(ageContentWr, bufContent); err != nil {
		fatal(err)
	}

	if err = ageContentWr.Close(); err != nil {
		fatal(err)
	}

	if err = bufFile.Flush(); err != nil {
		return err
	}

	return nil
}

// fileName generates the file name of a privage encrypted file.
// The hash is a function of the header and the age public key.
func fileName(h *header.Header, identity id.Identity, suffix string) string {
	hash := append(h.Pad(), identity.Id.Recipient().String()...)
	hashStr := fmt.Sprintf("%x", sha256.Sum256(hash))
	return hashStr + suffix + AgeExtension
}

func fatal(err error) {
	// TODO: Remove fatal() and refactor encryptSave to return errors properly.
	_, _ = fmt.Fprintf(os.Stderr, "privage: %v\n", err)
	os.Exit(1)
}
