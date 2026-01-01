package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/revelaction/privage/header"
)

func TestCat_Logic(t *testing.T) {
	// 1. Mock Data
	targetLabel := "target"
	secretContent := "secret payload"

	// 2. Mock Scanner (HeaderStreamFunc)
	// Returns a channel with a non-matching header and then the matching one
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: "other"}
			ch <- &header.Header{Label: targetLabel}
			close(ch)
		}()
		return ch
	}

	// 3. Mock Opener (ContentOpenFunc)
	// Returns content only for the target label
	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		if h.Label == targetLabel {
			return strings.NewReader(secretContent), nil
		}
		return nil, errors.New("unexpected header")
	}

	// 4. Capture Output
	var buf bytes.Buffer

	// Run
	err := cat(targetLabel, mockStreamHeaders, mockOpenContent, &buf)

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

	// Mock Opener: Should not be called for the target
	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		return nil, errors.New("should not be called")
	}

	var buf bytes.Buffer
	err := cat("missing_label", mockStreamHeaders, mockOpenContent, &buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected no output, got '%s'", buf.String())
	}
}

func TestCat_OpenError(t *testing.T) {
	targetLabel := "broken"

	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: targetLabel}
			close(ch)
		}()
		return ch
	}

	// Mock Opener: Returns an error
	mockOpenContent := func(h *header.Header) (io.Reader, error) {
		return nil, errors.New("decrypt failed")
	}

	var buf bytes.Buffer
	err := cat(targetLabel, mockStreamHeaders, mockOpenContent, &buf)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "decrypt failed" {
		t.Errorf("expected error 'decrypt failed', got '%v'", err)
	}
}
