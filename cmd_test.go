package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ml-reg/internal/cmd"
)

// TestCLIBasic tests basic CLI functionality without external dependencies.
// This includes help output, error handling, and command parsing.
func TestCLIBasic(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int
		wantOut  []string
		wantErr  []string
	}{
		{
			name:     "no_args_shows_help",
			args:     []string{"ml-reg"},
			wantExit: 0,
			wantOut:  []string{"ml-reg is a single-binary CLI", "Usage:", "init", "push", "pull", "list"},
		},
		{
			name:     "help_flag",
			args:     []string{"ml-reg", "--help"},
			wantExit: 0,
			wantOut:  []string{"ml-reg is a single-binary CLI", "Usage:", "init", "push", "pull", "list"},
		},
		{
			name:     "unknown_command",
			args:     []string{"ml-reg", "unknown"},
			wantExit: 1,
			wantErr:  []string{"unknown"},
		},
		{
			name:     "init_missing_required_flags",
			args:     []string{"ml-reg", "init"},
			wantExit: 2,
			wantErr:  []string{"required flag", "endpoint", "bucket"},
		},
		{
			name:     "push_missing_required_flags",
			args:     []string{"ml-reg", "push", "test.pkl"},
			wantExit: 2,
			wantErr:  []string{"required flag", "tag"},
		},
		{
			name:     "pull_wrong_arg_count",
			args:     []string{"ml-reg", "pull", "v1"},
			wantExit: 1, // Cobra returns 1 for wrong argument count
			wantErr:  []string{"accepts 2 arg(s)", "received 1"},
		},
		{
			name:     "list_no_args",
			args:     []string{"ml-reg", "list"},
			wantExit: 3, // ErrNotInitialized since registry not initialized
			wantErr:  []string{"registry is not initialized"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tempDir := t.TempDir()
			originalWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(originalWd)

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			stdoutR, stdoutW, _ := os.Pipe()
			stderrR, stderrW, _ := os.Pipe()
			os.Stdout = stdoutW
			os.Stderr = stderrW

			// Set command-line arguments
			os.Args = tt.args

			// Run command
			exitCode := cmd.Execute()

			// Restore stdout/stderr and read output
			stdoutW.Close()
			stderrW.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			stdoutBytes, _ := io.ReadAll(stdoutR)
			stderrBytes, _ := io.ReadAll(stderrR)
			stdout := string(stdoutBytes)
			stderr := string(stderrBytes)

			// Check exit code
			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("stdout:\n%s", stdout)
				t.Logf("stderr:\n%s", stderr)
			}

			// Check stdout content
			for _, want := range tt.wantOut {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout doesn't contain %q\nstdout:\n%s", want, stdout)
				}
			}

			// Check stderr content
			for _, want := range tt.wantErr {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr doesn't contain %q\nstderr:\n%s", want, stderr)
				}
			}
		})
	}
}

// TestInitForceFlag tests the --force flag behavior.
func TestInitForceFlag(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// First init should succeed
	os.Args = []string{
		"ml-reg", "init",
		"--endpoint", "http://localhost:9000",
		"--bucket", "test-bucket",
		"--access-key", "minioadmin",
		"--secret-key", "minioadmin",
		"--force",
	}

	exitCode := cmd.Execute()
	if exitCode != 0 {
		t.Fatalf("first init failed with exit code %d", exitCode)
	}

	// Second init without force should fail
	os.Args = []string{
		"ml-reg", "init",
		"--endpoint", "http://localhost:9001",
		"--bucket", "different-bucket",
		"--access-key", "minioadmin",
		"--secret-key", "minioadmin",
		// No --force flag
	}

	exitCode = cmd.Execute()
	if exitCode != 4 { // ErrAlreadyInitialized = exit code 4
		t.Errorf("second init without force should fail with exit code 4, got %d", exitCode)
	}

	// Third init with force should succeed
	os.Args = []string{
		"ml-reg", "init",
		"--endpoint", "http://localhost:9002",
		"--bucket", "another-bucket",
		"--access-key", "minioadmin",
		"--secret-key", "minioadmin",
		"--force",
	}

	exitCode = cmd.Execute()
	if exitCode != 0 {
		t.Errorf("third init with force should succeed, got exit code %d", exitCode)
	}
}

// TestFileOperations tests file-based operations with a mock.
// This test is disabled as it requires a fully initialized registry with valid
// AWS credentials and proper database, which is complex for a simple unit test.
// The file validation logic is tested in registry unit tests.
func TestFileOperations(t *testing.T) {
	t.Skip("Skipping test that requires fully initialized registry with AWS credentials")
}

// TestExitCodeMapping tests that errors map to correct exit codes.
func TestExitCodeMapping(t *testing.T) {
	// Note: This test would be more comprehensive with mocked dependencies
	// For now, we test the basic error mapping through the registry package
	
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create .ml-reg directory and empty config to trigger config loading error
	mlRegDir := filepath.Join(tempDir, ".ml-reg")
	os.MkdirAll(mlRegDir, 0755)

	// Test list without proper config (should fail with ErrNotInitialized = exit code 3)
	os.Args = []string{"ml-reg", "list"}

	exitCode := cmd.Execute()
	if exitCode != 3 {
		t.Errorf("list without config should fail with exit code 3, got %d", exitCode)
	}
}

// captureOutput captures stdout/stderr during test execution.
func captureOutput(f func()) (stdout, stderr string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	f()

	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdoutBytes, _ := io.ReadAll(stdoutR)
	stderrBytes, _ := io.ReadAll(stderrR)

	return string(stdoutBytes), string(stderrBytes)
}

// Example usage output for documentation.
func Example() {
	// This shows how the CLI can be used programmatically
	os.Args = []string{"ml-reg", "--help"}
	exitCode := cmd.Execute()
	fmt.Printf("Exit code: %d\n", exitCode)
	// Output would show help text
}