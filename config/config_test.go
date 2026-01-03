package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		toml    string
		wantErr bool
	}{
		{
			name: "Valid config",
			toml: `
identity_path = "/path/to/key"
repository_path = "/path/to/repo"
login = "user"
`,
			wantErr: false,
		},
		{
			name:    "Invalid TOML",
			toml:    `identity_path = `,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.toml)
			conf, err := decode(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && conf == nil {
				t.Error("decode() returned nil config without error")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists")
	_ = os.WriteFile(existingFile, []byte("data"), 0600)

	tests := []struct {
		name    string
		conf    *Config
		wantErr bool
	}{
		{
			name: "Valid paths",
			conf: &Config{
				IdentityPath:   existingFile,
				RepositoryPath: tmpDir,
			},
			wantErr: false,
		},
		{
			name: "Missing identity_path",
			conf: &Config{
				RepositoryPath: tmpDir,
			},
			wantErr: true,
		},
		{
			name: "Non-existent identity_path",
			conf: &Config{
				IdentityPath:   "/non/existent",
				RepositoryPath: tmpDir,
			},
			wantErr: true,
		},
		{
			name: "Missing repository_path",
			conf: &Config{
				IdentityPath: existingFile,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	conf := &Config{
		IdentityPath:   "~/key.txt",
		RepositoryPath: "~/repo",
	}

	err := conf.expandHome()
	if err != nil {
		t.Fatalf("expandHome() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(conf.IdentityPath, home) {
		t.Errorf("expected IdentityPath to be expanded, got %s", conf.IdentityPath)
	}
	if !strings.HasPrefix(conf.RepositoryPath, home) {
		t.Errorf("expected RepositoryPath to be expanded, got %s", conf.RepositoryPath)
	}
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	idPath := filepath.Join(tmpDir, "id")
	_ = os.WriteFile(idPath, []byte("id"), 0600)

	content := fmt.Sprintf(`
identity_path = '%s'
repository_path = '%s'
`, idPath, tmpDir)

	conf, err := decode(strings.NewReader(content))
	if err != nil {
		t.Fatalf("decode() error = %v", err)
	}

	if err := conf.expandHome(); err != nil {
		t.Fatalf("expandHome() error = %v", err)
	}

	if err := conf.validate(); err != nil {
		t.Fatalf("validate() error = %v", err)
	}

	if conf.IdentityPath != idPath {
		t.Errorf("expected IdentityPath %s, got %s", idPath, conf.IdentityPath)
	}
}
func TestEncode(t *testing.T) {
	conf := &Config{
		IdentityPath:   "/path/to/id",
		IdentityType:   "AGE",
		RepositoryPath: "/path/to/repo",
		Login:          "user",
	}

	var buf strings.Builder
	err := conf.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	output := buf.String()
	// Basic checks for expected TOML content
	expected := []string{
		`identity_path = '/path/to/id'`,
		`identity_type = 'AGE'`,
		`repository_path = '/path/to/repo'`,
		`login = 'user'`,
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Encode() output missing expected string: %s\nOutput:\n%s", exp, output)
		}
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	idPath := filepath.Join(tmpDir, "id")
	_ = os.WriteFile(idPath, []byte("id"), 0600)

	tests := []struct {
		name    string
		toml    string
		wantErr bool
	}{
		{
			name: "Valid config",
			toml: fmt.Sprintf(`
identity_path = '%s'
repository_path = '%s'
login = 'user'
`, idPath, tmpDir),
			wantErr: false,
		},
		{
			name: "Valid config with ~/ paths",
			toml: `
identity_path = '~/test_id'
repository_path = '~/test_repo'
`,
			wantErr: true, // Will fail because ~/ paths don't exist
		},
		{
			name:    "Invalid TOML",
			toml:    `identity_path = `,
			wantErr: true,
		},
		{
			name: "Missing required field",
			toml: fmt.Sprintf(`
identity_path = '%s'
# missing repository_path
`, idPath),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.toml)
			conf, err := Load(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && conf == nil {
				t.Error("Load() returned nil config without error")
			}
		})
	}
}
