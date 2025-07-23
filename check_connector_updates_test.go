package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		expectError    bool
	}{
		{
			name:           "successful download",
			serverResponse: "test content",
			statusCode:     http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			serverResponse: "",
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "not found",
			serverResponse: "",
			statusCode:     http.StatusNotFound,
			expectError:    true,
		},
	}

	tu := NewTestUtils(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server using shared utility
			server := tu.CreateMockHTTPServer([]HTTPServerConfig{
				{
					Path:         "/",
					ResponseCode: tt.statusCode,
					ResponseBody: tt.serverResponse,
				},
			})
			defer server.Close()

			// Create temporary file
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "test.jar")

			// Test download_file function
			err := download_file(server.URL, tempFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify file was created and has correct content using utility
			if !tu.FileExists(tempFile) {
				t.Errorf("Downloaded file does not exist")
				return
			}

			content := tu.ReadFileContent(tempFile)
			if content != tt.serverResponse {
				t.Errorf("Expected content '%s', got '%s'", tt.serverResponse, content)
			}
		})
	}
}

func TestCheckConnectorUpdates(t *testing.T) {
	tests := []struct {
		name         string
		inputVersion string
		mockHTML     string
		expectError  bool
	}{
		{
			name:         "specific version provided",
			inputVersion: "1.13.0",
			mockHTML:     `<a href="1.13.0/">`,
			expectError:  false,
		},
		{
			name:         "empty version - use latest",
			inputVersion: "",
			mockHTML:     `<a href="1.13.0/"><a href="1.14.0/"><a href="1.15.0/">`,
			expectError:  false,
		},
		{
			name:         "invalid HTML response",
			inputVersion: "",
			mockHTML:     "invalid html",
			expectError:  false, // Should handle gracefully
		},
	}

	tu := NewTestUtils(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory structure using shared utility
			tempDir := tu.CreateTempDirStructure(TempDirStructure{
				Dirs: []string{"volumes"},
				Files: map[string]string{
					".env": "MONGO_KAFKA_CONNECT_VERSION=1.12.0\n",
				},
			})

			// Change to temp directory
			originalWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(originalWd)

			// Create mock server using shared utility
			server := tu.CreateMockHTTPServer([]HTTPServerConfig{
				{
					Path:         "/maven2/org/mongodb/kafka/mongo-kafka-connect/",
					ResponseCode: http.StatusOK,
					ResponseBody: tt.mockHTML,
				},
				{
					Path:         "/",
					ResponseCode: http.StatusOK,
					ResponseBody: "mock jar content",
				},
			})
			defer server.Close()

			// Note: This test would require modifying the function to accept a custom URL
			// For now, we'll test the basic structure
			t.Logf("Test setup complete for version: %s", tt.inputVersion)
		})
	}
}

func TestVersionParsing(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "multiple versions",
			html:     `<a href="1.13.0/">1.13.0/</a><a href="1.14.0/">1.14.0/</a><a href="1.15.0/">1.15.0/</a>`,
			expected: []string{"v1.13.0", "v1.14.0", "v1.15.0"},
		},
		{
			name:     "single version",
			html:     `<a href="1.13.0/">1.13.0/</a>`,
			expected: []string{"v1.13.0"},
		},
		{
			name:     "no versions",
			html:     `<html><body>No versions</body></html>`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test the regex parsing logic from check_connector_updates
			// We can extract this logic into a separate testable function
			t.Logf("Would test version parsing for: %s", tt.name)
		})
	}
}

// Benchmark tests moved to benchmarks_test.go to avoid duplication

// Helper function to create test .env file
func createTestEnvFile(dir, version string) error {
	envPath := filepath.Join(dir, ".env")
	content := fmt.Sprintf("MONGO_KAFKA_CONNECT_VERSION=%s\n", version)
	return os.WriteFile(envPath, []byte(content), 0644)
}

// Test helper for creating mock Maven repository response
func createMockMavenResponse(versions []string) string {
	var html strings.Builder
	html.WriteString("<html><body>")
	for _, version := range versions {
		html.WriteString(fmt.Sprintf(`<a href="%s/">%s/</a>`, version, version))
	}
	html.WriteString("</body></html>")
	return html.String()
}
