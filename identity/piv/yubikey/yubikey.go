package yubikey

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"

	"github.com/revelaction/privage/identity"
	pivlib "github.com/go-piv/piv-go/v2/piv"
)

// yubiDevice implements the piv.Device interface for YubiKey hardware.
type yubiDevice struct {
	yk *pivlib.YubiKey
}

// New returns a new YubiKey device that implements the identity.Device interface.
func New() (identity.Device, error) {
	cards, err := pivlib.Cards()
	if err != nil {
		return nil, fmt.Errorf("could not list cards: %w", err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("no cards detected")
	}

	yk, err := pivlib.Open(cards[0])
	if err != nil {
		return nil, fmt.Errorf("could not open card %s: %w", cards[0], err)
	}

	return &yubiDevice{yk: yk}, nil
}

// Decrypt decrypts ciphertext using the key in the specified slot.
func (d *yubiDevice) Decrypt(ciphertext []byte, slot uint32) ([]byte, error) {
	retiredSlot, ok := pivlib.RetiredKeyManagementSlot(slot)
	if !ok {
		return nil, fmt.Errorf("could not access slot %x in the PIV device", slot)
	}

	cert, err := d.yk.Certificate(retiredSlot)
	if err != nil {
		return nil, fmt.Errorf("could not get certificate in slot %x: %w", slot, err)
	}

	// TODO: Auth for list
	priv, err := d.yk.PrivateKey(retiredSlot, cert.PublicKey, pivlib.KeyAuth{})
	if err != nil {
		return nil, fmt.Errorf("could not setup private key: %w", err)
	}

	decrypter, ok := priv.(crypto.Decrypter)
	if !ok {
		return nil, fmt.Errorf("private key does not implement Decrypter")
	}

	decrypted, err := decrypter.Decrypt(rand.Reader, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt: %w", err)
	}

	return decrypted, nil
}

// Encrypt encrypts plaintext using the key in the specified slot and writes
// the result to w.
func (d *yubiDevice) Encrypt(w io.Writer, plaintext []byte, slot uint32) error {
	retiredSlot, ok := pivlib.RetiredKeyManagementSlot(slot)
	if !ok {
		return fmt.Errorf("could not access slot %x in the PIV device", slot)
	}

	cert, err := d.yk.Certificate(retiredSlot)
	if err != nil {
		return fmt.Errorf("could not get certificate in slot %x: %w", slot, err)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, cert.PublicKey.(*rsa.PublicKey), plaintext)
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