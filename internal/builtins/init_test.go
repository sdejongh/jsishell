package builtins

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sdejongh/jsishell/internal/parser"
)

func TestInitDefinition(t *testing.T) {
	def := InitDefinition()

	if def.Name != "init" {
		t.Errorf("Name = %q, want %q", def.Name, "init")
	}

	if def.Handler == nil {
		t.Error("Handler should not be nil")
	}

	// Check options
	hasForce := false
	hasHelp := false
	for _, opt := range def.Options {
		if opt.Long == "--force" && opt.Short == "-f" {
			hasForce = true
		}
		if opt.Long == "--help" {
			hasHelp = true
		}
	}
	if !hasForce {
		t.Error("Should have --force/-f option")
	}
	if !hasHelp {
		t.Error("Should have --help option")
	}
}

func TestInitHelp(t *testing.T) {
	def := InitDefinition()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	execCtx := &Context{
		Stdout: stdout,
		Stderr: stderr,
	}

	cmd := &parser.Command{
		Name:  "init",
		Flags: map[string]bool{"--help": true},
	}

	code, err := def.Handler(context.Background(), cmd, execCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("Exit code = %d, want 0", code)
	}

	output := stdout.String()
	if !strings.Contains(output, "init") {
		t.Error("Help should mention 'init'")
	}
	if !strings.Contains(output, "--force") {
		t.Error("Help should mention --force")
	}
}

func TestInitCreatesConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "jsishell-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CONFIG_HOME to our temp directory
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	def := InitDefinition()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	execCtx := &Context{
		Stdout: stdout,
		Stderr: stderr,
	}

	cmd := &parser.Command{
		Name:  "init",
		Flags: map[string]bool{},
	}

	code, err := def.Handler(context.Background(), cmd, execCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("Exit code = %d, want 0", code)
	}

	// Check config file was created
	configPath := filepath.Join(tmpDir, "jsishell", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}

	// Check content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if !strings.Contains(string(content), "prompt:") {
		t.Error("Config should contain prompt setting")
	}
	if !strings.Contains(string(content), "history:") {
		t.Error("Config should contain history section")
	}
	if !strings.Contains(string(content), "colors:") {
		t.Error("Config should contain colors section")
	}

	// Check output
	if !strings.Contains(stdout.String(), "created") {
		t.Error("Output should indicate config was created")
	}
}

func TestInitExistingConfigNoForce(t *testing.T) {
	// Create a temporary directory with existing config
	tmpDir, err := os.MkdirTemp("", "jsishell-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing config
	configDir := filepath.Join(tmpDir, "jsishell")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yaml")
	os.WriteFile(configPath, []byte("existing: config"), 0644)

	// Set XDG_CONFIG_HOME
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	def := InitDefinition()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	execCtx := &Context{
		Stdout: stdout,
		Stderr: stderr,
	}

	cmd := &parser.Command{
		Name:  "init",
		Flags: map[string]bool{},
	}

	code, err := def.Handler(context.Background(), cmd, execCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code != 1 {
		t.Errorf("Exit code = %d, want 1 (should fail without --force)", code)
	}

	// Check error message
	if !strings.Contains(stderr.String(), "already exists") {
		t.Error("Error should mention file already exists")
	}

	// Check original file is unchanged
	content, _ := os.ReadFile(configPath)
	if string(content) != "existing: config" {
		t.Error("Original config should not be modified")
	}
}

func TestInitForceOverwrite(t *testing.T) {
	// Create a temporary directory with existing config
	tmpDir, err := os.MkdirTemp("", "jsishell-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing config
	configDir := filepath.Join(tmpDir, "jsishell")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yaml")
	os.WriteFile(configPath, []byte("existing: config"), 0644)

	// Set XDG_CONFIG_HOME
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	def := InitDefinition()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	execCtx := &Context{
		Stdout: stdout,
		Stderr: stderr,
	}

	cmd := &parser.Command{
		Name:  "init",
		Flags: map[string]bool{"--force": true},
	}

	code, err := def.Handler(context.Background(), cmd, execCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("Exit code = %d, want 0", code)
	}

	// Check file was overwritten
	content, _ := os.ReadFile(configPath)
	if strings.Contains(string(content), "existing: config") {
		t.Error("Config should have been overwritten")
	}
	if !strings.Contains(string(content), "prompt:") {
		t.Error("Config should contain new default content")
	}
}

func TestDefaultConfigContent(t *testing.T) {
	content := defaultConfigContent()

	requiredSections := []string{
		"prompt:",
		"history:",
		"colors:",
		"abbreviations:",
		"editor:",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("Default config should contain %q", section)
		}
	}

	// Check documentation comments
	if !strings.Contains(content, "%d") {
		t.Error("Should document percent-d variable")
	}
	if !strings.Contains(content, "%{") {
		t.Error("Should document color syntax")
	}
}
