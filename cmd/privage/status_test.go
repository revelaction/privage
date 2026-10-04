package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/revelaction/privage/setup"
)

func TestStatusCommand(t *testing.T) {
	t.Run("Repository", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		s := &setup.Setup{Repository: "/home/user/mysecrets"}

		err := statusCommand(s, true, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := outBuf.String()
		want := "/home/user/mysecrets\n"
		if got != want {
			t.Errorf("got output %q, want %q", got, want)
		}
	})

	t.Run("Default", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		s := &setup.Setup{Repository: "/home/user/mysecrets"}

		err := statusCommand(s, false, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(outBuf.String(), "directory of the encrypted files is /home/user/mysecrets") {
			t.Error("expected the repository directory in the status output")
		}
	})
}
