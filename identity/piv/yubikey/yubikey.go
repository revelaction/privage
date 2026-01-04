// Package yubikey provides YubiKey PIV device integration for privage.
//
// CRYPTOGRAPHIC ARCHITECTURE:
// Privage uses hybrid encryption - fast symmetric encryption (Age with ChaCha20-Poly1305)
// for actual file data, and asymmetric encryption (RSA) to protect the Age key itself.
//
// WHY RSA?
// - RSA is the only PIV algorithm suitable for "encrypt-once, store, decrypt-later" pattern
// - X25519 (available in YubiKey 5.7+) only supports interactive key agreement, not storage encryption
// - ECDSA (P-256/P-384) only supports signing, not encryption
// - Works across all YubiKey versions with PIV support
//
// PERFORMANCE:
// RSA is ~1000x slower than symmetric encryption, but this is acceptable because:
// - We only encrypt a 32-byte Age key, not the actual file data
// - This operation happens once per identity creation/loading
// - The ~1ms overhead is negligible compared to user entering PIN
//
// SECURITY:
// - Age key never leaves memory unencrypted except when stored on disk (encrypted by YubiKey RSA key)
// - YubiKey private key never leaves the device
// - Protects against cold boot attacks when identity file is properly secured
package yubikey

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"

	"github.com/revelaction/privage/identity"
	"github.com/go-piv/piv-go/v2/piv"
)

// yubiDevice implements the piv.Device interface for YubiKey hardware.
type yubiDevice struct {
	yk *piv.YubiKey
}

// New returns a new YubiKey device that implements the identity.Device interface.
func New() (identity.Device, error) {
	cards, err := piv.Cards()
	if err != nil {
		return nil, fmt.Errorf("could not list cards: %w", err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("no cards detected")
	}

	yk, err := piv.Open(cards[0])
	if err != nil {
		return nil, fmt.Errorf("could not open card %s: %w", cards[0], err)
	}

	return &yubiDevice{yk: yk}, nil
}

// Decrypt decrypts ciphertext using the key in the specified slot.
//
// REQUIREMENTS:
// - The slot must contain an RSA key that was used for encryption
// - The YubiKey must be present and unlocked (PIN may be required)
//
// Returns an error if the slot contains a non-RSA key or if decryption fails.
func (d *yubiDevice) Decrypt(ciphertext []byte, slot uint32) ([]byte, error) {
	retiredSlot, ok := piv.RetiredKeyManagementSlot(slot)
	if !ok {
		return nil, fmt.Errorf("could not access slot %x in the PIV device", slot)
	}

	cert, err := d.yk.Certificate(retiredSlot)
	if err != nil {
		return nil, fmt.Errorf("could not get certificate in slot %x: %w", slot, err)
	}

	// Validate that the public key is RSA (the private key must match)
	if _, ok := cert.PublicKey.(*rsa.PublicKey); !ok {
		return nil, fmt.Errorf("slot %x contains %T key, but RSA key is required. "+
			"The identity file was likely created with an RSA key - ensure the same key type is in the slot",
			slot, cert.PublicKey)
	}

	// TODO: Support KeyAuth for PIN-protected operations
	priv, err := d.yk.PrivateKey(retiredSlot, cert.PublicKey, piv.KeyAuth{})
	if err != nil {
		return nil, fmt.Errorf("could not setup private key: %w", err)
	}

	// For RSA keys, piv-go returns a crypto.Decrypter
	decrypter, ok := priv.(crypto.Decrypter)
	if !ok {
		// This should never happen for RSA keys, but provides a safety check
		return nil, fmt.Errorf("private key does not implement crypto.Decrypter (got %T). "+
			"This indicates an unexpected key type in the slot", priv)
	}

	decrypted, err := decrypter.Decrypt(rand.Reader, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt: %w (verify correct YubiKey and slot are being used)", err)
	}

	return decrypted, nil
}

// Encrypt encrypts plaintext using the key in the specified slot and writes
// the result to w.
//
// REQUIREMENTS:
// - The slot must contain an RSA key (RSA-2048, RSA-3072, or RSA-4096)
// - A valid X.509 certificate must be installed in the slot
//
// Returns an error if the slot contains a non-RSA key (e.g., ECDSA, X25519, Ed25519).
func (d *yubiDevice) Encrypt(w io.Writer, plaintext []byte, slot uint32) error {
	retiredSlot, ok := piv.RetiredKeyManagementSlot(slot)
	if !ok {
		return fmt.Errorf("could not access slot %x in the PIV device", slot)
	}

	cert, err := d.yk.Certificate(retiredSlot)
	if err != nil {
		return fmt.Errorf("could not get certificate in slot %x: %w", slot, err)
	}

	// Validate that the public key is RSA
	rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("slot %x contains %T key, but RSA key is required for encryption. "+
			"Privage only supports RSA keys because they allow 'encrypt-once, decrypt-later' pattern. "+
			"X25519 (key agreement) and ECDSA (signing only) are not suitable for this use case", 
			slot, cert.PublicKey)
	}

	// Additional validation: check RSA key size
	keySize := rsaPub.N.BitLen()
	if keySize < 2048 {
		return fmt.Errorf("RSA key in slot %x is only %d bits, minimum 2048 bits required for security", 
			slot, keySize)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, plaintext)
	if err != nil {
		return fmt.Errorf("could not encrypt private key: %w", err)
	}

	_, err = w.Write(encrypted)
	if err != nil {
		return fmt.Errorf("could not write encrypted payload: %w", err)
	}

	return nil
}

// Close closes the YubiKey device.
func (d *yubiDevice) Close() error {
	return d.yk.Close()
}
