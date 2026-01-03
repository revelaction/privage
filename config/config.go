package config

import (
	"bufio"
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

	// Identity types
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

// New reads, parses, and validates a configuration file from the given path.
func New(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf, err := Decode(bufio.NewReader(f))
	if err != nil {
		return nil, fmt.Errorf("invalid configuration file %s: %w", path, err)
	}

	if err := conf.ExpandHome(); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed for %s: %w", path, err)
	}

	conf.Path = path
	return conf, nil
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
		return errors.New("secrets_repository_path is required")
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

// FindPath searches for the configuration file in standard locations.
func FindPath() (string, error) {
	locations := []func() (string, error){
		os.UserHomeDir,
		os.Getwd,
	}

	for _, getDir := range locations {
		dir, err := getDir()
		if err != nil {
			continue
		}

		path := filepath.Join(dir,DefaultFileName)
		if fileExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not find configuration file %s in home or current directory",DefaultFileName)
}

// Create generates a new configuration file at the default home location.
func Create(identityPath, identityType, identityPivSlot, repositoryPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	conf := &Config{
		IdentityPath:    identityPath,
		IdentityType:    identityType,
		IdentityPivSlot: identityPivSlot,
		RepositoryPath:  repositoryPath,
	}

	path := filepath.Join(homeDir,DefaultFileName)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(conf); err != nil {
		return fmt.Errorf("could not encode configuration: %w", err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
