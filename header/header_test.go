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
	padded, err := original.Pad()
	if err != nil {
		t.Fatalf("unexpected error during Pad(): %v", err)
	}
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

	// Scenario 3: Error (Corruption - Non-padding bytes before prefix)
	t.Run("Error_Corruption", func(t *testing.T) {
		payload := []byte("age-encryption.org/header-data")
		// Simulate corruption: padding + garbage + payload
		// Using paddingChar (space) then 'x' (garbage)
		input := append([]byte("   x   "), payload...)

		_, err := Unpad(input)
		if err == nil {
			t.Fatal("expected error for corrupted padding, got nil")
		}
		if !strings.Contains(err.Error(), "header corruption") {
			t.Errorf("unexpected error message: %v", err)
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

func TestHeader_Hash(t *testing.T) {
	h := &Header{
		Version:  "v1",
		Category: "test_cat",
		Label:    "test_label",
	}
	ageIdentity := "age1recipient"

	// Test 1: Determinism
	hash1, err := h.Hash(ageIdentity)
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}
	
	hash2, err := h.Hash(ageIdentity)
	if err != nil {
		t.Fatalf("Hash() failed second time: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash() is not deterministic: %s != %s", hash1, hash2)
	}

	// Test 2: Different inputs produce different hashes
	h2 := &Header{
		Version:  "v1",
		Category: "test_cat",
		Label:    "other_label",
	}
	hash3, err := h2.Hash(ageIdentity)
	if err != nil {
		t.Fatalf("Hash() failed for h2: %v", err)
	}

	if hash1 == hash3 {
		t.Error("Different headers produced same hash")
	}

	// Test 3: Different recipient produces different hash
	hash4, err := h.Hash("age1otherrecipient")
	if err != nil {
		t.Fatalf("Hash() failed for other recipient: %v", err)
	}
	if hash1 == hash4 {
		t.Error("Different recipients produced same hash")
	}
}

func TestHeader_PadAndParse_UTF8(t *testing.T) {
	tests := []struct {
		name     string
		category string
		label    string
		wantErr  bool
	}{
		{
			name:     "Standard ASCII",
			category: "personal",
			label:    "bank_account",
			wantErr:  false,
		},
		{
			name:     "Emoji in Label",
			category: "social",
			label:    "Instagram 📸", // 11 chars, 13 bytes (📸 is 4 bytes)
			wantErr:  false,
		},
		{
			name:     "Emoji in Category",
			category: "work💼", // 5 chars, 8 bytes (💼 is 4 bytes)
			label:    "report",
			wantErr:  false,
		},
		{
			name:     "Multi-byte Script (Chinese)",
			category: "finance",
			label:    "银行账户", // 4 chars, 12 bytes (3 bytes each)
			wantErr:  false,
		},
		{
			name:     "Mixed Script and Emojis",
			category: "旅行✈️", // Mixed
			label:    "Passport 🛂 & Ticket 🎫",
			wantErr:  false,
		},
		{
			name:     "Exact Limit Category (ASCII)",
			category: strings.Repeat("a", MaxLenghtCategory),
			label:    "limit_test",
			wantErr:  false,
		},
		{
			name:     "Exact Limit Label (ASCII)",
			category: "test",
			label:    strings.Repeat("b", MaxLenghtLabel),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{
				Version:  "v1",
				Category: tt.category,
				Label:    tt.label,
			}

			// 1. Pad
			padded, err := h.Pad()
			if (err != nil) != tt.wantErr {
				t.Errorf("Header.Pad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// 2. Verify Total Length
			// BlockSize checking is internal to Pad, but we know the sum of fields
			// version(5) + category(40) + label(200) = 245 bytes
			expectedLen := maxLenghtVersion + MaxLenghtCategory + MaxLenghtLabel
			if len(padded) != expectedLen {
				t.Errorf("Padded length = %d, want %d", len(padded), expectedLen)
			}

			// 3. Parse back
			parsed := Parse(padded)

			// 4. Verify Content matches original
			if parsed.Category != tt.category {
				t.Errorf("Category mismatch. Got %q, want %q", parsed.Category, tt.category)
			}
			if parsed.Label != tt.label {
				t.Errorf("Label mismatch. Got %q, want %q", parsed.Label, tt.label)
			}
		})
	}
}

func TestHeader_Pad_UTF8_Overflow(t *testing.T) {
	// Calculate how many 3-byte characters fit in Category (Max 40 bytes)
	// 13 chars * 3 bytes = 39 bytes (Fits)
	// 14 chars * 3 bytes = 42 bytes (Overflows)
	// Note: 14 chars would fit if we counted characters (length 14 < 40), 
	// but fails correctly because we count bytes.
	
	overflowCategory := strings.Repeat("字", 14) // 14 Chinese chars
	
h := &Header{
		Version:  "v1",
		Category: overflowCategory,
		Label:    "test",
	}

	_, err := h.Pad()
	if err == nil {
		t.Error("Header.Pad() expected error for UTF-8 byte overflow, got nil")
	} else {
		// Optional: check error message
		if !strings.Contains(err.Error(), "category exceeds maximum length") {
			t.Errorf("Unexpected error message: %v", err)
		}
	}
}

func TestHeader_Pad_Emoji_Boundary(t *testing.T) {
	// Test a label that is exactly the max length in bytes using emojis
	// MaxLenghtLabel = 200
	// 🔒 is 4 bytes. 
	// 50 * 4 = 200 bytes.
	
	exactLabel := strings.Repeat("🔒", 50)
	
h := &Header{
		Version:  "v1",
		Category: "boundary",
		Label:    exactLabel,
	}
	padded, err := h.Pad()
	if err != nil {
		t.Fatalf("Header.Pad() failed on exact boundary: %v", err)
	}
	
	parsed := Parse(padded)
	if parsed.Label != exactLabel {
		t.Errorf("Label mismatch on boundary.\nGot length: %d\nWant length: %d", len(parsed.Label), len(exactLabel))
	}
}
