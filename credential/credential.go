package credential

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/atotto/clipboard"
)

const (
	// password constants
	passLenght = 25
	alphabet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	// Template is the content template of a credential.
	// credentials files are .toml files.
	Template = `# 
login = "%s"
password = "%s"

email = "%s"
url = ""

# API keys
api_key = ""
api_secret = ""
api_name = ""
api_passphrase = ""
verification_code = ""

# two factor backup code
two_factor_auth = ""
# Other fields can be put in multiline
remarks = '''
- Put here all the rest
- ....
'''
`
)

// A Credential contains all relevant information for accessing and controlling
// an online resource (password/s, api keys, 2FA backup code)
type Credential struct {
	Login            string `toml:"login"`
	Password         string `toml:"password"`
	Email            string `toml:"email"`
	ApiKey           string `toml:"api_key"`
	ApiSecret        string `toml:"api_secret"`
	ApiName          string `toml:"api_name"`
	ApiPassphrase    string `toml:"api_passphrase"`
	VerificationCode string `toml:"verification_code"`
	TwoFactorAuth    string `toml:"two_factor_auth"`
	// All other stuff here as multiline
	Remarks string `toml:"remarks"`
}

// LogFields prints in the terminal the the most important fields of the
// credential file.
func LogFields(r io.Reader) error {

	var conf Credential
	err := toml.NewDecoder(r).Decode(&conf)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("%8s%10s👤 %s\n", "", "Login:", conf.Login)
	fmt.Printf("%8s%10s🔑 %s\n", "", "Password:", conf.Password)
	fmt.Println()

	return nil
}

// CopyClipboard copies the password field to the clipboard.
func CopyClipboard(r io.Reader) error {

	var conf Credential
	err := toml.NewDecoder(r).Decode(&conf)
	if err != nil {
		return err
	}

	if err := clipboard.WriteAll(conf.Password); err != nil {
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

// ValidateFile validates a file as toml credential file.
func ValidateFile(filePath string) error {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var conf Credential
	if err := toml.Unmarshal(data, &conf); err != nil {
		return err
	}

	return nil
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
