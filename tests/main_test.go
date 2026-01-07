//go:build integration
// This tag ensures these tests are excluded from the default 'go test ./...' run
// and are only executed when '-tags=integration' is explicitly provided.

package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var binDir string

func TestMain(m *testing.M) {
	// 1. Setup: Build binary
	var err error
	binDir, err = os.MkdirTemp("", "privage-test-bin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = os.RemoveAll(binDir)
	}()

	binPath := filepath.Join(binDir, "privage")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	// Build from ../cmd/privage relative to this file
	// We assume "go" is in the path
	cmd := exec.Command("go", "build", "-o", binPath, "../cmd/privage")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build privage: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// 2. Run
	testscript.Main(m, map[string]func(){})
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			// Prepend binDir to PATH so "privage" command is found
			env.Vars = append(env.Vars, fmt.Sprintf("PATH=%s%c%s", binDir, filepath.ListSeparator, os.Getenv("PATH")))
			// Set HOME to the test work directory to allow config file creation
			env.Vars = append(env.Vars, fmt.Sprintf("HOME=%s", env.WorkDir))
			return nil
		},
	})
}
