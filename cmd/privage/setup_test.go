package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/revelaction/privage/config"
	"github.com/revelaction/privage/identity"
	"github.com/revelaction/privage/setup"
)

func TestSetupEnv_ExplicitFlags(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// 1. Create identity file
	idPath := filepath.Join(tmpDir, "my-key.txt")
	f, err := os.Create(idPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := identity.GenerateAge(f); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// 2. Create repository directory
	repoPath := filepath.Join(tmpDir, "my-repo")
	if err := os.Mkdir(repoPath, 0755); err != nil {
		t.Fatal(err)
	}

	opts := setup.Options{
		KeyFile:  idPath,
		RepoPath: repoPath,
	}

	s, err := setupEnv(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Repository != repoPath {
		t.Errorf("got repository %q, want %q", s.Repository, repoPath)
	}
	if s.Id.Path != idPath {
		t.Errorf("got identity path %q, want %q", s.Id.Path, idPath)
	}
	if s.Id.Id == nil {
		t.Error("expected loaded identity, got nil")
	}
}

func TestSetupEnv_ConfigFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// 1. Create identity file
	idPath := filepath.Join(tmpDir, "key.txt")
	f, err := os.Create(idPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = identity.GenerateAge(f)
	f.Close()

	// 2. Create repository
	repoPath := filepath.Join(tmpDir, "repo")
	_ = os.Mkdir(repoPath, 0755)

	// 3. Create config file
	conf := &config.Config{
		IdentityPath:   idPath,
		RepositoryPath: repoPath,
	}
	confPath := filepath.Join(tmpDir, "test.conf")
	cf, _ := os.Create(confPath)
	_ = conf.Encode(cf)
	cf.Close()

	opts := setup.Options{
		ConfigFile: confPath,
	}

	s, err := setupEnv(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Repository != repoPath {
		t.Errorf("got repository %q, want %q", s.Repository, repoPath)
	}
	if s.Id.Path != idPath {
		t.Errorf("got identity path %q, want %q", s.Id.Path, idPath)
	}
}

func TestSetupEnv_DiscoveryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create identity and repo
	idPath := filepath.Join(tmpDir, "key.txt")
	f, _ := os.Create(idPath)
	_ = identity.GenerateAge(f)
	f.Close()

	repoPath := filepath.Join(tmpDir, "repo")
	_ = os.Mkdir(repoPath, 0755)

	// Create config in HOME
	conf := &config.Config{
		IdentityPath:   idPath,
		RepositoryPath: repoPath,
	}
	confPath := filepath.Join(tmpDir, config.DefaultFileName)
	cf, _ := os.Create(confPath)
	_ = conf.Encode(cf)
	cf.Close()

	opts := setup.Options{} // No flags

	s, err := setupEnv(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Repository != repoPath {
		t.Errorf("got repo %q, want %q", s.Repository, repoPath)
	}
}

func TestSetupEnv_DiscoveryIdentity(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Mock current directory for repository discovery
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create identity in HOME/privage-key.txt (FindIdentityFile searches here)
	idPath := filepath.Join(tmpDir, identity.DefaultFileName)
	f, _ := os.Create(idPath)
	_ = identity.GenerateAge(f)
	f.Close()

	opts := setup.Options{}

	s, err := setupEnv(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Id.Path != idPath {
		t.Errorf("got id path %q, want %q", s.Id.Path, idPath)
	}
	// Repository should be current directory (tmpDir)
	if s.Repository != tmpDir {
		t.Errorf("got repo %q, want %q", s.Repository, tmpDir)
	}
}

func TestSetupEnv_PriorityFlagsOverConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// 1. Create a config file that points to WRONG paths
	badRepo := filepath.Join(tmpDir, "bad-repo")
	_ = os.Mkdir(badRepo, 0755)
	conf := &config.Config{
		IdentityPath:   filepath.Join(tmpDir, "bad-key.txt"),
		RepositoryPath: badRepo,
	}
	confPath := filepath.Join(tmpDir, config.DefaultFileName)
	cf, _ := os.Create(confPath)
	_ = conf.Encode(cf)
	cf.Close()

	// 2. Create CORRECT paths for flags
	goodRepo := filepath.Join(tmpDir, "good-repo")
	_ = os.Mkdir(goodRepo, 0755)
	goodKey := filepath.Join(tmpDir, "good-key.txt")
	f, _ := os.Create(goodKey)
	_ = identity.GenerateAge(f)
	f.Close()

	// 3. Provide explicit flags. They should overrule the discovery of the config file.
	// Note: currently setupEnv prioritize WithKeyRepo over NoKeyRepoConfig.
	opts := setup.Options{
		KeyFile:  goodKey,
		RepoPath: goodRepo,
	}

	s, err := setupEnv(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Repository != goodRepo {
		t.Errorf("flags should have priority over config file. got %q, want %q", s.Repository, goodRepo)
	}
}

func TestSetupEnv_Validation(t *testing.T) {
	opts := setup.Options{
		ConfigFile: "a",
		KeyFile:    "b", // Incompatible
	}

	_, err := setupEnv(opts)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSetupEnv_PIV_ParsingOnly(t *testing.T) {
	// We only verify the behavior when a slot is provided.
	// If a Yubikey happens to be present, this test might actually succeed or fail differently.
	
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	
	idPath := filepath.Join(tmpDir, "key.txt")
	f, _ := os.Create(idPath)
	_ = identity.GenerateAge(f)
	f.Close()

	opts := setup.Options{
		KeyFile:  idPath,
		RepoPath: tmpDir,
		PivSlot:  "9c",
	}

	_, err := setupEnv(opts)
	if err == nil {
		// It might succeed if a Yubikey is actually plugged in and slot 9c is valid.
		// But in most CI/dev envs it will fail.
		return 
	}
	
	// If it fails, it should be due to Yubikey missing or disabled build tag
	errMsg := err.Error()
	if !strings.Contains(errMsg, "card") && 
	   !strings.Contains(errMsg, "yubikey") && 
	   !strings.Contains(errMsg, "device") {
		t.Errorf("expected yubikey related error, got: %v", err)
	}
}