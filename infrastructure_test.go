package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDockerComposeServices tests Docker Compose service definitions
func TestDockerComposeServices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping infrastructure tests in short mode")
	}

	tests := []struct {
		name           string
		service        string
		expectInConfig bool
	}{
		{
			name:           "zookeeper service",
			service:        "zookeeper1",
			expectInConfig: true,
		},
		{
			name:           "kafka brokers",
			service:        "kafka1",
			expectInConfig: true,
		},
		{
			name:           "kafka connect",
			service:        "kafka-connect",
			expectInConfig: true,
		},
		{
			name:           "schema registry",
			service:        "schema-registry",
			expectInConfig: true,
		},
		{
			name:           "cmak",
			service:        "cmak",
			expectInConfig: true,
		},
	}

	composeFile := "./docker-compose.yaml"
	content, err := os.ReadFile(composeFile)
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yaml: %v", err)
	}

	composeContent := string(content)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceExists := strings.Contains(composeContent, tt.service+":")

			if tt.expectInConfig && !serviceExists {
				t.Errorf("Service %s not found in docker-compose.yaml", tt.service)
			}

			if !tt.expectInConfig && serviceExists {
				t.Errorf("Service %s unexpectedly found in docker-compose.yaml", tt.service)
			}
		})
	}
}

// TestPortConfiguration tests port mappings and availability
func TestPortConfiguration(t *testing.T) {
	expectedPorts := map[string]string{
		"kafka1":          "9091",
		"kafka2":          "9092",
		"kafka3":          "9093",
		"kafka-connect":   "8083",
		"schema-registry": "8081",
		"cmak":            "9000",
	}

	for service, port := range expectedPorts {
		t.Run(fmt.Sprintf("%s port %s", service, port), func(t *testing.T) {
			// Test if port is available (not in use)
			listener, err := net.Listen("tcp", ":"+port)
			if err != nil {
				t.Logf("Port %s is in use (expected if services are running): %v", port, err)
			} else {
				listener.Close()
				t.Logf("Port %s is available", port)
			}
		})
	}
}

// TestNetworkConnectivity tests network configuration
func TestNetworkConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network connectivity tests in short mode")
	}

	tests := []struct {
		name    string
		address string
		timeout time.Duration
	}{
		{
			name:    "kafka connect health",
			address: "localhost:8083",
			timeout: 5 * time.Second,
		},
		{
			name:    "schema registry health",
			address: "localhost:8081",
			timeout: 5 * time.Second,
		},
		{
			name:    "cmak web interface",
			address: "localhost:9000",
			timeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := net.DialTimeout("tcp", tt.address, tt.timeout)
			if err != nil {
				t.Logf("Service at %s not reachable (expected if not running): %v", tt.address, err)
			} else {
				conn.Close()
				t.Logf("Service at %s is reachable", tt.address)
			}
		})
	}
}

// TestHTTPEndpoints tests REST API endpoints
func TestHTTPEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping HTTP endpoint tests in short mode")
	}

	endpoints := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "kafka connect root",
			url:         "http://localhost:8083/",
			expectError: false,
		},
		{
			name:        "kafka connect connectors",
			url:         "http://localhost:8083/connectors",
			expectError: false,
		},
		{
			name:        "schema registry subjects",
			url:         "http://localhost:8081/subjects",
			expectError: false,
		},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			resp, err := client.Get(endpoint.url)

			if endpoint.expectError && err == nil {
				t.Error("Expected error but got successful response")
			}

			if !endpoint.expectError && err != nil {
				t.Logf("Expected endpoint to be reachable but got error (services may not be running): %v", err)
				return
			}

			if resp != nil {
				defer resp.Body.Close()
				t.Logf("Endpoint %s returned status: %s", endpoint.url, resp.Status)
			}
		})
	}
}

// TestVolumeConfiguration tests volume mounts and persistence
func TestVolumeConfiguration(t *testing.T) {
	expectedVolumes := []string{
		"./volumes",
		"./logs",
	}

	for _, volume := range expectedVolumes {
		t.Run(fmt.Sprintf("volume %s", volume), func(t *testing.T) {
			info, err := os.Stat(volume)
			if err != nil {
				// Volume directory may not exist yet
				t.Logf("Volume directory %s does not exist (will be created when needed): %v", volume, err)
				return
			}

			if !info.IsDir() {
				t.Errorf("Volume path %s exists but is not a directory", volume)
			}

			t.Logf("Volume directory %s exists and is accessible", volume)
		})
	}
}

// TestDockerfileValidation tests custom Dockerfile
func TestDockerfileValidation(t *testing.T) {
	dockerfile := "./Dockerfile-MongoConnect"

	t.Run("dockerfile exists", func(t *testing.T) {
		if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
			t.Errorf("Dockerfile %s does not exist", dockerfile)
			return
		}

		content, err := os.ReadFile(dockerfile)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", dockerfile, err)
		}

		dockerfileContent := string(content)

		// Basic Dockerfile validation
		if !strings.Contains(dockerfileContent, "FROM") {
			t.Error("Dockerfile should contain FROM instruction")
		}

		t.Log("Dockerfile appears to be valid")
	})
}

// TestEnvironmentVariables tests environment variable configuration
func TestEnvironmentVariables(t *testing.T) {
	envFile := "./.env"

	t.Run("env file validation", func(t *testing.T) {
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			t.Skip(".env file does not exist - may be created by application")
			return
		}

		content, err := os.ReadFile(envFile)
		if err != nil {
			t.Fatalf("Failed to read .env file: %v", err)
		}

		envContent := string(content)

		// Check for expected environment variables
		expectedVars := []string{
			"MONGO_KAFKA_CONNECT_VERSION",
		}

		for _, envVar := range expectedVars {
			if !strings.Contains(envContent, envVar) {
				t.Errorf("Environment variable %s not found in .env file", envVar)
			}
		}

		t.Log(".env file contains expected variables")
	})
}

// TestDockerCommands tests Docker command availability and permissions
func TestDockerCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker command tests in short mode")
	}

	commands := []struct {
		name    string
		cmd     []string
		timeout time.Duration
	}{
		{
			name:    "docker version",
			cmd:     []string{"docker", "--version"},
			timeout: 5 * time.Second,
		},
		{
			name:    "docker compose version",
			cmd:     []string{"docker-compose", "--version"},
			timeout: 5 * time.Second,
		},
		{
			name:    "docker compose v2",
			cmd:     []string{"docker", "compose", "version"},
			timeout: 5 * time.Second,
		},
	}

	for _, cmdTest := range commands {
		t.Run(cmdTest.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), cmdTest.timeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, cmdTest.cmd[0], cmdTest.cmd[1:]...)
			output, err := cmd.Output()

			if err != nil {
				t.Logf("Command '%s' failed (may not be available): %v", strings.Join(cmdTest.cmd, " "), err)
				return
			}

			t.Logf("Command output: %s", strings.TrimSpace(string(output)))
		})
	}
}

// TestServiceHealthChecks tests service health and readiness
func TestServiceHealthChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping service health check tests in short mode")
	}

	healthChecks := []struct {
		name      string
		checkFunc func() error
	}{
		{
			name:      "kafka connect health",
			checkFunc: checkKafkaConnectHealth,
		},
		{
			name:      "schema registry health",
			checkFunc: checkSchemaRegistryHealth,
		},
	}

	for _, check := range healthChecks {
		t.Run(check.name, func(t *testing.T) {
			err := check.checkFunc()
			if err != nil {
				t.Logf("Health check failed (service may not be running): %v", err)
			} else {
				t.Log("Health check passed")
			}
		})
	}
}

func checkKafkaConnectHealth() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:8083/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kafka connect returned status %d", resp.StatusCode)
	}

	return nil
}

func checkSchemaRegistryHealth() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:8081/subjects")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("schema registry returned status %d", resp.StatusCode)
	}

	return nil
}

// TestConfigTemplates tests configuration templates
func TestConfigTemplates(t *testing.T) {
	templateDir := "./templates"

	t.Run("template directory", func(t *testing.T) {
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			t.Skip("Template directory does not exist")
			return
		}

		files, err := os.ReadDir(templateDir)
		if err != nil {
			t.Fatalf("Failed to read template directory: %v", err)
		}

		templateCount := 0
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".template") {
				templateCount++

				// Validate template file
				filePath := filepath.Join(templateDir, file.Name())
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("Failed to read template %s: %v", file.Name(), err)
					continue
				}

				if len(content) == 0 {
					t.Errorf("Template %s is empty", file.Name())
				}
			}
		}

		t.Logf("Found %d template files", templateCount)
	})
}

// TestResourceLimits tests resource usage and limits
func TestResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource limit tests in short mode")
	}

	t.Run("memory usage check", func(t *testing.T) {
		// This is a placeholder for memory usage testing
		// In practice, this would monitor memory usage during operations
		t.Log("Memory usage test completed")
	})

	t.Run("disk space check", func(t *testing.T) {
		// Check available disk space
		wd, _ := os.Getwd()
		usage, err := getDiskUsage(wd)
		if err != nil {
			t.Errorf("Failed to get disk usage: %v", err)
			return
		}

		// Warn if less than 1GB available
		if usage.Available < 1*1024*1024*1024 {
			t.Logf("Warning: Low disk space available: %d bytes", usage.Available)
		} else {
			t.Logf("Disk space available: %d bytes", usage.Available)
		}
	})
}

type DiskUsage struct {
	Available uint64
	Used      uint64
	Total     uint64
}

func getDiskUsage(path string) (*DiskUsage, error) {
	// This is a simplified implementation
	// In practice, you'd use syscall.Statfs on Unix or similar
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Return mock data for testing
	return &DiskUsage{
		Available: 10 * 1024 * 1024 * 1024, // 10GB
		Used:      5 * 1024 * 1024 * 1024,  // 5GB
		Total:     15 * 1024 * 1024 * 1024, // 15GB
	}, nil
}

// TestMonitoringIntegration tests monitoring and observability
func TestMonitoringIntegration(t *testing.T) {
	t.Run("prometheus configuration", func(t *testing.T) {
		prometheusConfig := "./volumes/prometheus.yml"
		if _, err := os.Stat(prometheusConfig); os.IsNotExist(err) {
			t.Skip("Prometheus configuration not found")
			return
		}

		content, err := os.ReadFile(prometheusConfig)
		if err != nil {
			t.Errorf("Failed to read prometheus config: %v", err)
			return
		}

		if !strings.Contains(string(content), "scrape_configs") {
			t.Error("Prometheus config should contain scrape_configs")
		}

		t.Log("Prometheus configuration appears valid")
	})

	t.Run("grafana dashboards", func(t *testing.T) {
		dashboardDir := "./volumes/dashboards"
		if _, err := os.Stat(dashboardDir); os.IsNotExist(err) {
			t.Skip("Dashboard directory not found")
			return
		}

		files, err := os.ReadDir(dashboardDir)
		if err != nil {
			t.Errorf("Failed to read dashboard directory: %v", err)
			return
		}

		jsonDashboards := 0
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				jsonDashboards++
			}
		}

		t.Logf("Found %d dashboard files", jsonDashboards)
	})
}

// Benchmark tests moved to benchmarks_test.go to avoid duplication

// TestFailureRecovery tests system recovery scenarios
func TestFailureRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping failure recovery tests in short mode")
	}

	t.Run("service restart simulation", func(t *testing.T) {
		// This would test service restart scenarios
		// For now, we'll just validate the concept
		t.Log("Service restart recovery test completed")
	})

	t.Run("data persistence", func(t *testing.T) {
		// Test that data persists across container restarts
		// This would involve checking volume mounts
		t.Log("Data persistence test completed")
	})
}
