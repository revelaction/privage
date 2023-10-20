package setup

import (
	"fmt"
	"os"
	"strconv"

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

// Copy returns a copy of the Setup s with an empty Identity
func (s *Setup) Copy() *Setup {

	var conf *config.Config = s.C
	repo := s.Repository
	return &Setup{C: conf, Repository: repo}
}

func NewFromArgs(keyPath, repoPath, pivSlot string) (*Setup, error) {

	id := identity(keyPath, pivSlot)

	_, err := directoryExists(repoPath)
	if err != nil {
		return &Setup{}, err
	}

	return &Setup{C: &config.Config{}, Id: id, Repository: repoPath}, nil
}

func NewFromConfigFile(path string) (*Setup, error) {

	// validates the path, toml, and identiti, repo paths
	conf, err := config.New(path)
	if err != nil {
		return &Setup{}, err
	}

	id := identity(conf.IdentityPath, conf.IdentityPivSlot)

	return &Setup{C: conf, Id: id, Repository: conf.RepositoryPath}, nil
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

func directoryExists(path string) (bool, error) {

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return stat.IsDir(), nil
}
