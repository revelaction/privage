package identity

import (
	"filippo.io/age"
	"fmt"
	"io/ioutil"
	"os"

	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/ascii85"

	"github.com/go-piv/piv-go/piv"
)

const PivAlgoRsa2048 = "RSA2048"

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

	retiredSlot, ok := piv.RetiredKeyManagementSlot(slot)
	if !ok {
		return fmt.Errorf("Could not access slot %x in the PIV device", slot)
	}

	cert, err := yubikey.Certificate(retiredSlot)
	if err != nil {
		return fmt.Errorf("Could not get certificate in slot %x: %v", slot, err)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, cert.PublicKey.(*rsa.PublicKey), []byte(k.String()))
	if err != nil {
		return fmt.Errorf("Could not encrypt private key: %v", err)
	}

	encoder := ascii85.NewEncoder(f)
	encoder.Write(encrypted)
	encoder.Close()

	return nil
}

// LoadPiv returns the age identity that is encrypted in the file at path,
// by decrypting it with the key at the yubikey slot slot.
func LoadPiv(path string, slot uint32, algo string) Identity {
	raw, err := LoadRaw(path, slot, algo)
	if err != nil {
		return Identity{Err: err}
	}

	identity, err := age.ParseX25519Identity(string(raw))
	if err != nil {
		return Identity{Err: fmt.Errorf("could not parse identity as age key: %v", err)}
	}

	ident := Identity{}
	ident.Id = identity
	ident.Path = path
	return ident
}

// LoadRaw returns the decrypted contents of the file at path, using the
// key in the slot slot
func LoadRaw(path string, slot uint32, algo string) ([]byte, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file at path %s: %v", path, err)

	}

	defer f.Close()

	decoded, err := ioutil.ReadAll(ascii85.NewDecoder(f))
	if err != nil {
		return nil, fmt.Errorf("could not read message file: %v", err)
	}

	yubikey, err := yubikey()
	if err != nil {
		return nil, fmt.Errorf("could not open yubikey card: %v", err)
	}
	defer yubikey.Close()

	retiredSlot, ok := piv.RetiredKeyManagementSlot(slot)
	if !ok {
		return nil, fmt.Errorf("Could not access slot %x in the PIV device", slot)
	}

	cert, err := yubikey.Certificate(retiredSlot)
	if err != nil {
		return nil, fmt.Errorf("Could not get certificate in slot %x: %v", slot, err)
	}

	// TODO Auth for list
	priv, err := yubikey.PrivateKey(retiredSlot, cert.PublicKey, piv.KeyAuth{})
	if err != nil {
		return nil, fmt.Errorf("could not setup private key: %v", err)
	}

	decrypter, ok := priv.(crypto.Decrypter)
	if !ok {
		return nil, fmt.Errorf("priv does not implement Decrypter")
	}

	decrypted, err := decrypter.Decrypt(rand.Reader, decoded, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt key file: %v", err)
	}

	return decrypted, nil
}

func yubikey() (*piv.YubiKey, error) {
	cards, err := piv.Cards()
	if err != nil {
		return nil, fmt.Errorf("Could not list cards: %v", err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("No Cards detected")
	}

	yubikey, err := piv.Open(cards[0])
	if err != nil {
		return nil, fmt.Errorf("could not open card %s: %v", cards[0], err)
	}

	return yubikey, nil
}
