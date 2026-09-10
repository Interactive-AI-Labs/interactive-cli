package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

func TestNewSession(t *testing.T) {
	t.Run("creates session with config dir", func(t *testing.T) {
		cfgDir := ".test-config"
		s := NewSession(cfgDir)
		if s == nil {
			t.Fatal("NewSession() returned nil")
		}
		if s.cfgDirName != cfgDir {
			t.Errorf("NewSession() cfgDirName = %q, want %q", s.cfgDirName, cfgDir)
		}
	})

	t.Run("creates session with empty config dir", func(t *testing.T) {
		s := NewSession("")
		if s == nil {
			t.Fatal("NewSession() returned nil")
		}
		if s.cfgDirName != "" {
			t.Errorf("NewSession() cfgDirName = %q, want empty", s.cfgDirName)
		}
	})
}

func TestResolveOrganization(t *testing.T) {
	tests := []struct {
		name        string
		cfgOrg      string
		flagOrg     string
		envOrg      string
		selectedOrg string
		want        string
		wantErr     bool
	}{
		{
			name:        "flag org takes precedence over cfg, env and selected",
			cfgOrg:      "cfg-org",
			flagOrg:     "flag-org",
			envOrg:      "env-org",
			selectedOrg: "selected-org",
			want:        "flag-org",
			wantErr:     false,
		},
		{
			name:        "cfg org takes precedence over env and selected",
			cfgOrg:      "cfg-org",
			flagOrg:     "",
			envOrg:      "env-org",
			selectedOrg: "selected-org",
			want:        "cfg-org",
			wantErr:     false,
		},
		{
			name:        "env org takes precedence over selected",
			envOrg:      "env-org",
			selectedOrg: "selected-org",
			want:        "env-org",
		},
		{
			name:   "env org works without a saved selection and is trimmed",
			envOrg: "  env-org\t",
			want:   "env-org",
		},
		{
			name:        "whitespace env org falls back to selected",
			envOrg:      " \t\n",
			selectedOrg: "selected-org",
			want:        "selected-org",
		},
		{
			name:        "selected org when no flag, cfg or env",
			cfgOrg:      "",
			flagOrg:     "",
			selectedOrg: "selected-org",
			want:        "selected-org",
			wantErr:     false,
		},
		{
			name:        "error when all empty",
			cfgOrg:      "",
			flagOrg:     "",
			selectedOrg: "",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "trims whitespace from flag org",
			cfgOrg:      "",
			flagOrg:     "  flag-org  ",
			selectedOrg: "",
			want:        "flag-org",
			wantErr:     false,
		},
		{
			name:        "trims whitespace from cfg org",
			cfgOrg:      "  cfg-org  ",
			flagOrg:     "",
			selectedOrg: "",
			want:        "cfg-org",
			wantErr:     false,
		},
		{
			name:        "whitespace-only values are treated as empty",
			cfgOrg:      "   ",
			flagOrg:     "  ",
			envOrg:      "  ",
			selectedOrg: "",
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("INTERACTIVE_ORGANIZATION", tt.envOrg)
			cfgDir := ".test-session"

			if tt.selectedOrg != "" {
				if err := files.SelectOrg(cfgDir, tt.selectedOrg); err != nil {
					t.Fatalf("SelectOrg() error = %v", err)
				}
			}

			s := NewSession(cfgDir)
			got, err := s.ResolveOrganization(tt.cfgOrg, tt.flagOrg)

			if tt.selectedOrg == "" {
				_, statErr := os.Stat(filepath.Join(os.Getenv("HOME"), cfgDir, "config.yaml"))
				if !os.IsNotExist(statErr) {
					t.Fatal("resolving organization must not create a saved selection")
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveOrganization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveOrganization() = %q, want %q", got, tt.want)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "organization is required") {
					t.Errorf(
						"error message should mention 'organization is required', got: %v",
						err,
					)
				}
			}
		})
	}
}

func TestResolveProject(t *testing.T) {
	tests := []struct {
		name            string
		cfgProject      string
		flagProject     string
		envProject      string
		selectedProject string
		want            string
		wantErr         bool
	}{
		{
			name:            "flag project takes precedence over cfg, env and selected",
			cfgProject:      "cfg-proj",
			flagProject:     "flag-proj",
			envProject:      "env-proj",
			selectedProject: "selected-proj",
			want:            "flag-proj",
			wantErr:         false,
		},
		{
			name:            "cfg project takes precedence over env and selected",
			cfgProject:      "cfg-proj",
			flagProject:     "",
			envProject:      "env-proj",
			selectedProject: "selected-proj",
			want:            "cfg-proj",
			wantErr:         false,
		},
		{
			name:            "env project takes precedence over selected",
			envProject:      "env-proj",
			selectedProject: "selected-proj",
			want:            "env-proj",
		},
		{
			name:       "env project works without a saved selection and is trimmed",
			envProject: "  env-proj\t",
			want:       "env-proj",
		},
		{
			name:            "whitespace env project falls back to selected",
			envProject:      " \t\n",
			selectedProject: "selected-proj",
			want:            "selected-proj",
		},
		{
			name:            "selected project when no flag, cfg or env",
			cfgProject:      "",
			flagProject:     "",
			selectedProject: "selected-proj",
			want:            "selected-proj",
			wantErr:         false,
		},
		{
			name:            "error when all empty",
			cfgProject:      "",
			flagProject:     "",
			selectedProject: "",
			want:            "",
			wantErr:         true,
		},
		{
			name:            "trims whitespace from flag project",
			cfgProject:      "",
			flagProject:     "  flag-proj  ",
			selectedProject: "",
			want:            "flag-proj",
			wantErr:         false,
		},
		{
			name:            "trims whitespace from cfg project",
			cfgProject:      "  cfg-proj  ",
			flagProject:     "",
			selectedProject: "",
			want:            "cfg-proj",
			wantErr:         false,
		},
		{
			name:            "whitespace-only values are treated as empty",
			cfgProject:      "   ",
			flagProject:     "  ",
			envProject:      "  ",
			selectedProject: "",
			want:            "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("INTERACTIVE_PROJECT", tt.envProject)
			cfgDir := ".test-session"

			if tt.selectedProject != "" {
				if err := files.SelectProject(cfgDir, tt.selectedProject); err != nil {
					t.Fatalf("SelectProject() error = %v", err)
				}
			}

			s := NewSession(cfgDir)
			got, err := s.ResolveProject(tt.cfgProject, tt.flagProject)

			if tt.selectedProject == "" {
				_, statErr := os.Stat(filepath.Join(os.Getenv("HOME"), cfgDir, "config.yaml"))
				if !os.IsNotExist(statErr) {
					t.Fatal("resolving project must not create a saved selection")
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveProject() = %q, want %q", got, tt.want)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "project is required") {
					t.Errorf("error message should mention 'project is required', got: %v", err)
				}
			}
		})
	}
}
