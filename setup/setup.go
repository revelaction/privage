package setup

import (
	"errors"

	"github.com/revelaction/privage/config"
	id "github.com/revelaction/privage/identity"
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
//   - KeyFile and RepoPath must both be set
//   - ConfigFile must be empty
//   - PivSlot is optional
//
// 2. WithConfig(): -c flag only
//   - ConfigFile must be set
//   - KeyFile and RepoPath must be empty
//
// 3. NoFlags(): no flags specified
//   - All fields empty
//   - Will search for config file in standard locations
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

// NoKeyRepoConfig returns true if no explicit setup options (config, key, repo) are specified.
// This allows for auto-discovery, optionally with a PIV slot.
func (o *Options) NoKeyRepoConfig() bool {
	return o.ConfigFile == "" && o.KeyFile == "" && o.RepoPath == ""
}

// Copy returns a copy of the Setup s with an empty Identity
func (s *Setup) Copy() *Setup {

	conf := s.C
	repo := s.Repository
	return &Setup{C: conf, Repository: repo}
}