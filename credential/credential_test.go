package credential

import (
	"bytes"
	"strings"
	"testing"

	"github.com/revelaction/privage/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		wantLogin string
	}{
		{
			name:      "WithEmailOnly",
			cfg:       &config.Config{Email: "user@example.com"},
			wantLogin: "user@example.com",
		},
		{
			name:      "WithLoginAndEmail",
			cfg:       &config.Config{Login: "myuser", Email: "user@example.com"},
			wantLogin: "myuser",
		},
		{
			name:      "EmptyConfig",
			cfg:       &config.Config{},
			wantLogin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := New(tt.cfg)
			if err != nil {
				t.Fatalf("New failed: %v", err)
			}
			if cred.Login != tt.wantLogin {
				t.Errorf("expected login %s, got %s", tt.wantLogin, cred.Login)
			}
			if len(cred.Password) != 25 {
				t.Errorf("expected password length 25, got %d", len(cred.Password))
			}
			if !strings.Contains(cred.Remarks, "Put here all the rest") {
				t.Error("expected default remarks to be present")
			}
		})
	}
}

func TestCredential_RoundTrip(t *testing.T) {
	// Setup a credential with fixed and custom fields
	original := &Credential{
		Login:    "lipo",
		Password: "password123",
		Email:    "lipo@example.com",
		Others: map[string]any{
			"pin":           int64(1234), // toml decodes numbers as int64
			"security_hint": "cat name",
		},
	}

	// 1. Encode
	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// 2. Decode
	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// 3. Assert fixed fields
	if decoded.Login != original.Login {
		t.Errorf("expected login %s, got %s", original.Login, decoded.Login)
	}
	if decoded.Password != original.Password {
		t.Errorf("expected password %s, got %s", original.Password, decoded.Password)
	}

	// 4. Assert custom fields (Others)
	if decoded.Others["pin"] != original.Others["pin"] {
		t.Errorf("expected pin %v, got %v", original.Others["pin"], decoded.Others["pin"])
	}
	if decoded.Others["security_hint"] != original.Others["security_hint"] {
		t.Errorf("expected hint %v, got %v", original.Others["security_hint"], decoded.Others["security_hint"])
	}
}

func TestCredential_Fprint(t *testing.T) {
	cred := &Credential{
		Login:    "user123",
		Password: "secret-password",
		Others: map[string]any{
			"custom_field": "custom_value",
		},
	}

	var buf bytes.Buffer
	if err := cred.Fprint(&buf); err != nil {
		t.Fatalf("Fprint failed: %v", err)
	}

	output := buf.String()

	// Check primary fields
	if !strings.Contains(output, "👤 user123") {
		t.Error("output missing login with icon")
	}
	if !strings.Contains(output, "🔑 secret-password") {
		t.Error("output missing password with icon")
	}

	// Check custom fields
	if !strings.Contains(output, "custom_field: custom_value") {
		t.Error("output missing custom field from Others map")
	}
}

func TestCredential_Formatting(t *testing.T) {
	cred := &Credential{
		Remarks: "line 1\nline 2",
	}

	var buf bytes.Buffer
	if err := cred.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	output := buf.String()

	// Verify multiline remarks formatting (should use triple quotes)
	if !strings.Contains(output, `"""`) {
		t.Errorf("expected multiline triple quotes for remarks, got:\n%s", output)
	}

	// Verify comments are present (as defined in struct tags)
	if !strings.Contains(output, "# Associated email address") {
		t.Error("expected comments from struct tags in TOML output")
	}
}
