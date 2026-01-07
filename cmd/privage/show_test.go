package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestShow_FullContent(t *testing.T) {
	th := NewTestHelper(t)
	content := "login = \"user123\"\npassword = \"supersecret\"\n"
	th.AddEncryptedFile("mycred", "credential", content)

	var outBuf bytes.Buffer
	ui := UI{Out: &outBuf, Err: &bytes.Buffer{}}

	// Act
	err := showCommand(th.Setup, "mycred", "", ui)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := outBuf.String()
	if !strings.Contains(output, "user123") {
		t.Errorf("expected output to contain 'user123', got:\n%s", output)
	}
}

func TestShow_SpecificField(t *testing.T) {
	th := NewTestHelper(t)
	content := "login = \"user123\"\npassword = \"supersecret\"\n"
	th.AddEncryptedFile("mycred", "credential", content)

	tests := []struct {
		field    string
		expected string
	}{
		{"login", "user123"},
		{"password", "supersecret"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			var outBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &bytes.Buffer{}}

			err := showCommand(th.Setup, "mycred", tt.field, ui)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if outBuf.String() != tt.expected {
				t.Errorf("field %s: expected %q, got %q", tt.field, tt.expected, outBuf.String())
			}
		})
	}
}

func TestShow_Errors(t *testing.T) {
	tests := []struct {
		name      string
		setupData func(th *TestHelper)
		label     string
		field     string
		wantErr   error
	}{
		{
			name:      "File Not Found",
			setupData: func(th *TestHelper) {},
			label:     "missing",
			wantErr:   ErrFileNotFound,
		},
		{
			name: "Field Not Found",
			setupData: func(th *TestHelper) {
				th.AddEncryptedFile("mycred", "credential", "login=\"u\"")
			},
			label:   "mycred",
			field:   "bad_field",
			wantErr: ErrFieldNotFound,
		},
		{
			name: "Wrong Category",
			setupData: func(th *TestHelper) {
				th.AddEncryptedFile("notes", "work", "stuff")
			},
			label:   "notes",
			wantErr: ErrNotCredential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelper(t)
			tt.setupData(th)
			ui := UI{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

			err := showCommand(th.Setup, tt.label, tt.field, ui)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
