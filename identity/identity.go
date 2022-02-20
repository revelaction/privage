package identity

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"filippo.io/age"
)

const (

	// default privage secret key. It can be an age key of a PIV encoded age
	// key.
	FileName = "privage-key.txt"
	TypePiv  = "PIV"
	TypeAge  = "AGE"
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

// Load returns an Age identity
//
// it tries:
// 1) the path given in the config file
// 2) FILE in the current dir
// 3) FILE in the user HOME
// init method check if exist
func Load(confPath string) Identity {

	if len(confPath) > 0 {
		ff, err := os.Open(confPath)
		defer ff.Close()
		if err == nil {
			return parseIdentity(ff, confPath)
		}

		// no Path
		return Identity{Err: err}
	}

	// try current dir
	currentDir, err := os.Getwd()
	if err != nil {
		return Identity{Err: err}
	}

	currentPath := currentDir + "/" + FileName
	fl, err := os.Open(currentPath)
	defer fl.Close()
	if err == nil {
		return parseIdentity(fl, currentPath)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return Identity{Err: err}
	}

	// try home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Identity{Err: err}
	}

	homePath := homeDir + "/" + FileName
	f, err := os.Open(homePath)
	defer f.Close()
	if err == nil {
		return parseIdentity(f, homePath)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return Identity{Err: err}
	}

	return Identity{Err: err}
}

func parseIdentity(f io.Reader, path string) Identity {

	identity := Identity{}
	identity.Path = path

	identities, err := age.ParseIdentities(f)
	if err != nil {
		identity.Err = err
	} else {
		identity.Id = identities[0].(*age.X25519Identity)
	}

	return identity
}

func FmtType(slot string) string {

	if len(slot) > 0 {
		return fmt.Sprintf("🔏 yubikey encrypted age, slot 📌 %s", slot)

	}
	return "🔐 age key"
}

// Create generates a age Identity and writes it in the file at filePath
func Create(filePath string) error {

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
		return err
	}

	fmt.Fprintf(f, "# created: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "# public key: %s\n", k.Recipient())
	fmt.Fprintf(f, "%s\n", k)
	return nil
}

// BackupFilePath returns a path for a backup identity file.
func BackupFilePath(dir string) string {

	now := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf("%s/%s-%s.bak", dir, FileName, now)
}
