package fs

import (
	"errors"
	"os"
	"path/filepath"
)

// FileExists checks if a file exists and is a regular file.
func FileExists(path string) (bool, error) {
	fi, err := os.Stat(path)
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
	fi, err := os.Stat(path)
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
// Returns the path if found, or an error if not found.
func FindIdentityFile() (string, error) {
	// Try current directory first
	currentDir, err := os.Getwd()
	if err == nil {
		currentPath := filepath.Join(currentDir, "privage-key.txt")
		if exists, _ := FileExists(currentPath); exists {
			return currentPath, nil
		}
	}

	// Try home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homePath := filepath.Join(homeDir, "privage-key.txt")
		if exists, _ := FileExists(homePath); exists {
			return homePath, nil
		}
	}

	return "", errors.New("identity file not found in current or home directory")
}