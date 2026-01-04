package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/revelaction/privage/header"
)

func TestCat_Logic(t *testing.T) {
	// 1. Mock Data
	targetLabel := "target"
	secretContent := "secret payload"

	// Create a temp file to satisfy os.Open
	tmpFile := t.TempDir() + "/target.age"
	if err := os.WriteFile(tmpFile, []byte("dummy header data and content"), 0600); err != nil {
		t.Fatal(err)
	}

	// 2. Mock Scanner (HeaderStreamFunc)
	// Returns a channel with a non-matching header and then the matching one
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: "other"}
			ch <- &header.Header{Label: targetLabel, Path: tmpFile}
			close(ch)
		}()
		return ch
	}

	// 3. Mock Reader (ContentReaderFunc)
	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return strings.NewReader(secretContent), nil
	}

	// 4. Capture Output
	var buf bytes.Buffer

	// Run
	err := cat(targetLabel, mockStreamHeaders, mockReadContent, &buf)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.String() != secretContent {
		t.Errorf("expected output '%s', got '%s'", secretContent, buf.String())
	}
}

func TestCat_NotFound(t *testing.T) {
	// Mock Scanner: Returns unrelated headers
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: "other1"}
			ch <- &header.Header{Label: "other2"}
			close(ch)
		}()
		return ch
	}

	// Mock Reader: Should not be called for the target
	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return nil, errors.New("should not be called")
	}

	var buf bytes.Buffer
	err := cat("missing_label", mockStreamHeaders, mockReadContent, &buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected no output, got '%s'", buf.String())
	}
}

func TestCat_OpenError(t *testing.T) {
	targetLabel := "broken"

	// Create a temp file to satisfy os.Open
	tmpFile := t.TempDir() + "/broken.age"
	if err := os.WriteFile(tmpFile, []byte("dummy data"), 0600); err != nil {
		t.Fatal(err)
	}

	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: targetLabel, Path: tmpFile}
			close(ch)
		}()
		return ch
	}

	// Mock Reader: Returns an error
	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return nil, errors.New("decrypt failed")
	}

	var buf bytes.Buffer
	err := cat(targetLabel, mockStreamHeaders, mockReadContent, &buf)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "decrypt failed" {
		t.Errorf("expected error 'decrypt failed', got '%v'", err)
	}
}
