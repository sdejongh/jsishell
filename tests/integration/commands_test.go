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

// TestCdPwd tests the cd and pwd commands.
// T040: Write integration test for cd/pwd
func TestCdPwd(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test 'pwd' command - should show current directory
	code, err := e.ExecuteInput(ctx, "pwd")
	if err != nil {
		t.Errorf("pwd error: %v", err)
	}
	if code != 0 {
		t.Errorf("pwd exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), tmpDir) {
		t.Errorf("pwd output = %q, want to contain %q", stdout.String(), tmpDir)
	}

	// Test 'cd' command to subdir
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "cd subdir")
	if err != nil {
		t.Errorf("cd error: %v", err)
	}
	if code != 0 {
		t.Errorf("cd exit code = %d, want 0", code)
	}

	// Verify we're in subdir
	stdout.Reset()
	code, err = e.ExecuteInput(ctx, "pwd")
	if err != nil {
		t.Errorf("pwd error: %v", err)
	}
	if !strings.Contains(stdout.String(), "subdir") {
		t.Errorf("after cd, pwd output = %q, want to contain 'subdir'", stdout.String())
	}

	// Test 'cd' to parent directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "cd ..")
	if err != nil {
		t.Errorf("cd .. error: %v", err)
	}
	if code != 0 {
		t.Errorf("cd .. exit code = %d, want 0", code)
	}

	// Test 'cd' to non-existent directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "cd nonexistent12345")
	if code == 0 {
		t.Error("cd to nonexistent dir should fail")
	}
}

// TestLsCommand tests the ls command.
// T041: Write integration test for ls command
func TestLsCommand(t *testing.T) {
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

	// Test basic ls
	code, err := e.ExecuteInput(ctx, "ls")
	if err != nil {
		t.Errorf("ls error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Should show visible files
	if !strings.Contains(output, "file1.txt") {
		t.Errorf("ls output should contain file1.txt, got: %s", output)
	}
	if !strings.Contains(output, "subdir") {
		t.Errorf("ls output should contain subdir, got: %s", output)
	}
	// Should NOT show hidden files by default
	if strings.Contains(output, ".hidden") {
		t.Errorf("ls output should not contain .hidden by default, got: %s", output)
	}

	// Test ls with --all flag
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "ls --all")
	if err != nil {
		t.Errorf("ls --all error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls --all exit code = %d, want 0", code)
	}

	output = stdout.String()
	// Should show hidden files
	if !strings.Contains(output, ".hidden") {
		t.Errorf("ls --all output should contain .hidden, got: %s", output)
	}

	// Test ls specific directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "ls "+tmpDir)
	if err != nil {
		t.Errorf("ls <dir> error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls <dir> exit code = %d, want 0", code)
	}

	// Test ls non-existent directory
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "ls /nonexistent12345")
	if code == 0 {
		t.Error("ls nonexistent dir should fail")
	}
}

// TestCpMvRm tests cp, mv, and rm commands.
// T042: Write integration test for cp/mv/rm
func TestCpMvRm(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	srcFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test cp command
	dstFile := filepath.Join(tmpDir, "destination.txt")
	code, err := e.ExecuteInput(ctx, "cp source.txt destination.txt")
	if err != nil {
		t.Errorf("cp error: %v", err)
	}
	if code != 0 {
		t.Errorf("cp exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify cp succeeded
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("cp failed: destination file does not exist")
	}
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("cp should not remove source file")
	}

	// Test mv command
	stdout.Reset()
	stderr.Reset()

	movedFile := filepath.Join(tmpDir, "moved.txt")
	code, err = e.ExecuteInput(ctx, "mv destination.txt moved.txt")
	if err != nil {
		t.Errorf("mv error: %v", err)
	}
	if code != 0 {
		t.Errorf("mv exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify mv succeeded
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Error("mv failed: destination file does not exist")
	}
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		t.Error("mv should remove source file")
	}

	// Test rm command
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "rm moved.txt")
	if err != nil {
		t.Errorf("rm error: %v", err)
	}
	if code != 0 {
		t.Errorf("rm exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify rm succeeded
	if _, err := os.Stat(movedFile); !os.IsNotExist(err) {
		t.Error("rm failed: file still exists")
	}

	// Test rm non-existent file
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "rm nonexistent12345.txt")
	if code == 0 {
		t.Error("rm nonexistent file should fail")
	}
}

// TestMkdir tests the mkdir command.
// T043: Write integration test for mkdir
func TestMkdir(t *testing.T) {
	tmpDir := t.TempDir()

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test creating a simple directory
	newDir := filepath.Join(tmpDir, "newdir")
	code, err := e.ExecuteInput(ctx, "mkdir newdir")
	if err != nil {
		t.Errorf("mkdir error: %v", err)
	}
	if code != 0 {
		t.Errorf("mkdir exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify directory was created
	info, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		t.Error("mkdir failed: directory does not exist")
	}
	if !info.IsDir() {
		t.Error("mkdir created a file instead of a directory")
	}

	// Test creating nested directories with --parents
	stdout.Reset()
	stderr.Reset()

	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	code, err = e.ExecuteInput(ctx, "mkdir --parents a/b/c")
	if err != nil {
		t.Errorf("mkdir --parents error: %v", err)
	}
	if code != 0 {
		t.Errorf("mkdir --parents exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify nested directory was created
	info, err = os.Stat(nestedDir)
	if os.IsNotExist(err) {
		t.Error("mkdir --parents failed: nested directory does not exist")
	}
	if !info.IsDir() {
		t.Error("mkdir --parents created a file instead of a directory")
	}

	// Test creating existing directory (should fail without --parents)
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "mkdir newdir")
	if code == 0 {
		t.Error("mkdir on existing dir should fail")
	}
}

// TestRecursiveOperations tests recursive cp and rm.
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

	// Test recursive cp
	dstDir := filepath.Join(tmpDir, "dest_dir")
	code, err := e.ExecuteInput(ctx, "cp --recursive source_dir dest_dir")
	if err != nil {
		t.Errorf("cp --recursive error: %v", err)
	}
	if code != 0 {
		t.Errorf("cp --recursive exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify recursive cp
	if _, err := os.Stat(filepath.Join(dstDir, "file.txt")); os.IsNotExist(err) {
		t.Error("recursive cp failed: file.txt not copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "subdir", "nested.txt")); os.IsNotExist(err) {
		t.Error("recursive cp failed: nested file not copied")
	}

	// Test recursive rm
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "rm --recursive dest_dir")
	if err != nil {
		t.Errorf("rm --recursive error: %v", err)
	}
	if code != 0 {
		t.Errorf("rm --recursive exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Verify recursive rm
	if _, err := os.Stat(dstDir); !os.IsNotExist(err) {
		t.Error("recursive rm failed: directory still exists")
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
	code, err := e.ExecuteInput(ctx, "ls testfile.txt")
	if err != nil {
		t.Errorf("ls relative path error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls relative path exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	// Test with absolute path
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "ls "+testFile)
	if err != nil {
		t.Errorf("ls absolute path error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls absolute path exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
}

// TestLsLongFormat tests the ls --long option.
func TestLsLongFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	if err := os.WriteFile(filepath.Join(tmpDir, "testfile.txt"), []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test ls with --long flag
	code, err := e.ExecuteInput(ctx, "ls --long")
	if err != nil {
		t.Errorf("ls --long error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls --long exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Long format should include permissions or size info
	if !strings.Contains(output, "testfile.txt") {
		t.Errorf("ls --long output should contain testfile.txt, got: %s", output)
	}
}

// TestLsVerboseFormat tests the ls --verbose option.
func TestLsVerboseFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	if err := os.WriteFile(filepath.Join(tmpDir, "testfile.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	code, err := e.ExecuteInput(ctx, "ls --verbose")
	if err != nil {
		t.Errorf("ls --verbose error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls --verbose exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Verbose format should include file type indicator
	if !strings.Contains(output, "[file]") {
		t.Errorf("ls --verbose output should contain [file], got: %s", output)
	}
}

// TestLsQuietMode tests the ls --quiet option.
func TestLsQuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	code, err := e.ExecuteInput(ctx, "ls --quiet")
	if err != nil {
		t.Errorf("ls --quiet error: %v", err)
	}
	if code != 0 {
		t.Errorf("ls --quiet exit code = %d, want 0", code)
	}

	output := stdout.String()
	// Quiet mode should only show file names
	if output != "file.txt\n" {
		t.Errorf("ls --quiet output = %q, want %q", output, "file.txt\n")
	}
}

// TestRmYesFlag tests the rm --yes flag.
func TestRmYesFlag(t *testing.T) {
	tmpDir := t.TempDir()

	e, _, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test that --yes suppresses error for missing file (like --force)
	code, _ := e.ExecuteInput(ctx, "rm --yes nonexistent.txt")

	if code != 0 {
		t.Errorf("rm --yes should return 0 for missing file, got %d", code)
	}
	if stderr.Len() > 0 {
		t.Errorf("rm --yes should not produce error output, got: %s", stderr.String())
	}
}

// TestRmQuietFlag tests the rm --quiet flag.
func TestRmQuietFlag(t *testing.T) {
	tmpDir := t.TempDir()

	e, stdout, stderr := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	// Test that --quiet suppresses error messages
	code, _ := e.ExecuteInput(ctx, "rm --quiet nonexistent.txt")

	// Exit code may be non-zero, but no error output
	_ = code
	if stderr.Len() > 0 {
		t.Errorf("rm --quiet should suppress errors, got stderr: %s", stderr.String())
	}
	if stdout.Len() > 0 {
		t.Errorf("rm --quiet should produce no stdout, got: %s", stdout.String())
	}
}

// TestHelpFlag tests --help flag on various commands.
func TestHelpFlag(t *testing.T) {
	tmpDir := t.TempDir()
	e, stdout, _ := setupTestExecutor(t, tmpDir)
	ctx := context.Background()

	commands := []string{"ls", "cp", "mv", "rm", "mkdir", "cd", "pwd", "echo"}

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
