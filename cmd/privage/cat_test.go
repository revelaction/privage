package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestCatCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupData      func(th *TestHelper)
		label          string
		expectedOutput string
		expectedErr    error
	}{
		{
			name: "Success",
			setupData: func(th *TestHelper) {
				th.AddEncryptedFile("secret.txt", "", "real secret content")
			},
			label:          "secret.txt",
			expectedOutput: "real secret content",
		},
		{
			name: "Label Not Found",
			setupData: func(th *TestHelper) {
				// No files encrypted
			},
			label:       "missing.txt",
			expectedErr: ErrFileNotFound,
		},
		{
			name: "Identity Error",
			setupData: func(th *TestHelper) {
				th.Id.Id = nil
				th.Id.Err = errors.New("key failed")
			},
			label:       "any.txt",
			expectedErr: ErrNoIdentity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelper(t)
			tt.setupData(th)

			var outBuf, errBuf bytes.Buffer
			ui := UI{Out: &outBuf, Err: &errBuf}

			// th.Setup is embedded
			err := catCommand(th.Setup, tt.label, ui)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if outBuf.String() != tt.expectedOutput {
					t.Errorf("expected output %q, got %q", tt.expectedOutput, outBuf.String())
				}
			}
		})
	}
}
