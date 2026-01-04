package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	// osUserHomeDir allows mocking os.UserHomeDir for testing
	osUserHomeDir = os.UserHomeDir
	// osGetwd allows mocking os.Getwd for testing
	osGetwd = os.Getwd
	// osStat allows mocking os.Stat for testing
	osStat = os.Stat
)

// FileExists checks if a file exists and is a regular file.
func FileExists(path string) (bool, error) {
	fi, err := osStat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !fi.IsDir(), nil
}

// DirExists checks if a directory exists.
func DirExists(path string) (bool, error) {
	fi, err := osStat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return fi.IsDir(), nil
}

// FindIdentityFile searches for the identity file in standard locations.
// It searches in order:
// 1. Current directory: ./privage-key.txt
// 2. Home directory: ~/privage-key.txt
// Returns the path if found, or an empty string if not found.
// Returns an error only if a system error occurs.
func FindIdentityFile() (string, error) {
	// Try current directory first
	currentDir, err := osGetwd()
	if err == nil {
		currentPath := filepath.Join(currentDir, "privage-key.txt")
		exists, err := FileExists(currentPath)
		if err != nil {
			return "", err
		}
		if exists {
			return currentPath, nil
		}
	}

	// Try home directory
	homeDir, err := osUserHomeDir()
	if err == nil {
		homePath := filepath.Join(homeDir, "privage-key.txt")
		exists, err := FileExists(homePath)
		if err != nil {
			return "", err
		}
		if exists {
			return homePath, nil
		}
	}

	return "", nil
}

// FindConfigFile searches for the config file in standard locations.
// It searches in order:
// 1. Home directory: ~/.privage.conf
// 2. Current directory: ./.privage.conf
// Returns the path if found, or an empty string if not found.
// Returns an error only if a system error occurs.
func FindConfigFile() (string, error) {
	// Try home directory first
	homeDir, err := osUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	homePath := filepath.Join(homeDir, ".privage.conf")
	exists, err := FileExists(homePath)
	if err != nil {
		return "", err
	}
	if exists {
		return homePath, nil
	}

	// Try current directory
	currentDir, err := osGetwd()
	if err != nil {
		return "", fmt.Errorf("could not get current directory: %w", err)
	}
	currentPath := filepath.Join(currentDir, ".privage.conf")
	exists, err = FileExists(currentPath)
	if err != nil {
		return "", err
	}
	if exists {
		return currentPath, nil
	}

	return "", nil
}

// OpenFile opens a file for reading.
// Returns an io.ReadCloser that should be closed by the caller.
func OpenFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// CreateFile creates a new file with exclusive access.
// Returns an io.WriteCloser that should be closed by the caller.
func CreateFile(path string, perm os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
}
