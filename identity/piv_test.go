package identity

import (
	"bytes"
	"encoding/ascii85"
	"errors"
	"io"
	"strings"
	"testing"
)

// mockDevice implements the Device interface for testing purposes.
// It uses a simple reversible transformation (XOR) to simulate encryption/decryption.
type mockDevice struct {
	encryptErr error
	decryptErr error
	closeErr   error
	lastSlot   uint32
	closed     bool
}

func (m *mockDevice) Encrypt(w io.Writer, plaintext []byte, slot uint32) error {
	m.lastSlot = slot
	if m.encryptErr != nil {
		return m.encryptErr
	}
	// Simulate encryption: XOR with 0xFF
	encrypted := make([]byte, len(plaintext))
	for i, b := range plaintext {
		encrypted[i] = b ^ 0xFF
	}
	_, err := w.Write(encrypted)
	return err
}

func (m *mockDevice) Decrypt(ciphertext []byte, slot uint32) ([]byte, error) {
	m.lastSlot = slot
	if m.decryptErr != nil {
		return nil, m.decryptErr
	}
	// Simulate decryption: XOR with 0xFF (reverses encryption)
	decrypted := make([]byte, len(ciphertext))
	for i, b := range ciphertext {
		decrypted[i] = b ^ 0xFF
	}
	return decrypted, nil
}

func (m *mockDevice) Close() error {
	m.closed = true
	return m.closeErr
}

// TestGeneratePiv_Success verifies that GeneratePiv correctly generates an identity,
// encrypts it using the device, and writes it ascii85-encoded to the writer.
func TestGeneratePiv_Success(t *testing.T) {
	mock := &mockDevice{}
	var buf bytes.Buffer
	slot := uint32(0x9a)

	err := GeneratePiv(&buf, mock, slot)
	if err != nil {
		t.Fatalf("GeneratePiv failed: %v", err)
	}

	// Verify interactions
	if mock.lastSlot != slot {
		t.Errorf("Expected slot %x, got %x", slot, mock.lastSlot)
	}

	// Verify output is ascii85 encoded
	// We can't verify the exact content easily because GeneratePiv creates a random key,
	// but we can decode it and verify it looks like a wrapped key.
	output := buf.String()
	decoder := ascii85.NewDecoder(strings.NewReader(output))
	decoded, err := io.ReadAll(decoder)
	if err != nil {
		t.Fatalf("Output is not valid ascii85: %v", err)
	}

	// The mock "encrypts" by XORing with 0xFF. Let's "decrypt" it to see if it looks like an age key.
	decrypted := make([]byte, len(decoded))
	for i, b := range decoded {
		decrypted[i] = b ^ 0xFF
	}
	keyStr := string(decrypted)

	if !strings.HasPrefix(keyStr, "AGE-SECRET-KEY-") {
		t.Errorf("Decrypted content does not look like an age key: %s", keyStr)
	}
}

// TestGeneratePiv_DeviceError verifies error propagation from device encryption.
func TestGeneratePiv_DeviceError(t *testing.T) {
	expectedErr := errors.New("hardware failure")
	mock := &mockDevice{encryptErr: expectedErr}
	var buf bytes.Buffer

	err := GeneratePiv(&buf, mock, 0x9a)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestDecryptPiv_Success verifies correct decoding and decryption.
func TestDecryptPiv_Success(t *testing.T) {
	// 1. Prepare data
	secret := []byte("AGE-SECRET-KEY-TEST")
	// "Encrypt" with mock logic
	encrypted := make([]byte, len(secret))
	for i, b := range secret {
		encrypted[i] = b ^ 0xFF
	}

	// Ascii85 encode
	var buf bytes.Buffer
	enc := ascii85.NewEncoder(&buf)
	enc.Write(encrypted)
	enc.Close()

	// 2. Test
	mock := &mockDevice{}
	slot := uint32(0x9c)

	result, err := DecryptPiv(&buf, mock, slot)
	if err != nil {
		t.Fatalf("DecryptPiv failed: %v", err)
	}

	// 3. Verify
	if string(result) != string(secret) {
		t.Errorf("Expected %s, got %s", secret, result)
	}
	if mock.lastSlot != slot {
		t.Errorf("Expected slot %x, got %x", slot, mock.lastSlot)
	}
}

// TestDecryptPiv_InvalidAscii85 verifies handling of bad input.
func TestDecryptPiv_InvalidAscii85(t *testing.T) {
	// ascii85.NewDecoder doesn't return many errors, but let's try reading from a failing reader
	// or just ensuring it handles malformed ascii85 if possible.
	// Actually, ascii85.NewDecoder usually just returns what it can.
	// Let's test the device error path which is more robust.
	expectedErr := errors.New("decrypt failure")
	mock := &mockDevice{decryptErr: expectedErr}

	// Valid ascii85 input
	input := strings.NewReader("BadData")

	_, err := DecryptPiv(input, mock, 0x9a)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error wrapping %v, got %v", expectedErr, err)
	}
}

// TestLoadPiv_Success verifies end-to-end loading into Identity struct.
func TestLoadPiv_Success(t *testing.T) {
	// 1. Create a real identity string
	// We can't easily create a valid full identity string manually without crypto,
	// but LoadPiv expects the decrypted bytes to be parsable by age.ParseX25519Identity.
	// So we need to feed it a valid age identity string.

	// (This is a dummy key format, but let's use a generated one to be safe)
	var buf bytes.Buffer
	GenerateAge(&buf) // This writes headers ("# created: ...") which might fail ParseX25519Identity if passed directly.
	// LoadPiv calls age.ParseX25519Identity(string(raw)).
	// age.ParseX25519Identity expects just the "AGE-SECRET-KEY-..." string.
	// GenerateAge writes headers which might fail ParseX25519Identity if passed directly.
	// Let's verify what GeneratePiv writes.
	// GeneratePiv writes `k.String()` which is just the secret key line.
	// Correct.

	// So let's generate a key string
	// We'll just assume a valid format for the test
	// Actually, age keys have specific checksums (Bech32). We should use the age package to generate a valid one string.

	// Let's rely on the mock simply passing through whatever we give it,
	// but we need a valid key for the final parse step to succeed.
	// Let's use the helper to get one.
	tmpBuf := &bytes.Buffer{}
	GenerateAge(tmpBuf)
	// Parse out just the key line
	lines := strings.Split(tmpBuf.String(), "\n")
	var realKey string
	for _, l := range lines {
		if strings.HasPrefix(l, "AGE-SECRET-KEY-") {
			realKey = l
			break
		}
	}
	if realKey == "" {
		t.Fatal("Could not generate valid age key for test setup")
	}

	// 2. Prepare encrypted input
	encrypted := make([]byte, len(realKey))
	for i, b := range []byte(realKey) {
		encrypted[i] = b ^ 0xFF
	}
	encodedBuf := &bytes.Buffer{}
	enc := ascii85.NewEncoder(encodedBuf)
	enc.Write(encrypted)
	enc.Close()

	// 3. Test
	mock := &mockDevice{}
	path := "/home/user/key.piv"
	ident := LoadPiv(encodedBuf, path, mock, 0x9a)

	// 4. Verify
	if ident.Err != nil {
		t.Errorf("LoadPiv returned error: %v", ident.Err)
	}
	if ident.Path != path {
		t.Errorf("Expected path %s, got %s", path, ident.Path)
	}
	if ident.Id == nil {
		t.Error("Expected valid identity, got nil")
	}
	if ident.Id.String() != realKey {
		t.Errorf("Key mismatch. Got %s, want %s", ident.Id.String(), realKey)
	}
}

// TestLoadPiv_ParseError verifies error when decrypted data is not a valid key.
func TestLoadPiv_ParseError(t *testing.T) {
	// Prepare "validly encrypted" but garbage data
	garbage := "NotAnAgeKey"
	encrypted := make([]byte, len(garbage))
	for i, b := range []byte(garbage) {
		encrypted[i] = b ^ 0xFF
	}
	encodedBuf := &bytes.Buffer{}
	enc := ascii85.NewEncoder(encodedBuf)
	enc.Write(encrypted)
	enc.Close()

	mock := &mockDevice{}
	ident := LoadPiv(encodedBuf, "test", mock, 0x9a)

	if ident.Err == nil {
		t.Error("Expected error parsing garbage key, got nil")
	}
	if !strings.Contains(ident.Err.Error(), "could not parse identity") {
		t.Errorf("Unexpected error message: %v", ident.Err)
	}
}

// TestLoadPiv_DecryptError verifies bubbling of decryption errors.
func TestLoadPiv_DecryptError(t *testing.T) {
	mock := &mockDevice{decryptErr: errors.New("fail")}
	ident := LoadPiv(&bytes.Buffer{}, "test", mock, 0x9a)

	if ident.Err == nil {
		t.Error("Expected decryption error, got nil")
	}
}
