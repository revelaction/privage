package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseCatArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantLabel string
		wantErr   bool
		isHelp    bool
	}{
		{"Valid label", []string{"mylabel"}, "mylabel", false, false},
		{"Missing label", []string{}, "", true, false},
		{"Help flag", []string{"-h"}, "", true, true},
		{"Unknown flag", []string{"--foo"}, "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			label, err := parseCatArgs(tt.args, ui)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseCatArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.isHelp && !errors.Is(err, flag.ErrHelp) {
				t.Errorf("expected flag.ErrHelp, got %v", err)
			}
			if label != tt.wantLabel {
				t.Errorf("got label %q, want %q", label, tt.wantLabel)
			}
		})
	}
}

func TestParseInitArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantSlot string
		wantErr  bool
	}{
		{"No args", []string{}, "", false},
		{"Valid slot", []string{"-p", "9c"}, "9c", false},
		{"Help", []string{"--help"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			slot, err := parseInitArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err)
			}
			if slot != tt.wantSlot {
				t.Errorf("got %q, want %q", slot, tt.wantSlot)
			}
		})
	}
}

func TestParseAddArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCat  string
		wantLab  string
		wantErr  bool
	}{
		{"Valid", []string{"cred", "mylabel"}, "cred", "mylabel", false},
		{"Missing one", []string{"cred"}, "", "", true},
		{"Category too long", []string{strings.Repeat("a", 33), "lab"}, "", "", true},
		{"Label too long", []string{"cat", strings.Repeat("a", 129)}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			cat, lab, err := parseAddArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err)
			}
			if cat != tt.wantCat || lab != tt.wantLab {
				t.Errorf("got %q/%q, want %q/%q", cat, lab, tt.wantCat, tt.wantLab)
			}
		})
	}
}

func TestParseShowArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantLabel string
		wantField string
		wantErr   bool
	}{
		{"Label only", []string{"my"}, "my", "", false},
		{"Label and field", []string{"my", "pass"}, "my", "pass", false},
		{"Missing label", []string{}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			label, field, err := parseShowArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err)
			}
			if label != tt.wantLabel || field != tt.wantField {
				t.Errorf("got %q/%q, want %q/%q", label, field, tt.wantLabel, tt.wantField)
			}
		})
	}
}

func TestParseReencryptArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantForce bool
		wantClean bool
		wantErr   bool
	}{
		{"Dry run", []string{}, false, false, false},
		{"Force", []string{"--force"}, true, false, false},
		{"Clean", []string{"-c"}, false, true, false},
		{"Both", []string{"-f", "-c"}, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			force, clean, err := parseReencryptArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err)
			}
			if force != tt.wantForce || clean != tt.wantClean {
				t.Errorf("got %v/%v, want %v/%v", force, clean, tt.wantForce, tt.wantClean)
			}
		})
	}
}

func TestParseRotateArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantClean bool
		wantSlot  string
		wantErr   bool
	}{
		{"Default", []string{}, false, "", false},
		{"Clean and Slot", []string{"--clean", "--piv-slot", "9e"}, true, "9e", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			clean, slot, err := parseRotateArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err)
			}
			if clean != tt.wantClean || slot != tt.wantSlot {
				t.Errorf("got %v/%q, want %v/%q", clean, slot, tt.wantClean, tt.wantSlot)
			}
		})
	}
}

func TestParseListArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantFilter string
		wantErr    bool
	}{
		{"Default", []string{}, "", false},
		{"Filter", []string{"myfilter"}, "myfilter", false},
		{"Help", []string{"--help"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			filter, err := parseListArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseListArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if filter != tt.wantFilter {
				t.Errorf("got filter %q, want %q", filter, tt.wantFilter)
			}
		})
	}
}

func TestParseDeleteArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantLabel string
		wantErr   bool
	}{
		{"Valid label", []string{"mylabel"}, "mylabel", false},
		{"Missing label", []string{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			label, err := parseDeleteArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDeleteArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if label != tt.wantLabel {
				t.Errorf("got label %q, want %q", label, tt.wantLabel)
			}
		})
	}
}

func TestParseKeyArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"No args", []string{}, false},
		{"Help", []string{"--help"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			err := parseKeyArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKeyArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseStatusArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"No args", []string{}, false},
		{"Help", []string{"--help"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			err := parseStatusArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStatusArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseClipboardArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantLabel string
		wantErr   bool
	}{
		{"Valid label", []string{"mylabel"}, "mylabel", false},
		{"Missing label", []string{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			label, err := parseClipboardArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseClipboardArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if label != tt.wantLabel {
				t.Errorf("got label %q, want %q", label, tt.wantLabel)
			}
		})
	}
}

func TestParseDecryptArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantLabel string
		wantErr   bool
	}{
		{"Valid label", []string{"mylabel"}, "mylabel", false},
		{"Missing label", []string{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}
			label, err := parseDecryptArgs(tt.args, ui)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDecryptArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if label != tt.wantLabel {
				t.Errorf("got label %q, want %q", label, tt.wantLabel)
			}
		})
	}
}
