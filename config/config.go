package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	// DefaultFileName is the default name for the privage configuration file.
	DefaultFileName = ".privage.conf"
)

// A Config contains configuration data for the privage application.
type Config struct {
	// Path is the filesystem path where the config was loaded from.
	Path string `toml:"-"`

	// Identity settings
	IdentityPath    string `toml:"identity_path" comment:"Path to the age identity file (supports ~/)"`
	IdentityType    string `toml:"identity_type" comment:"Type of identity: AGE or PIV"`
	IdentityPivSlot string `toml:"identity_piv_slot" comment:"Hex string for the Yubikey PIV slot (e.g., 9a)"`
	IdentityPivAlgo string `toml:"identity_piv_algo" comment:"[Reserved] Algorithm for PIV"`

	// Repository settings
	RepositoryPath string `toml:"repository_path" comment:"Directory containing encrypted files (supports ~/)"`

	// Default fields for credentials
	Login string `toml:"login" comment:"Default username/login for new credentials"`
	Email string `toml:"email" comment:"Default email for new credentials"`
}

// Decode decodes a configuration from an io.Reader.
func Decode(r io.Reader) (*Config, error) {
	var conf Config
	dec := toml.NewDecoder(r)
	if err := dec.Decode(&conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

// Encode encodes the configuration to an io.Writer.
func (c *Config) Encode(w io.Writer) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(c)
}

// Validate ensures that the configuration has all required fields and that
// referenced paths exist.
func (c *Config) Validate() error {
	if c.IdentityPath == "" {
		return errors.New("identity_path is required")
	}
	if !fileExists(c.IdentityPath) {
		return fmt.Errorf("identity file %s does not exist", c.IdentityPath)
	}

	if c.RepositoryPath == "" {
		return errors.New("repository_path is required")
	}
	if !fileExists(c.RepositoryPath) {
		return fmt.Errorf("repository directory %s does not exist", c.RepositoryPath)
	}

	return nil
}

// ExpandHome replaces "~/" prefixes in paths with the user's home directory.
func (c *Config) ExpandHome() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.New("could not determine user home directory")
	}

	if strings.HasPrefix(c.IdentityPath, "~/") {
		c.IdentityPath = filepath.Join(home, c.IdentityPath[2:])
	}

	if strings.HasPrefix(c.RepositoryPath, "~/") {
		c.RepositoryPath = filepath.Join(home, c.RepositoryPath[2:])
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}