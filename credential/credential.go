package credential

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/atotto/clipboard"
	"github.com/pelletier/go-toml/v2"
	"github.com/revelaction/privage/config"
)

const (
	// password constants
	passLenght = 25
	alphabet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// A Credential contains all relevant information for accessing and controlling
// an online resource (password/s, api keys, 2FA backup code)
type Credential struct {
	Login            string `toml:"login" comment:"The username or handle for the service"`
	Password         string `toml:"password" comment:"The primary password"`
	Email            string `toml:"email" comment:"Associated email address"`
	Url              string `toml:"url" comment:"The website URL"`
	ApiKey           string `toml:"api_key" comment:"API Key"`
	ApiSecret        string `toml:"api_secret" comment:"API Secret"`
	ApiName          string `toml:"api_name" comment:"API Name or Description"`
	ApiPassphrase    string `toml:"api_passphrase" comment:"API Passphrase"`
	VerificationCode string `toml:"verification_code" comment:"Verification or backup codes"`
	TwoFactorAuth    string `toml:"two_factor_auth" comment:"Two-factor authentication backup code"`
	// All other stuff here as multiline
	Remarks string `toml:"remarks,multiline" comment:"Additional notes and remarks"`

	// Others captures all custom TOML keys added by the user
	Others map[string]any `toml:",inline"`
}

// New creates a new Credential with default values from config and a generated password.
func New(c *config.Config) (*Credential, error) {
	password, err := GeneratePassword()
	if err != nil {
		return nil, err
	}

	var login string
	if len(c.Login) > 0 {
		login = c.Login
	} else if len(c.Email) > 0 {
		login = c.Email
	}

	return &Credential{
		Login:    login,
		Password: password,
		Email:    c.Email,
		Remarks:  "- Put here all the rest\n- ....\n",
	}, nil
}

// Decode decodes a credential from an io.Reader.
func Decode(r io.Reader) (*Credential, error) {
	var cred Credential
	err := toml.NewDecoder(r).Decode(&cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// Encode encodes the credential to an io.Writer.
func (c *Credential) Encode(w io.Writer) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(c)
}

// Fprint prints the most important fields of the credential to an io.Writer.
func (c *Credential) Fprint(w io.Writer) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%8s%10s👤 %s\n", "", "Login:", c.Login); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%8s%10s🔑 %s\n", "", "Password:", c.Password); err != nil {
		return err
	}

	for k, v := range c.Others {
		if _, err := fmt.Fprintf(w, "%8s%10s %v\n", "", k+":", v); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return nil
}

// CopyClipboard copies the password field to the clipboard.
func CopyClipboard(r io.Reader) error {
	cred, err := Decode(r)
	if err != nil {
		return err
	}

	if err := clipboard.WriteAll(cred.Password); err != nil {
		return err
	}

	return nil
}

// EmptyClipboard deletes the content of the clipboard.
func EmptyClipboard() error {

	if err := clipboard.WriteAll(""); err != nil {
		return err
	}

	return nil
}

// Validate validates a credential from an io.Reader.
func Validate(r io.Reader) error {
	_, err := Decode(r)
	return err
}

// ValidateFile validates a file as toml credential file.
func ValidateFile(filePath string) error {

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return Validate(f)
}


// GeneratePassword generates a random password
func GeneratePassword() (string, error) {

	b := make([]byte, passLenght)

	max := big.NewInt(int64(len(alphabet)))

	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}

	return string(b), nil
}
