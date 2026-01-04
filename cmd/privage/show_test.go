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

func TestShow_Logic(t *testing.T) {
	// 1. Mock Data
	targetLabel := "target"
	// Valid TOML for a credential
	secretContent := `login = "user123"
password = "supersecret"
`
	// Create a temp file to satisfy os.Open
	tmpFile := t.TempDir() + "/target.age"
	if err := os.WriteFile(tmpFile, []byte("dummy data"), 0600); err != nil {
		t.Fatal(err)
	}

	// 2. Mock Scanner (HeaderStreamFunc)
	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{
				Label:    targetLabel,
				Category: header.CategoryCredential,
				Path:     tmpFile,
			}
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

	// Run - No field requested (Full Show)
	err := show(targetLabel, "", mockStreamHeaders, mockReadContent, &buf)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "user123") {
		t.Errorf("expected output to contain login 'user123', got '%s'", output)
	}

	// Run - Specific field requested
	buf.Reset()
	err = show(targetLabel, "password", mockStreamHeaders, mockReadContent, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "supersecret" {
		t.Errorf("expected 'supersecret', got '%s'", buf.String())
	}
}

func TestShow_WrongCategory(t *testing.T) {
	targetLabel := "not-a-cred"

	// Create a temp file to satisfy os.Open
	tmpFile := t.TempDir() + "/not-a-cred.age"
	if err := os.WriteFile(tmpFile, []byte("dummy data"), 0600); err != nil {
		t.Fatal(err)
	}

	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{Label: targetLabel, Category: "other", Path: tmpFile}
			close(ch)
		}()
		return ch
	}

	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return strings.NewReader("some content"), nil
	}

	var buf bytes.Buffer
	err := show(targetLabel, "", mockStreamHeaders, mockReadContent, &buf)

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

	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return nil, errors.New("should not be called")
	}

	var buf bytes.Buffer
	err := show("missing", "", mockStreamHeaders, mockReadContent, &buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected no output for missing label, got '%s'", buf.String())
	}
}

func TestShow_FieldNotFound(t *testing.T) {
	targetLabel := "target"
	secretContent := `login = "user123"`

	// Create a temp file to satisfy os.Open
	tmpFile := t.TempDir() + "/target.age"
	if err := os.WriteFile(tmpFile, []byte("dummy data"), 0600); err != nil {
		t.Fatal(err)
	}

	mockStreamHeaders := func() <-chan *header.Header {
		ch := make(chan *header.Header)
		go func() {
			ch <- &header.Header{
				Label:    targetLabel,
				Category: header.CategoryCredential,
				Path:     tmpFile,
			}
			close(ch)
		}()
		return ch
	}

	mockReadContent := func(r io.Reader) (io.Reader, error) {
		return strings.NewReader(secretContent), nil
	}

	var buf bytes.Buffer
	err := show(targetLabel, "missing_field", mockStreamHeaders, mockReadContent, &buf)

	if err == nil {
		t.Fatal("expected error for missing field, got nil")
	}
	if !strings.Contains(err.Error(), "field 'missing_field' not found") {
		t.Errorf("expected 'field not found' error, got: %v", err)
	}
}
