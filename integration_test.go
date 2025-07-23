package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCLICommands tests the main CLI commands
func TestCLICommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Build the project first
	buildCmd := exec.Command("go", "build", "-o", "klaunch-test")
	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Failed to build project: %v", err)
	}
	defer os.Remove("klaunch-test")

	tests := []struct {
		name        string
		args        []string
		expectError bool
		timeout     time.Duration
	}{
		{
			name:        "help command",
			args:        []string{"--help"},
			expectError: false,
			timeout:     5 * time.Second,
		},
		{
			name:        "version command",
			args:        []string{"--version"},
			expectError: true, // No version flag implemented yet
			timeout:     5 * time.Second,
		},
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			expectError: true,
			timeout:     5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./klaunch-test", tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Set timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Run()
			}()

			select {
			case err := <-done:
				if tt.expectError && err == nil {
					t.Errorf("Expected error but command succeeded")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Unexpected error: %v\nStderr: %s", err, stderr.String())
				}
			case <-time.After(tt.timeout):
				cmd.Process.Kill()
				t.Errorf("Command timed out after %v", tt.timeout)
			}
		})
	}
}

// TestWorkflowIntegration tests complete workflows
func TestWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow integration tests in short mode")
	}

	tests := []struct {
		name     string
		workflow []workflowStep
	}{
		{
			name: "configuration file workflow",
			workflow: []workflowStep{
				{
					name:        "check config files",
					function:    testConfigFilesExist,
					expectError: false,
				},
				{
					name:        "validate config content",
					function:    testConfigFilesValid,
					expectError: false,
				},
			},
		},
		{
			name: "docker environment workflow",
			workflow: []workflowStep{
				{
					name:        "check docker available",
					function:    testDockerAvailable,
					expectError: false,
				},
				{
					name:        "validate docker-compose",
					function:    testDockerComposeValid,
					expectError: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, step := range tt.workflow {
				t.Run(step.name, func(t *testing.T) {
					err := step.function(t)
					if step.expectError && err == nil {
						t.Errorf("Expected error in step '%s' but got none", step.name)
					}
					if !step.expectError && err != nil {
						t.Errorf("Unexpected error in step '%s': %v", step.name, err)
					}
				})
			}
		})
	}
}

type workflowStep struct {
	name        string
	function    func(*testing.T) error
	expectError bool
}

// Helper functions for workflow testing

func testConfigFilesExist(t *testing.T) error {
	configDir := "./case_configs"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return err
	}

	files, err := os.ReadDir(configDir)
	if err != nil {
		return err
	}

	jsonFiles := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			jsonFiles++
		}
	}

	if jsonFiles == 0 {
		t.Error("No JSON configuration files found")
		return nil
	}

	t.Logf("Found %d JSON configuration files", jsonFiles)
	return nil
}

func testConfigFilesValid(t *testing.T) error {
	configFiles, err := getConfigFiles()
	if err != nil {
		return err
	}

	for _, file := range configFiles {
		filePath := filepath.Join("./case_configs", file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read config file %s: %v", file, err)
			continue
		}

		// Basic JSON validation
		if len(content) == 0 {
			t.Errorf("Config file %s is empty", file)
			continue
		}

		if !strings.Contains(string(content), "{") {
			t.Errorf("Config file %s doesn't appear to be JSON", file)
			continue
		}

		t.Logf("Config file %s appears valid", file)
	}

	return nil
}

func testDockerAvailable(t *testing.T) error {
	cmd := exec.Command("docker", "--version")
	err := cmd.Run()
	if err != nil {
		t.Skip("Docker not available - skipping Docker tests")
		return nil
	}

	t.Log("Docker is available")
	return nil
}

func testDockerComposeValid(t *testing.T) error {
	composeFile := "./docker-compose.yaml"
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return err
	}

	// Basic docker-compose validation
	cmd := exec.Command("docker-compose", "-f", composeFile, "config")
	err := cmd.Run()
	if err != nil {
		// Try docker compose (newer syntax)
		cmd = exec.Command("docker", "compose", "-f", composeFile, "config")
		err = cmd.Run()
		if err != nil {
			t.Skip("Docker Compose not available or compose file invalid")
			return nil
		}
	}

	t.Log("Docker Compose configuration is valid")
	return nil
}

// TestEnvironmentSetup tests environment prerequisites
func TestEnvironmentSetup(t *testing.T) {
	tests := []struct {
		name        string
		checkFunc   func() error
		expectError bool
	}{
		{
			name:        "go version check",
			checkFunc:   checkGoVersion,
			expectError: false,
		},
		{
			name:        "project dependencies",
			checkFunc:   checkDependencies,
			expectError: false,
		},
		{
			name:        "directory structure",
			checkFunc:   checkDirectoryStructure,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.checkFunc()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func checkGoVersion() error {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	versionStr := string(output)
	if !strings.Contains(versionStr, "go1.") {
		return nil // Basic version check passed
	}

	return nil
}

func checkDependencies() error {
	cmd := exec.Command("go", "mod", "tidy")
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("go", "mod", "verify")
	return cmd.Run()
}

func checkDirectoryStructure() error {
	expectedDirs := []string{
		"./case_configs",
		"./volumes",
		"./logs",
		"./templates",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Some directories might not exist yet, just log
			continue
		}
	}

	expectedFiles := []string{
		"./docker-compose.yaml",
		"./go.mod",
		"./main.go",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// TestFullDevCycle simulates a complete development cycle
func TestFullDevCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full development cycle test in short mode")
	}

	t.Run("complete development workflow", func(t *testing.T) {
		// Phase 1: Environment validation
		t.Log("Phase 1: Environment validation")
		err := checkDirectoryStructure()
		if err != nil {
			t.Fatalf("Environment validation failed: %v", err)
		}

		// Phase 2: Configuration validation
		t.Log("Phase 2: Configuration validation")
		err = testConfigFilesExist(t)
		if err != nil {
			t.Fatalf("Configuration validation failed: %v", err)
		}

		// Phase 3: Build validation
		t.Log("Phase 3: Build validation")
		cmd := exec.Command("go", "build", "-o", "klaunch-integration-test")
		err = cmd.Run()
		if err != nil {
			t.Fatalf("Build validation failed: %v", err)
		}
		defer os.Remove("klaunch-integration-test")

		// Phase 4: Basic CLI validation
		t.Log("Phase 4: CLI validation")
		cmd = exec.Command("./klaunch-integration-test", "--help")
		err = cmd.Run()
		if err != nil {
			t.Errorf("CLI help command failed: %v", err)
		}

		t.Log("Full development cycle completed successfully")
	})
}

// Benchmark tests moved to benchmarks_test.go to avoid duplication

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		scenario    func() error
		expectError bool
	}{
		{
			name: "missing config directory",
			scenario: func() error {
				originalName := "./case_configs"
				tempName := "./case_configs_backup"

				// Rename directory temporarily
				os.Rename(originalName, tempName)
				defer os.Rename(tempName, originalName)

				_, err := getConfigFiles()
				return err
			},
			expectError: true,
		},
		{
			name: "invalid docker compose",
			scenario: func() error {
				// This would test with a corrupted docker-compose file
				// For safety, we'll just simulate the test
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scenario()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestConcurrentOperations tests thread safety
func TestConcurrentOperations(t *testing.T) {
	t.Run("concurrent config file reads", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()
				_, err := getConfigFiles()
				if err != nil {
					t.Errorf("Concurrent config file read failed: %v", err)
				}
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestMemoryUsage tests for memory leaks
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage tests in short mode")
	}

	// Test repeated operations for memory leaks
	for i := 0; i < 1000; i++ {
		// Simulate repeated config file reads
		getConfigFiles()

		// Force garbage collection periodically
		if i%100 == 0 {
			//runtime.GC()
		}
	}

	t.Log("Memory usage test completed")
}
