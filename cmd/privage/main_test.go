package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/revelaction/privage/setup"
)

func TestParseMainArgs(t *testing.T) {
	t.Run("Valid flags and command", func(t *testing.T) {
		var out, err bytes.Buffer
		ui := UI{Out: &out, Err: &err}
		cmd, args, opts, parseErr := parseMainArgs([]string{"-c", "my.conf", "list", "filter"}, ui)
		if parseErr != nil {
			t.Fatalf("unexpected error: %v", parseErr)
		}
		if cmd != "list" {
			t.Errorf("got cmd %q, want list", cmd)
		}
		if len(args) != 1 || args[0] != "filter" {
			t.Errorf("got args %v, want [filter]", args)
		}
		if opts.ConfigFile != "my.conf" {
			t.Errorf("got config %q, want my.conf", opts.ConfigFile)
		}
	})

	t.Run("Help", func(t *testing.T) {
		var out, err bytes.Buffer
		ui := UI{Out: &out, Err: &err}
		_, _, _, parseErr := parseMainArgs([]string{"--help"}, ui)
		if !errors.Is(parseErr, flag.ErrHelp) {
			t.Fatalf("expected ErrHelp, got %v", parseErr)
		}
		if out.Len() == 0 {
			t.Error("expected help output in Out")
		}
	})

	t.Run("No command", func(t *testing.T) {
		var out, err bytes.Buffer
		ui := UI{Out: &out, Err: &err}
		_, _, _, parseErr := parseMainArgs([]string{"-k", "key", "-r", "repo"}, ui)
		if parseErr == nil {
			t.Fatal("expected error for missing command")
		}
		if err.Len() == 0 {
			t.Error("expected usage in Err")
		}
	})

	t.Run("Unknown flag", func(t *testing.T) {
		var out, err bytes.Buffer
		ui := UI{Out: &out, Err: &err}
		_, _, _, parseErr := parseMainArgs([]string{"--foo"}, ui)
		if parseErr == nil {
			t.Fatal("expected error for unknown flag")
		}
		if err.Len() == 0 {
			t.Error("expected error message in Err")
		}
	})
}

func TestRunCommand_Utility(t *testing.T) {
	// Utility commands don't need a real setup environment
	var out, err bytes.Buffer
	ui := UI{Out: &out, Err: &err}
	
	t.Run("version", func(t *testing.T) {
		err := runCommand("version", []string{}, setup.Options{}, ui)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("bash", func(t *testing.T) {
		err := runCommand("bash", []string{}, setup.Options{}, ui)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("help", func(t *testing.T) {
		err := runCommand("help", []string{}, setup.Options{}, ui)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("help sub-command", func(t *testing.T) {
		err := runCommand("help", []string{"cat"}, setup.Options{}, ui)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		err := runCommand("not-a-command", []string{}, setup.Options{}, ui)
		if err == nil {
			t.Fatal("expected error for unknown command")
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("expected unknown command error, got: %v", err)
		}
	})
}

func TestSetupUsage(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	
	setupUsage(fs)
	fs.Usage()

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