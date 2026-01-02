package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/revelaction/privage/header"
)

func TestShow_Logic(t *testing.T) {
	// 1. Mock Data
	targetLabel := "target"
	// Valid TOML for a credential
	secretContent := `login = "user123"
password = "supersecret"
`

	// 2. Mock Scanner (HeaderStreamFunc)
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: targetLabel, Category: header.CategoryCredential}
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

	// 4. Capture Output
	var buf bytes.Buffer

	// Run
	err := show(targetLabel, mockStreamHeaders, mockOpenContent, &buf)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "user123") {
		t.Errorf("expected output to contain login 'user123', got '%s'", output)
	}
	if !strings.Contains(output, "supersecret") {
		t.Errorf("expected output to contain password 'supersecret', got '%s'", output)
	}
}

func TestShow_WrongCategory(t *testing.T) {
	targetLabel := "not-a-cred"

	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: targetLabel, Category: "other"}
			close(ch)
		}()
		return ch
	}

	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		return strings.NewReader("some content"), nil
	}

	var buf bytes.Buffer
	err := show(targetLabel, mockStreamHeaders, mockOpenContent, &buf)

	if err == nil {
		t.Fatal("expected error for non-credential category, got nil")
	}
	if !strings.Contains(err.Error(), "is not a credential") {
		t.Errorf("expected 'is not a credential' error, got: %v", err)
	}
}

func TestShow_NotFound(t *testing.T) {
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: "other"}
			close(ch)
		}()
		return ch
	}

	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		return nil, errors.New("should not be called")
	}

	var buf bytes.Buffer
	err := show("missing", mockStreamHeaders, mockOpenContent, &buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected no output for missing label, got '%s'", buf.String())
	}
}
