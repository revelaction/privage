package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"errors"

	"github.com/BurntSushi/toml"
)

const (
	FileName = ".privage.conf"

	template = `# 
# it supports  "~/" as home directory
identity_path = "%s"
# can be piv for yubikeys
identity_type = "%s" 
# string representation of the Hex value without 0x 0x9a -> 9a 
identity_piv_slot = "%s"

# unimplemented 
#identity_piv_algo = ""

# Repository path
secrets_repository_path = "%s"

# Default login/Username. Leave empty if you want to use email as login
default_login = ""

# Default email. If default_login does not exists, default_email is choosed as credential login
default_email = ""
`
)

// A Config contains configuration data found in the .toml config file (if it exists)
type Config struct {

	// The Path of the existing config file. 
	// If this is not empty, a config file was found
	Path string

	//
	// Identity
	//
	IdentityPath    string `toml:"identity_path"`
	IdentityType    string `toml:"identity_type"`
	IdentityPivSlot string `toml:"identity_piv_slot"`
	IdentityPivAlgo string `toml:"identity_piv_algo"`

	// the repository for all credential and excripted files
	RepositoryPath string `toml:"secrets_repository_path"`

	// Default fields For credentials
	Email string `toml:"default_email"`
	Login string `toml:"default_login"`
}

// New creates a config struct.
//
// It tries to find a valid .toml config file. If not found, an empty
// struct is returned.
func New(argConfig string) (*Config, error) {

   // try conf from arguments
    if fileExists(argConfig) {
        return build(argConfig)
    }

    homePath, err := homeDirPath()
    if err == nil {
        return build(homePath)
    }

    currentPath, err := currentDirPath()
    if err == nil {
        return build(currentPath)
    }

	return &Config{}, err
}

func build(path string) (*Config, error) {
	var conf *Config

	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return &Config{}, fmt.Errorf("file %s is not a valid .toml file.", path)
	}

	// Expand tilde home directory from the config file
	homeDir, err := os.UserHomeDir()
    if err != nil {
        return &Config{}, errors.New("Found no home dir")
    }

	if strings.HasPrefix(conf.IdentityPath, "~/") {
		conf.IdentityPath = filepath.Join(homeDir, conf.IdentityPath[2:])
	}

	if strings.HasPrefix(conf.RepositoryPath, "~/") {
		conf.RepositoryPath = filepath.Join(homeDir, conf.RepositoryPath[2:])
	}

	conf.Path = path
	return conf, nil
}

func homeDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }

    path := homeDir + "/" + FileName
    if !fileExists(path) {
        return "", errors.New("Found no default conf file in home dir")
    }

    return path, nil
}

func currentDirPath() (string, error) {
	currentDir, err := os.Getwd()
    if err != nil {
        return "", err
    }

    path := currentDir+ "/" + FileName
    if !fileExists(path) {
        return "", errors.New("Found no default config file in current dir")
    }

    return path, nil
}

// Create creates a config file.
func Create(identityPath, identityType, identityPivSlot, repositoryPath string) error {

	// 1) try home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	homePath := homeDir + "/" + FileName
	f, err := os.OpenFile(homePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()

	content := fmt.Sprintf(template, identityPath, identityType, identityPivSlot, repositoryPath)

	_, err = f.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err != nil {
			return false
		}
	}

	return true
}
