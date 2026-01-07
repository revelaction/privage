package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestBashCommand(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &errBuf}

	err := bashCommand(ui)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "_privage_autocomplete") {
		t.Errorf("expected output to contain '_privage_autocomplete', got %q", output)
	}
	if !strings.Contains(output, "complete -F _privage_autocomplete privage") {
		t.Errorf("expected output to contain 'complete -F _privage_autocomplete privage', got %q", output)
	}
}
