package identity

import (
	"filippo.io/age"
	"fmt"
	"io"
	"os"

	"crypto/rand"
	"crypto/rsa"
	"encoding/ascii85"

	pivlib "github.com/go-piv/piv-go/v2/piv"
)

const PivAlgoRsa2048 = "RSA2048"

// Device represents a PIV-compatible hardware device that can perform
// cryptographic operations like decryption.
type Device interface {
	// Decrypt decrypts ciphertext using the key in the specified slot.
	Decrypt(ciphertext []byte, slot uint32) ([]byte, error)
	// Close releases any resources associated with the device.
	Close() error
}

// CreatePivRsa generates an age secret key and encrypts it using the yubikey
// key at slot slot.
//
// It writes the encrypted payload in filePath.
func CreatePivRsa(filePath string, slot uint32, algo string) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()

	k, err := age.GenerateX25519Identity()
	if err != nil {
		return fmt.Errorf("could not generate age identity: %w", err)
	}

	yubikey, err := yubikey()
	if err != nil {
		return fmt.Errorf("could not get yubikey: %v", err)
	}
	defer yubikey.Close()

	retiredSlot, ok := pivlib.RetiredKeyManagementSlot(slot)
	if !ok {
		return fmt.Errorf("could not access slot %x in the PIV device", slot)
	}

	cert, err := yubikey.Certificate(retiredSlot)
	if err != nil {
		return fmt.Errorf("could not get certificate in slot %x: %v", slot, err)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, cert.PublicKey.(*rsa.PublicKey), []byte(k.String()))
	if err != nil {
		return fmt.Errorf("could not encrypt private key: %v", err)
	}

	encoder := ascii85.NewEncoder(f)
	encoder.Write(encrypted)
	encoder.Close()

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

func yubikey() (*pivlib.YubiKey, error) {
	cards, err := pivlib.Cards()
	if err != nil {
		return nil, fmt.Errorf("could not list cards: %v", err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("no Cards detected")
	}

	yubikey, err := pivlib.Open(cards[0])
	if err != nil {
		return nil, fmt.Errorf("could not open card %s: %v", cards[0], err)
	}

	return yubikey, nil
}
