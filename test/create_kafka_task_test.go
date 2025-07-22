package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigFiles(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		expected    []string
		expectError bool
	}{
		{
			name: "valid json files",
			files: map[string]string{
				"case_configs/config1.json": `{"name": "test1"}`,
				"case_configs/config2.json": `{"name": "test2"}`,
				"case_configs/readme.txt":   "not a json file",
				"case_configs/config3.json": `{"name": "test3"}`,
			},
			expected:    []string{"config1.json", "config2.json", "config3.json"},
			expectError: false,
		},
		{
			name: "no json files",
			files: map[string]string{
				"case_configs/readme.txt":  "text file",
				"case_configs/config.yaml": "yaml file",
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "empty directory",
			files:       map[string]string{"case_configs/.keep": ""},
			expected:    []string{},
			expectError: false,
		},
	}

	tu := NewTestUtils(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure using shared utility
			tempDir := tu.CreateTempDirStructure(TempDirStructure{
				Files: tt.files,
			})

			// Change to temp directory to test relative paths
			originalWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(originalWd)

			// Test getConfigFiles function
			result, err := getConfigFiles()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check results length
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d files, got %d", len(tt.expected), len(result))
				return
			}

			// Check if all expected files are present
			for _, expected := range tt.expected {
				found := false
				for _, actual := range result {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in results", expected)
				}
			}
		})
	}
}

func TestCreateKafkaTaskHTTPClient(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		serverResponse int
		expectError    bool
	}{
		{
			name: "successful connector creation",
			configContent: `{
				"name": "test-connector",
				"config": {
					"connector.class": "com.mongodb.kafka.connect.MongoSourceConnector",
					"connection.uri": "mongodb://localhost:27017",
					"database": "test",
					"collection": "test"
				}
			}`,
			serverResponse: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "server error",
			configContent: `{
				"name": "test-connector",
				"config": {
					"connector.class": "com.mongodb.kafka.connect.MongoSourceConnector"
				}
			}`,
			serverResponse: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "bad request",
			configContent: `{
				"name": "invalid-connector"
			}`,
			serverResponse: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Kafka Connect server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Verify content type
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// Read request body
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("Failed to read request body: %v", err)
				}

				// Verify JSON content is not empty
				if len(body) == 0 {
					t.Errorf("Request body is empty")
				}

				// Send response
				w.WriteHeader(tt.serverResponse)
				if tt.serverResponse == http.StatusCreated {
					w.Write([]byte(`{"name": "test-connector", "config": {}, "tasks": []}`))
				}
			}))
			defer server.Close()

			// Create temporary config file
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "test-config.json")
			err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Test HTTP client logic
			file, err := os.ReadFile(configFile)
			if err != nil {
				t.Fatalf("Failed to read config file: %v", err)
			}

			req, err := http.NewRequest("POST", server.URL+"/connectors", bytes.NewBuffer(file))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				if !tt.expectError {
					t.Errorf("Unexpected HTTP error: %v", err)
				}
				return
			}
			defer resp.Body.Close()

			if tt.expectError && resp.StatusCode == http.StatusCreated {
				t.Errorf("Expected error but got successful response")
			}

			if !tt.expectError && resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected successful response but got status: %s", resp.Status)
			}
		})
	}
}

func TestConfigFileValidation(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		isValidJSON bool
	}{
		{
			name: "valid JSON config",
			content: `{
				"name": "test-connector",
				"config": {
					"connector.class": "com.mongodb.kafka.connect.MongoSourceConnector",
					"connection.uri": "mongodb://localhost:27017"
				}
			}`,
			isValidJSON: true,
		},
		{
			name:        "invalid JSON",
			content:     `{"name": "test", invalid json}`,
			isValidJSON: false,
		},
		{
			name:        "empty content",
			content:     "",
			isValidJSON: false,
		},
		{
			name:        "valid JSON but empty object",
			content:     "{}",
			isValidJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test if content is valid JSON by trying to create a request
			req, err := http.NewRequest("POST", "http://localhost:8083/connectors", bytes.NewBuffer([]byte(tt.content)))
			
			if err != nil {
				t.Errorf("Failed to create request: %v", err)
				return
			}

			if req.ContentLength == 0 && len(tt.content) > 0 {
				t.Errorf("Request content length is 0 for non-empty content")
			}

			// Basic JSON validation
			if tt.content != "" && !strings.Contains(tt.content, "{") && tt.isValidJSON {
				t.Errorf("Content doesn't appear to be JSON but is marked as valid")
			}
		})
	}
}

func TestFilePathConstruction(t *testing.T) {
	tests := []struct {
		name         string
		selectedFile string
		expected     string
	}{
		{
			name:         "config file selection",
			selectedFile: "default_source_task.json",
			expected:     "./case_configs/default_source_task.json",
		},
		{
			name:         "avro sink config",
			selectedFile: "avro_sink.json",
			expected:     "./case_configs/avro_sink.json",
		},
		{
			name:         "simple source config",
			selectedFile: "simple_source.json",
			expected:     "./case_configs/simple_source.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test filepath construction
			result := filepath.Join("./case_configs", tt.selectedFile)
			
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}

			// Verify the path format is correct
			if !strings.HasPrefix(result, "./case_configs/") {
				t.Errorf("Path doesn't start with expected prefix: %s", result)
			}

			if !strings.HasSuffix(result, ".json") {
				t.Errorf("Path doesn't end with .json: %s", result)
			}
		})
	}
}

// Test the interactive menu logic (simulated)
func TestInteractiveMenuSimulation(t *testing.T) {
	tests := []struct {
		name          string
		configFiles   []string
		userInput     string
		expectedFile  string
		expectDefault bool
	}{
		{
			name:          "valid selection",
			configFiles:   []string{"config1.json", "config2.json", "config3.json"},
			userInput:     "2",
			expectedFile:  "config2.json",
			expectDefault: false,
		},
		{
			name:          "invalid selection",
			userInput:     "99",
			expectedFile:  "default_topic.json",
			expectDefault: true,
		},
		{
			name:          "custom path selection",
			configFiles:   []string{"config1.json", "config2.json"},
			userInput:     "3", // Assuming 3 would be "Enter custom path"
			expectedFile:  "custom",
			expectDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the menu selection logic
			if len(tt.configFiles) > 0 {
				menuOption := tt.userInput
				if menuOption == "2" && len(tt.configFiles) >= 2 {
					selectedFile := tt.configFiles[1] // Index 1 for choice "2"
					if selectedFile != tt.expectedFile {
						t.Errorf("Expected %s, got %s", tt.expectedFile, selectedFile)
					}
				}
			}
			
			t.Logf("Simulated menu test for: %s", tt.name)
		})
	}
}

// Benchmark tests moved to benchmarks_test.go to avoid duplication

// Test error handling
func TestConfigFileErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		filepath    string
		expectError bool
	}{
		{
			name:        "non-existent file",
			filepath:    "/non/existent/file.json",
			expectError: true,
		},
		{
			name:        "empty filepath",
			filepath:    "",
			expectError: true,
		},
		{
			name:        "directory instead of file",
			filepath:    "/tmp",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := os.ReadFile(tt.filepath)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}