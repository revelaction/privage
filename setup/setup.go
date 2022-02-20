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

	var conf *config.Config
	conf = s.C
	repo := s.Repository
	return &Setup{C: conf, Repository: repo}
}

func New(conf *config.Config) (*Setup, error) {

	id := identity(conf)
	repo, err := repository(conf)
	if err != nil {
		return nil, err
	}

	return &Setup{C: conf, Id: id, Repository: repo}, nil
}

func identity(c *config.Config) id.Identity {
	path := ""
	if len(c.IdentityPath) > 0 {
		path = c.IdentityPath
	}

	// Only if piv_slot exist in the config and only then, we check the yubikey
	if c.IdentityType == id.TypePiv {
		slot, err := strconv.ParseUint(c.IdentityPivSlot, 16, 32)
		if err != nil {
			return id.Identity{Err: fmt.Errorf("could not convert slot %s to hex: %v", slot, err)}
		}
		return id.LoadPiv(c.IdentityPath, uint32(slot), c.IdentityPivAlgo)
	}

	// Try load as a normal age key
	return id.Load(path)
}

func repository(conf *config.Config) (string, error) {
	// 1) Config
	if len(conf.RepositoryPath) > 0 {
		// TODO we should check dir exist.
		return conf.RepositoryPath, nil
	}

    // 2) current dir
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("No repository directory found: %w", err)
	}

	return path, nil
}
