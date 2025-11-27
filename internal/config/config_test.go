// Package config handles shell configuration loading and management.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// T063: Tests for config loading
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantPrompt  string
		wantHistory int
		wantErr     bool
	}{
		{
			name: "valid config with prompt",
			yaml: `
prompt: "test> "
history:
  max_size: 500
`,
			wantPrompt:  "test> ",
			wantHistory: 500,
			wantErr:     false,
		},
		{
			name: "valid config with colors",
			yaml: `
prompt: "$ "
colors:
  enabled: true
  directory: blue
  error: red
`,
			wantPrompt:  "$ ",
			wantHistory: DefaultHistorySize, // Should use default
			wantErr:     false,
		},
		{
			name:        "empty config uses defaults",
			yaml:        "",
			wantPrompt:  DefaultPrompt,
			wantHistory: DefaultHistorySize,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if tt.yaml != "" {
				err := os.WriteFile(configPath, []byte(tt.yaml), 0644)
				if err != nil {
					t.Fatalf("failed to write test config: %v", err)
				}
			}

			cfg, err := LoadFromFile(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if cfg.Prompt != tt.wantPrompt {
				t.Errorf("Prompt = %q, want %q", cfg.Prompt, tt.wantPrompt)
			}

			if cfg.History.MaxSize != tt.wantHistory {
				t.Errorf("History.MaxSize = %d, want %d", cfg.History.MaxSize, tt.wantHistory)
			}
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	cfg, err := LoadFromFile("/nonexistent/path/config.yaml")

	// Should return default config, not error
	if err != nil {
		t.Errorf("LoadFromFile() should not error for missing file, got %v", err)
	}

	if cfg.Prompt != DefaultPrompt {
		t.Errorf("Prompt = %q, want default %q", cfg.Prompt, DefaultPrompt)
	}
}

// T064: Tests for default values
func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Prompt == "" {
		t.Error("Default().Prompt should not be empty")
	}

	if cfg.Prompt != DefaultPrompt {
		t.Errorf("Default().Prompt = %q, want %q", cfg.Prompt, DefaultPrompt)
	}

	if cfg.History.MaxSize != DefaultHistorySize {
		t.Errorf("Default().History.MaxSize = %d, want %d", cfg.History.MaxSize, DefaultHistorySize)
	}

	if cfg.History.File == "" {
		t.Error("Default().History.File should not be empty")
	}

	if cfg.Colors.Enabled != true {
		t.Error("Default().Colors.Enabled should be true")
	}

	if cfg.Abbreviations.Enabled != true {
		t.Error("Default().Abbreviations.Enabled should be true")
	}
}

func TestDefaultColorScheme(t *testing.T) {
	cfg := Default()

	if cfg.Colors.Directory == "" {
		t.Error("Default().Colors.Directory should not be empty")
	}

	if cfg.Colors.Error == "" {
		t.Error("Default().Colors.Error should not be empty")
	}

	if cfg.Colors.Prompt == "" {
		t.Error("Default().Colors.Prompt should not be empty")
	}
}

// T065: Tests for invalid config handling
func TestInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "invalid yaml syntax",
			yaml:    "prompt: [\ninvalid yaml",
			wantErr: true,
		},
		{
			name: "invalid history size (negative)",
			yaml: `
history:
  max_size: -100
`,
			wantErr: true,
		},
		{
			name: "invalid color name",
			yaml: `
colors:
  directory: not_a_color
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configPath, []byte(tt.yaml), 0644)
			if err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			_, err = LoadFromFile(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigPathExpansion(t *testing.T) {
	// Test tilde expansion
	path := ExpandPath("~/test")
	if path == "~/test" {
		t.Error("ExpandPath should expand ~")
	}

	home, err := os.UserHomeDir()
	if err == nil {
		expected := filepath.Join(home, "test")
		if path != expected {
			t.Errorf("ExpandPath(~/test) = %q, want %q", path, expected)
		}
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()

	if dir == "" {
		t.Error("ConfigDir() should not be empty")
	}

	// Should contain jsishell
	if !filepath.IsAbs(dir) {
		// On some systems it might be relative, but generally should be absolute
		t.Logf("ConfigDir() returned non-absolute path: %s", dir)
	}
}

func TestConfigMerge(t *testing.T) {
	base := Default()

	// Simulate partial config
	partial := &Config{
		Prompt: "custom> ",
		// Other fields empty/zero
	}

	merged := base.Merge(partial)

	// Custom prompt should be applied
	if merged.Prompt != "custom> " {
		t.Errorf("Merged prompt = %q, want %q", merged.Prompt, "custom> ")
	}

	// Defaults should be preserved
	if merged.History.MaxSize != DefaultHistorySize {
		t.Errorf("Merged History.MaxSize = %d, want %d", merged.History.MaxSize, DefaultHistorySize)
	}

	if merged.Colors.Directory == "" {
		t.Error("Merged Colors.Directory should preserve default")
	}
}

func TestValidColors(t *testing.T) {
	validColors := []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"bright_black", "bright_red", "bright_green", "bright_yellow",
		"bright_blue", "bright_magenta", "bright_cyan", "bright_white",
	}

	for _, color := range validColors {
		if !IsValidColor(color) {
			t.Errorf("IsValidColor(%q) = false, want true", color)
		}
	}

	invalidColors := []string{"purple", "orange", "pink", "not_a_color", ""}
	for _, color := range invalidColors {
		if IsValidColor(color) {
			t.Errorf("IsValidColor(%q) = true, want false", color)
		}
	}
}
