package main

import "errors"

var (
	// ErrFileNotFound is returned when a requested label does not exist in the directory.
	ErrFileNotFound = errors.New("file not found in directory")

	// ErrFieldNotFound is returned when a requested field (e.g. password) does not exist in the credential.
	ErrFieldNotFound = errors.New("field not found in credential")

	// ErrNotCredential is returned when attempting to show fields of a file that is not a credential.
	ErrNotCredential = errors.New("file is not a credential")

	// ErrNoIdentity is returned when the private key cannot be loaded.
	ErrNoIdentity = errors.New("found no privage key file")
)