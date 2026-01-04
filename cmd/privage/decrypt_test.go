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

// mockWriteCloser implements io.WriteCloser for testing.
// It writes to the embedded bytes.Buffer and offers a no-op Close.
type mockWriteCloser struct {
	bytes.Buffer
}

func (mwc *mockWriteCloser) Close() error {
	return nil
}

func TestDecrypt_Logic(t *testing.T) {
	// 1. Mock Data
	targetLabel := "target.txt"
	secretContent := "decrypted payload"

	// Create a temp file to satisfy os.Open
	tmpFile := t.TempDir() + "/target.age"
	if err := os.WriteFile(tmpFile, []byte("dummy data"), 0600); err != nil {
		t.Fatal(err)
	}

	// 2. Mock Scanner (HeaderStreamFunc)
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: "other.txt"}
			ch <- &header.Header{Label: targetLabel, Path: tmpFile}
			close(ch)
		}()
		return ch
	}

	// 3. Mock Reader (ContentReaderFunc)
	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return strings.NewReader(secretContent), nil
	}

	// 4. Mock File Creator (FileCreateFunc)
	// We capture the output in a map to verify what was written to which file
	createdFiles := make(map[string]*mockWriteCloser)
	mockCreateFile := func(name string) (io.WriteCloser, error) {
		mwc := &mockWriteCloser{}
		createdFiles[name] = mwc
		return mwc, nil
	}

	// 5. Capture Stdout - No longer needed for logic test as decrypt is silent on success
	// var outBuf bytes.Buffer

	// Run
	err := decrypt(targetLabel, mockStreamHeaders, mockReadContent, mockCreateFile)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file content
	if _, ok := createdFiles[targetLabel]; !ok {
		t.Fatalf("file '%s' was not created", targetLabel)
	}
	if createdFiles[targetLabel].String() != secretContent {
		t.Errorf("expected file content '%s', got '%s'", secretContent, createdFiles[targetLabel].String())
	}
}

func TestDecrypt_NotFound(t *testing.T) {
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: "other1"}
			close(ch)
		}()
		return ch
	}

	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return nil, errors.New("should not be called")
	}

	mockCreateFile := func(name string) (io.WriteCloser, error) {
		return nil, errors.New("should not be called")
	}

	err := decrypt("missing", mockStreamHeaders, mockReadContent, mockCreateFile)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrHeaderNotFound) {
		t.Errorf("expected ErrHeaderNotFound, got %v", err)
	}
}

func TestDecrypt_CreateError(t *testing.T) {
	targetLabel := "fail_create"

	// Create a temp file to satisfy os.Open
	tmpFile := t.TempDir() + "/fail_create.age"
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

	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return strings.NewReader("content"), nil
	}

	// Mock Creator: Returns error
	mockCreateFile := func(name string) (io.WriteCloser, error) {
		return nil, errors.New("permission denied")
	}

	err := decrypt(targetLabel, mockStreamHeaders, mockReadContent, mockCreateFile)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "permission denied" {
		t.Errorf("expected error 'permission denied', got '%v'", err)
	}
}
