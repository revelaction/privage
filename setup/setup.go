package setup

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/fs"

	id "github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/identity/piv/yubikey"
)

// Setup contains the path of the secrets repository and the path of
// the age identity.
//
// This allow to run the privage command in any directory
type Setup struct {

	// Config object, can be empty struct
	C *config.Config

	// Identity wrapper, can be a empty age identity, in that case the error is
	// present
	Id id.Identity

	// The real repo path, if no from config toml, this will be current path
	Repository string
}

// Options contains the configuration parameters provided via global flags
// or environment to initialize the Setup.
//
// Priority of options (checked in order):
// 1. WithKeyRepo(): -k and -r flags (with optional -p)
//    - KeyFile and RepoPath must both be set
//    - ConfigFile must be empty
//    - PivSlot is optional
//
// 2. WithConfig(): -c flag only
//    - ConfigFile must be set
//    - KeyFile and RepoPath must be empty
//
// 3. NoFlags(): no flags specified
//    - All fields empty
//    - Will search for config file in standard locations
//
// Use Validate() to check option validity and the helper methods
// (WithKeyRepo(), WithConfig(), NoFlags()) to determine which case applies.
type Options struct {
	KeyFile    string
	ConfigFile string
	RepoPath   string
	PivSlot    string
}

// Validate checks that the Options are in a valid state.
// See Options documentation for valid states and priority.
func (o *Options) Validate() error {
	// Check incompatible flags: -c and -k cannot both be set
	if o.ConfigFile != "" && o.KeyFile != "" {
		return errors.New("flags -c and -k are incompatible")
	}

	// If -k is set, -r must also be set
	if o.KeyFile != "" && o.RepoPath == "" {
		return errors.New("flag -r is required when using -k")
	}

	// If -c is set, -k and -r must be empty (already checked -k)
	if o.ConfigFile != "" && o.RepoPath != "" {
		return errors.New("flag -r cannot be used with -c")
	}

	// -p can only be used with -k
	if o.PivSlot != "" && o.KeyFile == "" {
		return errors.New("flag -p can only be used with -k")
	}

	return nil
}

// WithConfig returns true if options specify a config file (-c flag).
func (o *Options) WithConfig() bool {
	return o.ConfigFile != ""
}

// WithKeyRepo returns true if options specify explicit key and repo (-k -r flags).
func (o *Options) WithKeyRepo() bool {
	return o.KeyFile != "" && o.RepoPath != ""
}

// NoFlags returns true if no options are specified (no flags).
func (o *Options) NoFlags() bool {
	return o.ConfigFile == "" && o.KeyFile == "" && o.RepoPath == "" && o.PivSlot == ""
}

// Copy returns a copy of the Setup s with an empty Identity
func (s *Setup) Copy() *Setup {

	conf := s.C
	repo := s.Repository
	return &Setup{C: conf, Repository: repo}
}

// NewFromKeyRepoFlags creates a Setup from explicit key and repository paths.
// Used when -k and -r flags are provided.
func NewFromKeyRepoFlags(keyPath, repoPath, pivSlot string) (*Setup, error) {
	// keyPath is never empty when called from WithKeyRepo() case
	// (validated by Options.Validate())
	id := identity(keyPath, pivSlot)

	exists, err := fs.DirExists(repoPath)
	if err != nil {
		return &Setup{}, err
	}
	if !exists {
		return &Setup{}, fmt.Errorf("repository directory %s does not exist", repoPath)
	}

	return &Setup{C: &config.Config{}, Id: id, Repository: repoPath}, nil
}

func NewFromConfigFile(path string) (s *Setup, err error) {

	f, err := os.Open(path)
	if err != nil {
		return &Setup{}, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	conf, err := config.Load(f)
	if err != nil {
		return &Setup{}, fmt.Errorf("invalid configuration file %s: %w", path, err)
	}

	conf.Path = path

	return &Setup{C: conf, Id: identity(conf.IdentityPath, conf.IdentityPivSlot), Repository: conf.RepositoryPath}, nil
}

func identity(keyPath, pivSlot string) id.Identity {

	if pivSlot == "" {
		f, err := fs.OpenFile(keyPath)
		if err != nil {
			return id.Identity{Err: err}
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				// TODO: Handle file close errors properly
				// For read operations, close errors after successful read
				// are less critical but should be logged or monitored
				// Currently we acknowledge but don't propagate the error
				_ = cerr
			}
		}()
		return id.LoadAge(f, keyPath)
	}

	slot, err := strconv.ParseUint(pivSlot, 16, 32)
	if err != nil {
		return id.Identity{Err: fmt.Errorf("could not convert slot %d to hex: %v", slot, err)}
	}

	device, err := yubikey.New()
	if err != nil {
		return id.Identity{Err: fmt.Errorf("could not create yubikey device: %w", err)}
	}
	defer device.Close()

	f, err := fs.OpenFile(keyPath)
	if err != nil {
		return id.Identity{Err: err}
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			// TODO: Handle file close errors properly
			// For read operations, close errors after successful read
			// are less critical but should be logged or monitored
			// Currently we acknowledge but don't propagate the error
			_ = cerr
		}
	}()

	return id.LoadPiv(f, keyPath, device, uint32(slot))
}

