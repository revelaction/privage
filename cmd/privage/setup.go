package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/fs"
	"github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/identity/piv/yubikey"
	"github.com/revelaction/privage/setup"
)

// setupEnv loads the identity file (secret key) and the repository command
// arguments have preference. Explicite -k, -r arguments have preference over
// the -c argument.  If no arguments provided, standard paths are searched for
// a configuration file .privage.conf
func setupEnv(opts setup.Options) (*setup.Setup, error) {
	// Validate options first
	if err := opts.Validate(); err != nil {
		return &setup.Setup{}, err
	}

	// Determine which case we're in using explicit methods
	switch {
	case opts.WithKeyRepo():
		// Case 1: -k -r with optional -p
		return setupFromKeyRepo(opts.KeyFile, opts.RepoPath, opts.PivSlot)

	case opts.WithConfig():
		// Case 2: -c only
		return setupFromConfig(opts.ConfigFile)

	case opts.NoKeyRepoConfig():
		// Case 3: Nothing - try config file first, then identity file
		return setupFromDiscovery(opts.PivSlot)

	default:
		// This should never happen due to Validate()
		return &setup.Setup{}, fmt.Errorf("invalid option state")
	}
}

func setupFromKeyRepo(keyPath, repoPath, pivSlot string) (*setup.Setup, error) {
	id := loadIdentity(keyPath, pivSlot)

	exists, err := fs.DirExists(repoPath)
	if err != nil {
		return &setup.Setup{}, err
	}
	if !exists {
		return &setup.Setup{}, fmt.Errorf("repository directory %s does not exist", repoPath)
	}

	return &setup.Setup{C: &config.Config{}, Id: id, Repository: repoPath}, nil
}

func setupFromConfig(path string) (*setup.Setup, error) {
	f, err := os.Open(path)
	if err != nil {
		return &setup.Setup{}, err
	}
	defer func() {
		_ = f.Close()
	}()

	conf, err := config.Load(f)
	if err != nil {
		return &setup.Setup{}, fmt.Errorf("invalid configuration file %s: %w", path, err)
	}

	conf.Path = path

	return &setup.Setup{
		C:          conf,
		Id:         loadIdentity(conf.IdentityPath, conf.IdentityPivSlot),
		Repository: conf.RepositoryPath,
	}, nil
}

func setupFromDiscovery(pivSlot string) (*setup.Setup, error) {
	configPath, err := fs.FindConfigFile()
	if err != nil {
		// A real system error occurred (permission denied, etc)
		return &setup.Setup{}, fmt.Errorf("error finding config file: %w", err)
	}

	if configPath != "" {
		// Config file found - use it
		return setupFromConfig(configPath)
	}

	// No config file found (and no error) - search for identity file
	idPath, err := fs.FindIdentityFile()
	if err != nil {
		return &setup.Setup{}, fmt.Errorf("error finding identity file: %w", err)
	}
	if idPath == "" {
		return &setup.Setup{}, fmt.Errorf("no config or identity file found")
	}

	// Use current directory as repository
	repoPath, err := os.Getwd()
	if err != nil {
		return &setup.Setup{}, fmt.Errorf("could not get current directory: %w", err)
	}

	// Create setup with identity file and current directory as repo
	// For auto-discovery we assume standard file-based identity (no PIV slot)
	id := loadIdentity(idPath, pivSlot)
	if id.Err != nil {
		return &setup.Setup{}, id.Err
	}

	return &setup.Setup{
		C:          nil, // No config when using auto-discovery
		Id:         id,
		Repository: repoPath,
	}, nil
}

func loadIdentity(keyPath, pivSlot string) (ident identity.Identity) {

	if pivSlot == "" {
		f, err := fs.OpenFile(keyPath)
		if err != nil {
			return identity.Identity{Err: err}
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && ident.Err == nil {
				ident.Err = cerr
			}
		}()
		return identity.LoadAge(f, keyPath)
	}

	slot, err := strconv.ParseUint(pivSlot, 16, 32)
	if err != nil {
		return identity.Identity{Err: fmt.Errorf("could not convert slot %d to hex: %v", slot, err)}
	}

	device, err := yubikey.New()
	if err != nil {
		return identity.Identity{Err: fmt.Errorf("could not create yubikey device: %w", err)}
	}
	defer func() {
		if cerr := device.Close(); cerr != nil && ident.Err == nil {
			ident.Err = cerr
		}
	}()

	f, err := fs.OpenFile(keyPath)
	if err != nil {
		return identity.Identity{Err: err}
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && ident.Err == nil {
			ident.Err = cerr
		}
	}()

	return identity.LoadPiv(f, keyPath, device, uint32(slot))
}
