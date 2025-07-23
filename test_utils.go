package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestUtils provides shared utilities for testing across all test files
type TestUtils struct {
	t *testing.T
}

// NewTestUtils creates a new test utilities instance
func NewTestUtils(t *testing.T) *TestUtils {
	return &TestUtils{t: t}
}

// HTTPServerConfig defines configuration for mock HTTP servers
type HTTPServerConfig struct {
	Path         string
	Method       string
	ResponseCode int
	ResponseBody string
	Headers      map[string]string
}

// CreateMockHTTPServer creates a mock HTTP server with specified responses
func (tu *TestUtils) CreateMockHTTPServer(configs []HTTPServerConfig) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, config := range configs {
			if r.URL.Path == config.Path && (config.Method == "" || r.Method == config.Method) {
				// Set headers
				for key, value := range config.Headers {
					w.Header().Set(key, value)
				}

				w.WriteHeader(config.ResponseCode)
				if config.ResponseBody != "" {
					w.Write([]byte(config.ResponseBody))
				}
				return
			}
		}

		// Default 404 for unmatched requests
		w.WriteHeader(http.StatusNotFound)
	}))
}

// CreateKafkaConnectMockServer creates a standard Kafka Connect API mock server
func (tu *TestUtils) CreateKafkaConnectMockServer(connectors []string, statusCode int) *httptest.Server {
	connectorsJSON := `["` + strings.Join(connectors, `","`) + `"]`
	if len(connectors) == 0 {
		connectorsJSON = "[]"
	}

	configs := []HTTPServerConfig{
		{
			Path:         "/connectors",
			Method:       "GET",
			ResponseCode: http.StatusOK,
			ResponseBody: connectorsJSON,
			Headers:      map[string]string{"Content-Type": "application/json"},
		},
	}

	// Add DELETE handlers for each connector
	for _, connector := range connectors {
		configs = append(configs, HTTPServerConfig{
			Path:         "/connectors/" + connector,
			Method:       "DELETE",
			ResponseCode: statusCode,
		})
	}

	return tu.CreateMockHTTPServer(configs)
}

// TempDirStructure defines a temporary directory structure for testing
type TempDirStructure struct {
	Dirs  []string
	Files map[string]string // filename -> content
}

// CreateTempDirStructure creates a temporary directory with specified structure
func (tu *TestUtils) CreateTempDirStructure(structure TempDirStructure) string {
	tempDir := tu.t.TempDir()

	// Create directories
	for _, dir := range structure.Dirs {
		dirPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			tu.t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
	}

	// Create files
	for filename, content := range structure.Files {
		filePath := filepath.Join(tempDir, filename)

		// Ensure parent directory exists
		parentDir := filepath.Dir(filePath)
		if parentDir != tempDir {
			err := os.MkdirAll(parentDir, 0755)
			if err != nil {
				tu.t.Fatalf("Failed to create parent directory %s: %v", parentDir, err)
			}
		}

		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			tu.t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	return tempDir
}

// CreateConfigTestEnv creates a standard config test environment
func (tu *TestUtils) CreateConfigTestEnv(configs map[string]string) string {
	structure := TempDirStructure{
		Dirs:  []string{"case_configs", "volumes", "logs"},
		Files: make(map[string]string),
	}

	// Add config files
	for name, content := range configs {
		structure.Files["case_configs/"+name] = content
	}

	// Add standard files
	structure.Files[".env"] = "MONGO_KAFKA_CONNECT_VERSION=1.13.0\n"
	structure.Files["docker/docker-compose.yaml"] = "version: '3'\nservices:\n  test: {}\n"

	return tu.CreateTempDirStructure(structure)
}

// DockerCommand represents a Docker command for testing
type DockerCommand struct {
	Args        []string
	ExpectError bool
	Timeout     time.Duration
}

// ValidateDockerCommand validates Docker command structure without executing
func (tu *TestUtils) ValidateDockerCommand(cmd DockerCommand) {
	if len(cmd.Args) == 0 {
		tu.t.Error("Docker command should not be empty")
		return
	}

	if cmd.Args[0] != "docker" {
		tu.t.Errorf("Expected first argument to be 'docker', got '%s'", cmd.Args[0])
	}

	// Validate common patterns
	if len(cmd.Args) >= 3 && cmd.Args[1] == "exec" {
		if len(cmd.Args) < 4 {
			tu.t.Error("Docker exec command should have at least 4 arguments")
		}
	}
}

// TopicFilterConfig defines configuration for topic filtering tests
type TopicFilterConfig struct {
	RawTopics      []string
	ExcludedTopics []string
	ExpectedTopics []string
}

// ValidateTopicFiltering tests topic filtering logic
func (tu *TestUtils) ValidateTopicFiltering(config TopicFilterConfig) {
	var filteredTopics []string

	for _, topic := range config.RawTopics {
		topic = strings.TrimSpace(topic)
		if len(topic) == 0 {
			continue
		}

		excluded := false
		for _, excludedTopic := range config.ExcludedTopics {
			if topic == excludedTopic {
				excluded = true
				break
			}
		}

		if !excluded {
			filteredTopics = append(filteredTopics, topic)
		}
	}

	if len(filteredTopics) != len(config.ExpectedTopics) {
		tu.t.Errorf("Expected %d topics, got %d", len(config.ExpectedTopics), len(filteredTopics))
		return
	}

	for i, expected := range config.ExpectedTopics {
		if filteredTopics[i] != expected {
			tu.t.Errorf("Topic %d: expected '%s', got '%s'", i, expected, filteredTopics[i])
		}
	}
}

// StringCleaningTest represents a test case for string cleaning functions
type StringCleaningTest struct {
	Name     string
	Input    string
	Expected string
}

// ValidateStringCleaning tests string cleaning operations
func (tu *TestUtils) ValidateStringCleaning(tests []StringCleaningTest, cleanFunc func(string) string) {
	for _, test := range tests {
		tu.t.Run(test.Name, func(t *testing.T) {
			result := cleanFunc(test.Input)
			if result != test.Expected {
				t.Errorf("Expected '%s', got '%s'", test.Expected, result)
			}
		})
	}
}

// PortAvailability checks if a port is available for testing
func (tu *TestUtils) PortAvailability(port string) bool {
	// This is a simplified check - in real testing you might want more robust checking
	return len(port) > 0 && port != "0"
}

// ErrorTestCase represents a test case for error handling
type ErrorTestCase struct {
	Name        string
	Setup       func() error
	ExpectError bool
	ErrorCheck  func(error) bool // Optional: custom error validation
}

// RunErrorTests runs a series of error handling tests
func (tu *TestUtils) RunErrorTests(cases []ErrorTestCase) {
	for _, testCase := range cases {
		tu.t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.Setup()

			if testCase.ExpectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !testCase.ExpectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if testCase.ErrorCheck != nil && err != nil {
				if !testCase.ErrorCheck(err) {
					t.Errorf("Error validation failed for: %v", err)
				}
			}
		})
	}
}

// FileExists checks if a file exists
func (tu *TestUtils) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFileContent reads file content and handles errors
func (tu *TestUtils) ReadFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		tu.t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}

// CreateJSONFile creates a JSON file with specified content
func (tu *TestUtils) CreateJSONFile(dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		tu.t.Fatalf("Failed to create JSON file %s: %v", filePath, err)
	}
	return filePath
}

// StandardKafkaBootstrapServers returns the standard Kafka bootstrap servers used in tests
func (tu *TestUtils) StandardKafkaBootstrapServers() string {
	return "kafka2:19092,kafka3:19093,kafka1:19091"
}

// ValidateBootstrapServers validates that all expected Kafka brokers are present
func (tu *TestUtils) ValidateBootstrapServers(bootstrapServers string) {
	expectedServers := []string{"kafka1:19091", "kafka2:19092", "kafka3:19093"}

	for _, server := range expectedServers {
		if !strings.Contains(bootstrapServers, server) {
			tu.t.Errorf("Bootstrap servers missing expected server: %s", server)
		}
	}
}

// BenchmarkHelper provides utilities for benchmark tests
type BenchmarkHelper struct {
	b *testing.B
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	return &BenchmarkHelper{b: b}
}

// RunWithReset runs a function with timer reset for benchmarks
func (bh *BenchmarkHelper) RunWithReset(fn func()) {
	bh.b.ResetTimer()
	fn()
}

// CleanupHelper provides utilities for test cleanup
type CleanupHelper struct {
	t     *testing.T
	items []func()
}

// NewCleanupHelper creates a new cleanup helper
func NewCleanupHelper(t *testing.T) *CleanupHelper {
	return &CleanupHelper{t: t}
}

// Add adds a cleanup function
func (ch *CleanupHelper) Add(fn func()) {
	ch.items = append(ch.items, fn)
}

// Cleanup runs all cleanup functions
func (ch *CleanupHelper) Cleanup() {
	for _, fn := range ch.items {
		fn()
	}
}
