// Package integration provides integration tests for JSIShell file system commands.
package integration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sdejongh/jsishell/internal/builtins"
	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/executor"
)

// setupTestExecutor creates a new executor with all builtins registered.
// It also changes the process working directory to match the executor's workDir.
func setupTestExecutor(t *testing.T, workDir string) (*executor.Executor, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original working directory: %v", err)
	}

	// Change to the test working directory
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to change to workDir: %v", err)
	}

	// Register cleanup to restore original directory
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	reg := builtins.NewRegistry()
	builtins.RegisterAll(reg)

	environment := env.New()
	environment.Set("PWD", workDir)

	e := executor.New(
		executor.WithRegistry(reg),
		executor.WithEnv(environment),
		executor.WithStdout(stdout),
		executor.WithStderr(stderr),
		executor.WithWorkDir(workDir),
	)

	return e, stdout, stderr
}

// TestGotoHere tests the goto (cd) and here (pwd) commands.
// T040: Write integration test for cd/pwd
func TestGotoHere(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test 'here' command - should show current directory
	code, err := e.ExecuteInput(ctx, "here")
	if err != nil {
		t.Errorf("here error: %v", err)
	}
	if code != 0 {
		t.Errorf("here exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), tmpDir) {
		t.Errorf("here output = %q, want to contain %q", stdout.String(), tmpDir)
	}

	// Test 'goto' command to subdir
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "goto subdir")
	if err != nil {
		t.Errorf("goto error: %v", err)
	}
	if code != 0 {
		t.Errorf("goto exit code = %d, want 0", code)
	}

	// Verify we're in subdir
	stdout.Reset()
	code, err = e.ExecuteInput(ctx, "here")
	if err != nil {
		t.Errorf("here error: %v", err)
	}
	if !strings.Contains(stdout.String(), "subdir") {
		t.Errorf("after goto, here output = %q, want to contain 'subdir'", stdout.String())
	}

	// Test 'goto' to parent directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "goto ..")
	if err != nil {
		t.Errorf("goto .. error: %v", err)
	}
	if code != 0 {
		t.Errorf("goto .. exit code = %d, want 0", code)
	}

	// Test 'goto' to non-existent directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "goto nonexistent12345")
	if code == 0 {
		t.Error("goto to nonexistent dir should fail")
	}
}

// TestListCommand tests the list command.
// T041: Write integration test for list command
func TestListCommand(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create some test files and directories
	testFiles := []string{"file1.txt", "file2.txt", ".hidden"}
	for _, name := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, name))
		if err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
		f.Close()
	}

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test basic list
	code, err := e.ExecuteInput(ctx, "list")
	if err != nil {
		t.Errorf("list error: %v", err)
	}
	if code != 0 {
		t.Errorf("list exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Should show visible files
	if !strings.Contains(output, "file1.txt") {
		t.Errorf("list output should contain file1.txt, got: %s", output)
	}
	if !strings.Contains(output, "subdir") {
		t.Errorf("list output should contain subdir, got: %s", output)
	}
	// Should NOT show hidden files by default
	if strings.Contains(output, ".hidden") {
		t.Errorf("list output should not contain .hidden by default, got: %s", output)
	}

	// Test list with --all flag
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "list --all")
	if err != nil {
		t.Errorf("list --all error: %v", err)
	}
	if code != 0 {
		t.Errorf("list --all exit code = %d, want 0", code)
	}

	output = stdout.String()
	// Should show hidden files
	if !strings.Contains(output, ".hidden") {
		t.Errorf("list --all output should contain .hidden, got: %s", output)
	}

	// Test list specific directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "list "+tmpDir)
	if err != nil {
		t.Errorf("list <dir> error: %v", err)
	}
	if code != 0 {
		t.Errorf("list <dir> exit code = %d, want 0", code)
	}

	// Test list non-existent directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "list /nonexistent12345")
	if code == 0 {
		t.Error("list nonexistent dir should fail")
	}
}

// TestCopyMoveRemove tests copy, move, and remove commands.
// T042: Write integration test for copy/move/remove
func TestCopyMoveRemove(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	srcFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test copy command
	dstFile := filepath.Join(tmpDir, "destination.txt")
	code, err := e.ExecuteInput(ctx, "copy source.txt destination.txt")
	if err != nil {
		t.Errorf("copy error: %v", err)
	}
	if code != 0 {
		t.Errorf("copy exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify copy succeeded
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("copy failed: destination file does not exist")
	}
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("copy should not remove source file")
	}

	// Test move command
	stdout.Reset()
	stderr.Reset()

	movedFile := filepath.Join(tmpDir, "moved.txt")
	code, err = e.ExecuteInput(ctx, "move destination.txt moved.txt")
	if err != nil {
		t.Errorf("move error: %v", err)
	}
	if code != 0 {
		t.Errorf("move exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify move succeeded
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Error("move failed: destination file does not exist")
	}
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		t.Error("move should remove source file")
	}

	// Test remove command
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "remove moved.txt")
	if err != nil {
		t.Errorf("remove error: %v", err)
	}
	if code != 0 {
		t.Errorf("remove exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify remove succeeded
	if _, err := os.Stat(movedFile); !os.IsNotExist(err) {
		t.Error("remove failed: file still exists")
	}

	// Test remove non-existent file
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "remove nonexistent12345.txt")
	if code == 0 {
		t.Error("remove nonexistent file should fail")
	}
}

// TestMakedir tests the makedir (mkdir) command.
// T043: Write integration test for mkdir
func TestMakedir(t *testing.T) {
	tmpDir := t.TempDir()

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test creating a simple directory
	newDir := filepath.Join(tmpDir, "newdir")
	code, err := e.ExecuteInput(ctx, "makedir newdir")
	if err != nil {
		t.Errorf("makedir error: %v", err)
	}
	if code != 0 {
		t.Errorf("makedir exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify directory was created
	info, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		t.Error("makedir failed: directory does not exist")
	}
	if !info.IsDir() {
		t.Error("makedir created a file instead of a directory")
	}

	// Test creating nested directories with --parents
	stdout.Reset()
	stderr.Reset()

	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	code, err = e.ExecuteInput(ctx, "makedir --parents a/b/c")
	if err != nil {
		t.Errorf("makedir --parents error: %v", err)
	}
	if code != 0 {
		t.Errorf("makedir --parents exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify nested directory was created
	info, err = os.Stat(nestedDir)
	if os.IsNotExist(err) {
		t.Error("makedir --parents failed: nested directory does not exist")
	}
	if !info.IsDir() {
		t.Error("makedir --parents created a file instead of a directory")
	}

	// Test creating existing directory (should fail without --parents)
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "makedir newdir")
	if code == 0 {
		t.Error("makedir on existing dir should fail")
	}
}

// TestRecursiveOperations tests recursive copy and remove.
func TestRecursiveOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory structure
	srcDir := filepath.Join(tmpDir, "source_dir")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test recursive copy
	dstDir := filepath.Join(tmpDir, "dest_dir")
	code, err := e.ExecuteInput(ctx, "copy --recursive source_dir dest_dir")
	if err != nil {
		t.Errorf("copy --recursive error: %v", err)
	}
	if code != 0 {
		t.Errorf("copy --recursive exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify recursive copy
	if _, err := os.Stat(filepath.Join(dstDir, "file.txt")); os.IsNotExist(err) {
		t.Error("recursive copy failed: file.txt not copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "subdir", "nested.txt")); os.IsNotExist(err) {
		t.Error("recursive copy failed: nested file not copied")
	}

	// Test recursive remove
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "remove --recursive dest_dir")
	if err != nil {
		t.Errorf("remove --recursive error: %v", err)
	}
	if code != 0 {
		t.Errorf("remove --recursive exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify recursive remove
	if _, err := os.Stat(dstDir); !os.IsNotExist(err) {
		t.Error("recursive remove failed: directory still exists")
	}
}

// TestAbsoluteAndRelativePaths tests that both path types work correctly.
func TestAbsoluteAndRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test with relative path
	code, err := e.ExecuteInput(ctx, "list testfile.txt")
	if err != nil {
		t.Errorf("list relative path error: %v", err)
	}
	if code != 0 {
		t.Errorf("list relative path exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Test with absolute path
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "list "+testFile)
	if err != nil {
		t.Errorf("list absolute path error: %v", err)
	}
	if code != 0 {
		t.Errorf("list absolute path exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
}

// TestListLongFormat tests the list --long option.
func TestListLongFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	if err := os.WriteFile(filepath.Join(tmpDir, "testfile.txt"), []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test list with --long flag
	code, err := e.ExecuteInput(ctx, "list --long")
	if err != nil {
		t.Errorf("list --long error: %v", err)
	}
	if code != 0 {
		t.Errorf("list --long exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Long format should include permissions or size info
	if !strings.Contains(output, "testfile.txt") {
		t.Errorf("list --long output should contain testfile.txt, got: %s", output)
	}
}

// TestListVerboseFormat tests the list --verbose option.
func TestListVerboseFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	if err := os.WriteFile(filepath.Join(tmpDir, "testfile.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	code, err := e.ExecuteInput(ctx, "list --verbose")
	if err != nil {
		t.Errorf("list --verbose error: %v", err)
	}
	if code != 0 {
		t.Errorf("list --verbose exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Verbose format should include file type indicator
	if !strings.Contains(output, "[file]") {
		t.Errorf("list --verbose output should contain [file], got: %s", output)
	}
}

// TestListQuietMode tests the list --quiet option.
func TestListQuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	code, err := e.ExecuteInput(ctx, "list --quiet")
	if err != nil {
		t.Errorf("list --quiet error: %v", err)
	}
	if code != 0 {
		t.Errorf("list --quiet exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Quiet mode should only show file names
	if output != "file.txt\n" {
		t.Errorf("list --quiet output = %q, want %q", output, "file.txt\n")
	}
}

// TestRemoveYesFlag tests the remove --yes flag.
func TestRemoveYesFlag(t *testing.T) {
	tmpDir := t.TempDir()

	e, _, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test that --yes suppresses error for missing file (like --force)
	code, _ := e.ExecuteInput(ctx, "remove --yes nonexistent.txt")

	if code != 0 {
		t.Errorf("remove --yes should return 0 for missing file, got %d", code)
	}
	if stderr.Len() > 0 {
		t.Errorf("remove --yes should not produce error output, got: %s", stderr.String())
	}
}

// TestRemoveQuietFlag tests the remove --quiet flag.
func TestRemoveQuietFlag(t *testing.T) {
	tmpDir := t.TempDir()

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test that --quiet suppresses error messages
	code, _ := e.ExecuteInput(ctx, "remove --quiet nonexistent.txt")

	// Exit code may be non-zero, but no error output
	_ = code
	if stderr.Len() > 0 {
		t.Errorf("remove --quiet should suppress errors, got stderr: %s", stderr.String())
	}
	if stdout.Len() > 0 {
		t.Errorf("remove --quiet should produce no stdout, got: %s", stdout.String())
	}
}

// TestHelpFlag tests --help flag on various commands.
func TestHelpFlag(t *testing.T) {
	tmpDir := t.TempDir()
	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	commands := []string{"list", "copy", "move", "remove", "makedir", "goto", "here", "echo"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			stdout.Reset()

			code, err := e.ExecuteInput(ctx, cmd+" --help")
			if err != nil {
				t.Errorf("%s --help error: %v", cmd, err)
			}
			if code != 0 {
				t.Errorf("%s --help exit code = %d, want 0", cmd, code)
			}

			output := stdout.String()
			if !strings.Contains(output, "Usage:") {
				t.Errorf("%s --help should contain 'Usage:', got: %s", cmd, output)
			}
		})
	}
}
