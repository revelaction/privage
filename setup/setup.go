package setup

import (
	"fmt"
	"os"
	"strconv"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/fs"

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
type Options struct {
	KeyFile    string
	ConfigFile string
	RepoPath   string
	PivSlot    string
}

// Copy returns a copy of the Setup s with an empty Identity
func (s *Setup) Copy() *Setup {

	conf := s.C
	repo := s.Repository
	return &Setup{C: conf, Repository: repo}
}

func NewFromArgs(keyPath, repoPath, pivSlot string) (*Setup, error) {

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
		return id.Load(keyPath)
	}

	slot, err := strconv.ParseUint(pivSlot, 16, 32)
	if err != nil {
		return id.Identity{Err: fmt.Errorf("could not convert slot %d to hex: %v", slot, err)}
	}

	return id.LoadPiv(keyPath, uint32(slot), "")
}

