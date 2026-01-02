package main

import (
	"io"

	"github.com/revelaction/privage/header"
)

// HeaderStreamFunc represents a function that streams headers.
// Used for injecting header scanning logic.
type HeaderStreamFunc func() <-chan *header.Header

// ContentOpenFunc represents a function that opens the content for a given header.
// Used for injecting content reading/decryption logic.
type ContentOpenFunc func(*header.Header) (io.Reader, error)

// FileCreateFunc represents a function that creates a file writer.
// Used for injecting file system write operations.
type FileCreateFunc func(name string) (io.WriteCloser, error)
