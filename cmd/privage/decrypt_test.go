package main

import (
	"bytes"
	"errors"
	"io"
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
	repoPath := "/mock/repo"

	// 2. Mock Scanner (HeaderStreamFunc)
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: "other.txt"}
			ch <- &header.Header{Label: targetLabel}
			close(ch)
		}()
		return ch
	}

	// 3. Mock Opener (ContentOpenFunc)
	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		if h.Label == targetLabel {
			return strings.NewReader(secretContent), nil
		}
		return nil, errors.New("unexpected header")
	}

	// 4. Mock File Creator (FileCreateFunc)
	// We capture the output in a map to verify what was written to which file
	createdFiles := make(map[string]*mockWriteCloser)
	mockCreateFile := func(name string) (io.WriteCloser, error) {
		mwc := &mockWriteCloser{}
		createdFiles[name] = mwc
		return mwc, nil
	}

	// 5. Capture Stdout
	var outBuf bytes.Buffer

	// Run
	err := decrypt(targetLabel, repoPath, mockStreamHeaders, mockOpenContent, mockCreateFile, &outBuf)

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

	// Verify status output
	expectedOutput := "The file target.txt was decrypted in the directory /mock/repo."
	if !strings.Contains(outBuf.String(), expectedOutput) {
		t.Errorf("expected output to contain '%s', got '%s'", expectedOutput, outBuf.String())
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

	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		return nil, errors.New("should not be called")
	}

	mockCreateFile := func(name string) (io.WriteCloser, error) {
		return nil, errors.New("should not be called")
	}

	var outBuf bytes.Buffer
	err := decrypt("missing", "/repo", mockStreamHeaders, mockOpenContent, mockCreateFile, &outBuf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() > 0 {
		t.Errorf("expected no output, got '%s'", outBuf.String())
	}
}

func TestDecrypt_CreateError(t *testing.T) {
	targetLabel := "fail_create"

	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: targetLabel}
			close(ch)
		}()
		return ch
	}

	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		return strings.NewReader("content"), nil
	}

	// Mock Creator: Returns error
	mockCreateFile := func(name string) (io.WriteCloser, error) {
		return nil, errors.New("permission denied")
	}

	var outBuf bytes.Buffer
	err := decrypt(targetLabel, "/repo", mockStreamHeaders, mockOpenContent, mockCreateFile, &outBuf)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "permission denied" {
		t.Errorf("expected error 'permission denied', got '%v'", err)
	}
}
