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
		selectedOrg string
		want        string
		wantErr     bool
	}{
		{
			name:        "flag org takes precedence over cfg and selected",
			cfgOrg:      "cfg-org",
			flagOrg:     "flag-org",
			selectedOrg: "selected-org",
			want:        "flag-org",
			wantErr:     false,
		},
		{
			name:        "cfg org takes precedence over selected",
			cfgOrg:      "cfg-org",
			flagOrg:     "",
			selectedOrg: "selected-org",
			want:        "cfg-org",
			wantErr:     false,
		},
		{
			name:        "selected org when no flag or cfg",
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
			selectedOrg: "",
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfgDir := ".test-session-" + t.Name()
			home := os.Getenv("HOME")
			testHome := filepath.Join(tmpDir, "home")
			os.Setenv("HOME", testHome)
			defer os.Setenv("HOME", home)

			if tt.selectedOrg != "" {
				if err := files.SelectOrg(cfgDir, tt.selectedOrg); err != nil {
					t.Fatalf("SelectOrg() error = %v", err)
				}
			}

			s := NewSession(cfgDir)
			got, err := s.ResolveOrganization(tt.cfgOrg, tt.flagOrg)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveOrganization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveOrganization() = %q, want %q", got, tt.want)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "organization is required") {
					t.Errorf("error message should mention 'organization is required', got: %v", err)
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
		selectedProject string
		want            string
		wantErr         bool
	}{
		{
			name:            "flag project takes precedence over cfg and selected",
			cfgProject:      "cfg-proj",
			flagProject:     "flag-proj",
			selectedProject: "selected-proj",
			want:            "flag-proj",
			wantErr:         false,
		},
		{
			name:            "cfg project takes precedence over selected",
			cfgProject:      "cfg-proj",
			flagProject:     "",
			selectedProject: "selected-proj",
			want:            "cfg-proj",
			wantErr:         false,
		},
		{
			name:            "selected project when no flag or cfg",
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
			selectedProject: "",
			want:            "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfgDir := ".test-session-" + t.Name()
			home := os.Getenv("HOME")
			testHome := filepath.Join(tmpDir, "home")
			os.Setenv("HOME", testHome)
			defer os.Setenv("HOME", home)

			if tt.selectedProject != "" {
				if err := files.SelectProject(cfgDir, tt.selectedProject); err != nil {
					t.Fatalf("SelectProject() error = %v", err)
				}
			}

			s := NewSession(cfgDir)
			got, err := s.ResolveProject(tt.cfgProject, tt.flagProject)

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
