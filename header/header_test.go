package header

import (
	"bytes"
	"strings"
	"testing"
)

func TestHeader_PadAndParse(t *testing.T) {
	// 1. Setup Data
	original := &Header{
		Version:  "v1",
		Category: "test_cat",
		Label:    "test_label",
	}

	// 2. Action
	padded := original.Pad()
	parsed := Parse(padded)

	// 3. Assert
	if parsed.Version != original.Version {
		t.Errorf("expected Version '%s', got '%s'", original.Version, parsed.Version)
	}
	if parsed.Category != original.Category {
		t.Errorf("expected Category '%s', got '%s'", original.Category, parsed.Category)
	}
	if parsed.Label != original.Label {
		t.Errorf("expected Label '%s', got '%s'", original.Label, parsed.Label)
	}
}

func TestPadEncrypted_Success(t *testing.T) {
	// 1. Setup Data
	input := []byte("short payload")

	// 2. Action
	output, err := PadEncrypted(input)

	// 3. Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output) != BlockSize {
		t.Errorf("expected length %d, got %d", BlockSize, len(output))
	}
	// Verify padding (spaces)
	expectedPaddingLen := BlockSize - len(input)
	padding := output[:expectedPaddingLen]
	if !bytes.Equal(padding, bytes.Repeat([]byte{0x20}, expectedPaddingLen)) {
		t.Error("padding did not consist of expected 0x20 bytes")
	}
	// Verify payload
	if !bytes.Equal(output[expectedPaddingLen:], input) {
		t.Error("payload was not preserved at the end of the block")
	}
}

func TestUnpad_Logic(t *testing.T) {
	// Scenario 1: Success
	t.Run("Success", func(t *testing.T) {
		// 1. Setup Data
		payload := []byte("age-encryption.org/header-data")
		// Simulate padding: spaces + payload
		input := append(bytes.Repeat([]byte{0x20}, 10), payload...)

		// 2. Action
		output, err := Unpad(input)

		// 3. Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(output, payload) {
			t.Errorf("expected '%s', got '%s'", payload, output)
		}
	})

	// Scenario 2: Error (Prefix missing)
	t.Run("Error_PrefixMissing", func(t *testing.T) {
		input := []byte("       invalid-header-data")
		_, err := Unpad(input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestHeader_String(t *testing.T) {
	// 1. Setup Data
	credHeader := &Header{Category: CategoryCredential, Label: "MyCred"}
	fileHeader := &Header{Category: "file", Label: "MyFile"}

	// 2. Assert
	if !strings.Contains(credHeader.String(), "📝") {
		t.Errorf("expected credential icon, got %s", credHeader.String())
	}
	if !strings.Contains(fileHeader.String(), "💼") {
		t.Errorf("expected file icon, got %s", fileHeader.String())
	}
}
