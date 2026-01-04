package main

import (
	"io"

	"github.com/revelaction/privage/header"
)

// HeaderStreamFunc represents a function that streams headers.
// Used for injecting header scanning logic.
type HeaderStreamFunc func() <-chan *header.Header

// ContentReaderFunc represents a function that returns a reader for the content
// from an existing reader. Used for injecting content decryption logic.
type ContentReaderFunc func(io.Reader) (io.Reader, error)

// FileCreateFunc represents a function that creates a file writer.
// Used for injecting file system write operations.
type FileCreateFunc func(name string) (io.WriteCloser, error)
