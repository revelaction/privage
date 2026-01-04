package fs

import (
	"os"
	"path/filepath"
	"strings"
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

	// Mock osUserHomeDir and osGetwd
	origHomeDir := osUserHomeDir
	origGetwd := osGetwd
	defer func() {
		osUserHomeDir = origHomeDir
		osGetwd = origGetwd
	}()

	osUserHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	osGetwd = func() (string, error) {
		return tmpDir, nil
	}

	// Create identity file in current directory (which is same as home in this mock)
	idFile := filepath.Join(tmpDir, "privage-key.txt")
	requireSafePath(t, idFile, tmpDir)
	if err := os.WriteFile(idFile, []byte("test key"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Should find in current directory
	path, err := FindIdentityFile()
	if err != nil {
		t.Errorf("FindIdentityFile() error = %v, expected no error", err)
	}
	if path != idFile {
		t.Errorf("FindIdentityFile() = %v, want %v", path, idFile)
	}

	// In this simplified mock where CWD == HOME, testing the fallback requires
	// a slightly more complex mock setup if we want to differentiate.
	// However, for basic safety/logic check, ensuring it finds it in the temp dir is key.

	// Remove file to test not found case
	if err := os.Remove(idFile); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	path, err = FindIdentityFile()
	if err != nil {
		t.Errorf("FindIdentityFile() error = %v, expected no error", err)
	}
	if path != "" {
		t.Errorf("FindIdentityFile() = %q, want empty string", path)
	}
}

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	cwdDir := filepath.Join(tmpDir, "cwd")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cwdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Mock osUserHomeDir and osGetwd
	origHomeDir := osUserHomeDir
	origGetwd := osGetwd
	defer func() {
		osUserHomeDir = origHomeDir
		osGetwd = origGetwd
	}()

	osUserHomeDir = func() (string, error) {
		return homeDir, nil
	}
	osGetwd = func() (string, error) {
		return cwdDir, nil
	}

	// Test 1: Config in Home (priority)
	homeConfigFile := filepath.Join(homeDir, ".privage.conf")
	requireSafePath(t, homeConfigFile, tmpDir)
	if err := os.WriteFile(homeConfigFile, []byte("home config"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	path, err := FindConfigFile()
	if err != nil {
		t.Errorf("FindConfigFile() error = %v, expected no error", err)
	}
	if path != homeConfigFile {
		t.Errorf("FindConfigFile() = %v, want %v", path, homeConfigFile)
	}

	// Test 2: Config in CWD (Home config removed)
	if err := os.Remove(homeConfigFile); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	cwdConfigFile := filepath.Join(cwdDir, ".privage.conf")
	requireSafePath(t, cwdConfigFile, tmpDir)
	if err := os.WriteFile(cwdConfigFile, []byte("cwd config"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	path, err = FindConfigFile()
	if err != nil {
		t.Errorf("FindConfigFile() error = %v, expected no error", err)
	}
	if path != cwdConfigFile {
		t.Errorf("FindConfigFile() = %v, want %v", path, cwdConfigFile)
	}

	// Test 3: No config
	if err := os.Remove(cwdConfigFile); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	path, err = FindConfigFile()
	if err != nil {
		t.Errorf("FindConfigFile() error = %v, expected no error", err)
	}
	if path != "" {
		t.Errorf("FindConfigFile() = %q, want empty string", path)
	}
}

// requireSafePath ensures that the given path is inside the allowed base directory
// AND does not already exist. If it violates either, the test fails immediately.
func requireSafePath(t *testing.T, path, baseDir string) {
	t.Helper()
	
	// Safety Check 1: Must be inside the temp directory
	if !strings.HasPrefix(path, baseDir) {
		t.Fatalf("SAFETY CHECK FAILED: Path %s is NOT within allowed temp dir %s. Aborting.", path, baseDir)
	}

	// Safety Check 2: Must not already exist (to prevent overwriting)
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("SAFETY CHECK FAILED: Path %s already exists. Aborting test to prevent data loss.", path)
	} else if !os.IsNotExist(err) {
		// If we can't determine if it exists (e.g. permission error), fail safely
		t.Fatalf("SAFETY CHECK FAILED: Unable to check if path %s exists: %v", path, err)
	}
}