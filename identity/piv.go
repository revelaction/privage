package identity

import (
	"fmt"
	"io"

	"encoding/ascii85"

	"filippo.io/age"
)

const PivAlgoRsa2048 = "RSA2048"

// Device represents a PIV-compatible hardware device that can perform
// cryptographic operations like decryption.
type Device interface {
	// Decrypt decrypts ciphertext using the key in the specified slot.
	Decrypt(ciphertext []byte, slot uint32) ([]byte, error)
	// Encrypt encrypts plaintext using the key in the specified slot and writes
	// the result to w.
	Encrypt(w io.Writer, plaintext []byte, slot uint32) error
	// Close releases any resources associated with the device.
	Close() error
}

// GeneratePiv generates a new age identity, encrypts it using the PIV device
// at the specified slot, and writes the ascii85-encoded result to w.
func GeneratePiv(w io.Writer, device Device, slot uint32) error {
	k, err := age.GenerateX25519Identity()
	if err != nil {
		return fmt.Errorf("could not generate age identity: %w", err)
	}

	encoder := ascii85.NewEncoder(w)
	defer encoder.Close()

	if err := device.Encrypt(encoder, []byte(k.String()), slot); err != nil {
		return fmt.Errorf("could not encrypt identity: %w", err)
	}

	return nil
}

// LoadPiv returns the age identity read from r that is encrypted with PIV.
// The path parameter is used for error messages and tracking (no filesystem operations).
// TODO: Revisit signature - consider whether path should be part of Identity struct.
func LoadPiv(r io.Reader, path string, device Device, slot uint32) Identity {
	raw, err := LoadRaw(r, device, slot)
	if err != nil {
		return Identity{Err: err}
	}

	identity, err := age.ParseX25519Identity(string(raw))
	if err != nil {
		return Identity{Err: fmt.Errorf("could not parse identity as age key: %w", err)}
	}

	ident := Identity{}
	ident.Id = identity
	ident.Path = path
	return ident
}

// LoadRaw returns the decrypted contents read from r, using the
// PIV device to decrypt with the key in the specified slot.
func LoadRaw(r io.Reader, device Device, slot uint32) ([]byte, error) {
	decoded, err := io.ReadAll(ascii85.NewDecoder(r))
	if err != nil {
		return nil, fmt.Errorf("could not read message file: %w", err)
	}

	decrypted, err := device.Decrypt(decoded, slot)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt key file: %w", err)
	}

	return decrypted, nil
}
