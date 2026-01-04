package identity

import (
	"fmt"
	"io"
	"time"

	"filippo.io/age"
)

const (

	// default privage secret key. It can be an age key of a PIV encoded age
	// key.
	DefaultFileName = "privage-key.txt"
	TypePiv         = "PIV"
	TypeAge         = "AGE"
)

// An Identity is a wrapper for the age Identity.
type Identity struct {

	// The age identity
	Id *age.X25519Identity

	// Path of the found key.
	// Path can contain a normal age key or a PIV encrypted one.
	//
	// Path can be not empty and still a null Id because of a decoding error.
	//
	// A empty Path means all possible paths were searched and no files were
	// found
	Path string

	// Err is the error raised finding or validating the a age identity.
	Err error
}

// LoadAge returns an Age identity from an io.Reader.
// The path parameter is used for error messages and tracking.
func LoadAge(r io.Reader, path string) Identity {
	return parseIdentity(r, path)
}

func parseIdentity(f io.Reader, path string) Identity {

	identity := Identity{}
	identity.Path = path

	identities, err := age.ParseIdentities(f)
	if err != nil {
		identity.Err = err
	} else if id, ok := identities[0].(*age.X25519Identity); ok {
		identity.Id = id
	} else {
		identity.Err = fmt.Errorf("expected X25519Identity, got %T", identities[0])
	}

	return identity
}

func FmtType(slot string) string {

	if len(slot) > 0 {
		return fmt.Sprintf("🔏 yubikey encrypted age, slot 📌 %s", slot)

	}
	return "🔐 age key"
}

// GenerateAge generates an age Identity and writes it to the writer.
func GenerateAge(w io.Writer) error {
	k, err := age.GenerateX25519Identity()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "# created: %s\n", time.Now().Format(time.RFC3339)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "# public key: %s\n", k.Recipient()); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\n", k); err != nil {
		return err
	}
	return nil
}

// BackupFilePath returns a path for a backup identity file.
func BackupFilePath(dir string) string {

	now := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf("%s/%s-%s.bak", dir, DefaultFileName, now)
}
