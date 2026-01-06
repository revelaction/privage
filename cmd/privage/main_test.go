package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/revelaction/privage/setup"
)

func TestRunCommand_Utility(t *testing.T) {
	// Utility commands don't need a real setup environment
	
	t.Run("version", func(t *testing.T) {
		err := runCommand("version", []string{}, setup.Options{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("bash", func(t *testing.T) {
		err := runCommand("bash", []string{}, setup.Options{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("help", func(t *testing.T) {
		err := runCommand("help", []string{}, setup.Options{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("help sub-command", func(t *testing.T) {
		err := runCommand("help", []string{"cat"}, setup.Options{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		err := runCommand("not-a-command", []string{}, setup.Options{})
		if err == nil {
			t.Fatal("expected error for unknown command")
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("expected unknown command error, got: %v", err)
		}
	})
}

func TestSetupUsage(t *testing.T) {
	// Verify that setupUsage correctly configures the global flag.Usage
	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf)
	
	setupUsage()
	flag.Usage()

	output := buf.String()
	expectedContents := []string{
		"Usage:",
		"Commands:",
		"Global Options:",
		"Version:",
	}

	for _, s := range expectedContents {
		if !strings.Contains(output, s) {
			t.Errorf("expected usage to contain %q, but it didn't", s)
		}
	}
}
