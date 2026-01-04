package identity

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
)

// TestLoadAge_ValidIdentity tests that LoadAge correctly parses a valid age identity
func TestLoadAge_ValidIdentity(t *testing.T) {
	// Generate a real age identity
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("Failed to generate test identity: %v", err)
	}

	// Write it to a buffer in the expected format
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# created: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&buf, "# public key: %s\n", id.Recipient())
	fmt.Fprintf(&buf, "%s\n", id)

	// Load from the buffer
	result := LoadAge(&buf, "test-path.txt")

	// Verify
	if result.Err != nil {
		t.Errorf("LoadAge returned error for valid identity: %v", result.Err)
	}
	if result.Path != "test-path.txt" {
		t.Errorf("LoadAge path mismatch: got %q, want %q", result.Path, "test-path.txt")
	}
	if result.Id == nil {
		t.Error("LoadAge returned nil identity for valid input")
	}
}

// TestLoadAge_InvalidData tests that LoadAge returns error for malformed input
func TestLoadAge_InvalidData(t *testing.T) {
	invalidData := "not a valid age key\n"
	reader := strings.NewReader(invalidData)

	result := LoadAge(reader, "invalid.txt")

	if result.Err == nil {
		t.Error("Expected error for invalid age key data, got nil")
	}
	if result.Path != "invalid.txt" {
		t.Errorf("Path mismatch: got %q, want %q", result.Path, "invalid.txt")
	}
	if result.Id != nil {
		t.Error("Expected nil identity for invalid input")
	}
}

// TestLoadAge_EmptyReader tests that LoadAge handles empty input
func TestLoadAge_EmptyReader(t *testing.T) {
	emptyReader := strings.NewReader("")

	result := LoadAge(emptyReader, "empty.txt")

	if result.Err == nil {
		t.Error("Expected error for empty reader, got nil")
	}
}

// TestGenerateAge_Success tests that GenerateAge generates and writes a valid age identity
func TestGenerateAge_Success(t *testing.T) {
	var buf bytes.Buffer

	err := GenerateAge(&buf)
	if err != nil {
		t.Fatalf("GenerateAge failed: %v", err)
	}

	// Verify the output format
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines of output, got %d", len(lines))
	}

	// Check first line has timestamp
	if !strings.HasPrefix(lines[0], "# created: ") {
		t.Errorf("First line missing timestamp prefix: %q", lines[0])
	}

	// Check second line has public key
	if !strings.HasPrefix(lines[1], "# public key: age") {
		t.Errorf("Second line missing public key prefix: %q", lines[1])
	}

	// Check third line is the private key (starts with AGE-SECRET-KEY-)
	if !strings.HasPrefix(lines[2], "AGE-SECRET-KEY-") {
		t.Errorf("Third line missing age secret key prefix: %q", lines[2])
	}

	// Verify we can parse what we generated
	reader := strings.NewReader(output)
	result := LoadAge(reader, "test-new.txt")
	if result.Err != nil {
		t.Errorf("Failed to parse generated identity: %v", result.Err)
	}
	if result.Id == nil {
		t.Error("Generated identity failed to parse back")
	}
}

// TestGenerateAge_WriterError tests that GenerateAge propagates writer errors
func TestGenerateAge_WriterError(t *testing.T) {
	// Create a writer that always fails
	errWriter := &errorWriter{}

	err := GenerateAge(errWriter)
	if err == nil {
		t.Error("Expected error from failing writer, got nil")
	}
}

// errorWriter is an io.Writer that always returns an error
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, &testError{msg: "simulated write error"}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestFmtType tests the FmtType function
func TestFmtType(t *testing.T) {
	tests := []struct {
		name     string
		slot     string
		expected string
	}{
		{
			name:     "with slot",
			slot:     "9a",
			expected: "🔏 yubikey encrypted age, slot 📌 9a",
		},
		{
			name:     "empty slot",
			slot:     "",
			expected: "🔐 age key",
		},
		{
			name:     "slot with hex",
			slot:     "9c",
			expected: "🔏 yubikey encrypted age, slot 📌 9c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FmtType(tt.slot)
			if result != tt.expected {
				t.Errorf("FmtType(%q) = %q, want %q", tt.slot, result, tt.expected)
			}
		})
	}
}

// TestBackupFilePath tests the BackupFilePath function
func TestBackupFilePath(t *testing.T) {
	// We can't easily test the exact timestamp, but we can test the format
	dir := "/tmp/backup"
	result := BackupFilePath(dir)

	// Should start with directory
	if !strings.HasPrefix(result, dir+"/") {
		t.Errorf("BackupFilePath(%q) = %q, should start with %q/", dir, result, dir)
	}

	// Should contain the default filename
	if !strings.Contains(result, DefaultFileName) {
		t.Errorf("BackupFilePath(%q) = %q, should contain %q", dir, result, DefaultFileName)
	}

	// Should end with .bak
	if !strings.HasSuffix(result, ".bak") {
		t.Errorf("BackupFilePath(%q) = %q, should end with .bak", dir, result)
	}

	// Should have timestamp in the middle
	// Format: dir/DefaultFileName-timestamp.bak
	parts := strings.Split(strings.TrimSuffix(result, ".bak"), "-")
	if len(parts) < 2 {
		t.Errorf("BackupFilePath(%q) = %q, should have timestamp after hyphen", dir, result)
	}
}

// TestBackupFilePath_EdgeCases tests edge cases for BackupFilePath
func TestBackupFilePath_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		dir  string
	}{
		{"empty dir", ""},
		{"root dir", "/"},
		{"dir with trailing slash", "/tmp/"},
		{"relative dir", "backup"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BackupFilePath(tt.dir)
			// Just ensure it doesn't panic and returns something
			if result == "" {
				t.Error("BackupFilePath returned empty string")
			}
		})
	}
}
