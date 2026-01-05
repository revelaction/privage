package setup

import (
	"testing"

	"github.com/revelaction/privage/config"
)

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid: Key and Repo",
			options: Options{
				KeyFile:  "key.age",
				RepoPath: ".",
			},
			wantErr: false,
		},
		{
			name: "Valid: Key, Repo, and PivSlot",
			options: Options{
				KeyFile:  "key.age",
				RepoPath: ".",
				PivSlot:  "9a",
			},
			wantErr: false,
		},
		{
			name: "Valid: Config only",
			options: Options{
				ConfigFile: "privage.conf",
			},
			wantErr: false,
		},
		{
			name: "Valid: No options (Auto-discovery)",
			options: Options{
				KeyFile:    "",
				RepoPath:   "",
				ConfigFile: "",
			},
			wantErr: false,
		},
		{
			name: "Valid: No options with PivSlot (Auto-discovery + PIV)",
			options: Options{
				KeyFile:    "",
				RepoPath:   "",
				ConfigFile: "",
				PivSlot:    "9a",
			},
			wantErr: false,
		},
		{
			name: "Invalid: Config and Key",
			options: Options{
				ConfigFile: "privage.conf",
				KeyFile:    "key.age",
			},
			wantErr: true,
			errMsg:  "flags -c and -k are incompatible",
		},
		{
			name: "Invalid: Key without Repo",
			options: Options{
				KeyFile: "key.age",
			},
			wantErr: true,
			errMsg:  "flag -r is required when using -k",
		},
		{
			name: "Invalid: Config and Repo",
			options: Options{
				ConfigFile: "privage.conf",
				RepoPath:   ".",
			},
			wantErr: true,
			errMsg:  "flag -r cannot be used with -c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Options.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Options.Validate() error = %v, wantErrMsg %v", err, tt.errMsg)
			}
		})
	}
}

func TestOptions_StateHelpers(t *testing.T) {
	tests := []struct {
		name              string
		options           Options
		wantWithKeyRepo   bool
		wantWithConfig    bool
		wantNoKeyRepoConf bool
	}{
		{
			name: "State: Key and Repo",
			options: Options{
				KeyFile:  "key.age",
				RepoPath: ".",
			},
			wantWithKeyRepo:   true,
			wantWithConfig:    false,
			wantNoKeyRepoConf: false,
		},
		{
			name: "State: Config only",
			options: Options{
				ConfigFile: "privage.conf",
			},
			wantWithKeyRepo:   false,
			wantWithConfig:    true,
			wantNoKeyRepoConf: false,
		},
		{
			name: "State: Empty (Discovery)",
			options: Options{
				KeyFile:    "",
				RepoPath:   "",
				ConfigFile: "",
			},
			wantWithKeyRepo:   false,
			wantWithConfig:    false,
			wantNoKeyRepoConf: true,
		},
		{
			name: "State: Empty with PivSlot (Discovery)",
			options: Options{
				KeyFile:    "",
				RepoPath:   "",
				ConfigFile: "",
				PivSlot:    "9a",
			},
			wantWithKeyRepo:   false,
			wantWithConfig:    false,
			wantNoKeyRepoConf: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.options.WithKeyRepo(); got != tt.wantWithKeyRepo {
				t.Errorf("WithKeyRepo() = %v, want %v", got, tt.wantWithKeyRepo)
			}
			if got := tt.options.WithConfig(); got != tt.wantWithConfig {
				t.Errorf("WithConfig() = %v, want %v", got, tt.wantWithConfig)
			}
			if got := tt.options.NoKeyRepoConfig(); got != tt.wantNoKeyRepoConf {
				t.Errorf("NoKeyRepoConfig() = %v, want %v", got, tt.wantNoKeyRepoConf)
			}
		})
	}
}

func TestSetup_Copy(t *testing.T) {
	s := &Setup{
		C:          &config.Config{RepositoryPath: "/tmp"},
		Repository: "/tmp",
		// Id is intentionally not deep copied or irrelevant for this test as per implementation
	}

	copied := s.Copy()

	if copied.C != s.C {
		t.Errorf("Copy() Config pointer = %v, want %v", copied.C, s.C)
	}
	if copied.Repository != s.Repository {
		t.Errorf("Copy() Repository = %v, want %v", copied.Repository, s.Repository)
	}
	if copied.Id.Id != nil {
		t.Errorf("Copy() Identity should be empty, got %v", copied.Id)
	}
}
