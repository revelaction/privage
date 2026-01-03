package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists")
	_ = os.WriteFile(existingFile, []byte("data"), 0600)
	existingDir := filepath.Join(tmpDir, "dir")
	_ = os.Mkdir(existingDir, 0755)

	tests := []struct {
		name     string
		path     string
		want     bool
		wantErr  bool
	}{
		{
			name:    "Existing file",
			path:    existingFile,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Existing directory",
			path:    existingDir,
			want:    false, // DirExists should return false for directories
			wantErr: false,
		},
		{
			name:    "Non-existent path",
			path:    filepath.Join(tmpDir, "nonexistent"),
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileExists(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists")
	_ = os.WriteFile(existingFile, []byte("data"), 0600)
	existingDir := filepath.Join(tmpDir, "dir")
	_ = os.Mkdir(existingDir, 0755)

	tests := []struct {
		name     string
		path     string
		want     bool
		wantErr  bool
	}{
		{
			name:    "Existing directory",
			path:    existingDir,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Existing file",
			path:    existingFile,
			want:    false, // DirExists should return false for files
			wantErr: false,
		},
		{
			name:    "Non-existent path",
			path:    filepath.Join(tmpDir, "nonexistent"),
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DirExists(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("DirExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindIdentityFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Change to temp directory
	os.Chdir(tmpDir)

	// Create identity file in current directory
	idFile := filepath.Join(tmpDir, "privage-key.txt")
	_ = os.WriteFile(idFile, []byte("test key"), 0600)

	// Should find in current directory
	path, err := FindIdentityFile()
	if err != nil {
		t.Errorf("FindIdentityFile() error = %v, expected no error", err)
	}
	if path != idFile {
		t.Errorf("FindIdentityFile() = %v, want %v", path, idFile)
	}

	// Remove from current directory, create in home
	os.Remove(idFile)
	homeDir, _ := os.UserHomeDir()
	homeIdFile := filepath.Join(homeDir, "privage-key.txt")
	_ = os.WriteFile(homeIdFile, []byte("test key"), 0600)
	defer os.Remove(homeIdFile)

	// Should find in home directory
	path, err = FindIdentityFile()
	if err != nil {
		t.Errorf("FindIdentityFile() error = %v, expected no error", err)
	}
	if path != homeIdFile {
		t.Errorf("FindIdentityFile() = %v, want %v", path, homeIdFile)
	}

	// Remove both, should get error
	os.Remove(homeIdFile)
	path, err = FindIdentityFile()
	if err == nil {
		t.Error("FindIdentityFile() expected error when no file exists")
	}
}

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Change to temp directory
	os.Chdir(tmpDir)

	// Create config file in home directory first (home has priority)
	homeDir, _ := os.UserHomeDir()
	homeConfigFile := filepath.Join(homeDir, ".privage.conf")
	_ = os.WriteFile(homeConfigFile, []byte("test config"), 0600)
	defer os.Remove(homeConfigFile)

	// Should find in home directory (priority)
	path, err := FindConfigFile()
	if err != nil {
		t.Errorf("FindConfigFile() error = %v, expected no error", err)
	}
	if path != homeConfigFile {
		t.Errorf("FindConfigFile() = %v, want %v", path, homeConfigFile)
	}

	// Remove from home, create in current directory
	os.Remove(homeConfigFile)
	configFile := filepath.Join(tmpDir, ".privage.conf")
	_ = os.WriteFile(configFile, []byte("test config"), 0600)

	// Should find in current directory
	path, err = FindConfigFile()
	if err != nil {
		t.Errorf("FindConfigFile() error = %v, expected no error", err)
	}
	if path != configFile {
		t.Errorf("FindConfigFile() = %v, want %v", path, configFile)
	}

	// Remove both, should get error
	os.Remove(configFile)
	path, err = FindConfigFile()
	if err == nil {
		t.Error("FindConfigFile() expected error when no file exists")
	}
}