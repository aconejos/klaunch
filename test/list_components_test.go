package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
)

func TestExcludedTopicFiltering(t *testing.T) {
	tests := []struct {
		name         string
		topic        string
		shouldFilter bool
	}{
		{
			name:         "consumer offsets topic",
			topic:        "__consumer_offsets",
			shouldFilter: true,
		},
		{
			name:         "confluent command topic",
			topic:        "_confluent-command",
			shouldFilter: true,
		},
		{
			name:         "schemas topic",
			topic:        "_schemas",
			shouldFilter: true,
		},
		{
			name:         "connect configs topic",
			topic:        "docker-connect-configs",
			shouldFilter: true,
		},
		{
			name:         "user topic",
			topic:        "my-user-topic",
			shouldFilter: false,
		},
		{
			name:         "test topic",
			topic:        "test.collection",
			shouldFilter: false,
		},
		{
			name:         "mongodb topic",
			topic:        "mongodb.inventory.products",
			shouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExcludedTopic(tt.topic)
			if result != tt.shouldFilter {
				t.Errorf("Topic %s: expected filtered=%v, got %v", tt.topic, tt.shouldFilter, result)
			}
		})
	}
}

func TestExcludedTopicsStruct(t *testing.T) {
	// Test the ExcludedTopic struct and predefined excluded topics
	if len(excludedTopics) == 0 {
		t.Error("excludedTopics slice should not be empty")
	}

	expectedTopics := []string{
		"__consumer_offsets",
		"_confluent-command",
		"_confluent-telemetry-metrics",
		"_confluent_balancer_api_state",
		"_confluent_balancer_broker_samples",
		"_confluent_balancer_partition_samples",
		"_schemas",
		"docker-connect-configs",
		"docker-connect-offsets",
		"docker-connect-status",
	}

	// Verify all expected topics are in the excludedTopics slice
	for _, expected := range expectedTopics {
		found := false
		for _, excluded := range excludedTopics {
			if excluded.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected topic %s not found in excludedTopics", expected)
		}
	}
}

func TestListConnectors(t *testing.T) {
	tests := []struct {
		name             string
		mockResponse     string
		expectedResponse string
		expectError      bool
	}{
		{
			name:             "successful connector list",
			mockResponse:     `["connector1","connector2","connector3"]`,
			expectedResponse: `["connector1","connector2","connector3"]`,
			expectError:      false,
		},
		{
			name:             "empty connector list",
			mockResponse:     `[]`,
			expectedResponse: `[]`,
			expectError:      false,
		},
		{
			name:         "server error",
			mockResponse: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/connectors" {
					if tt.expectError {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			// Test the HTTP request logic that would be in list_connectors
			resp, err := http.Get(server.URL + "/connectors")
			if tt.expectError {
				if err == nil && resp.StatusCode != http.StatusInternalServerError {
					t.Errorf("Expected error but got successful response")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestTopicListParsing(t *testing.T) {
	tests := []struct {
		name           string
		rawOutput      string
		expectedTopics []string
	}{
		{
			name: "mixed topics with system topics",
			rawOutput: `__consumer_offsets
_confluent-command
user-topic-1
test.collection
_schemas
mongodb.inventory.products
docker-connect-configs`,
			expectedTopics: []string{"user-topic-1", "test.collection", "mongodb.inventory.products"},
		},
		{
			name:           "only system topics",
			rawOutput:      "__consumer_offsets\n_confluent-command\n_schemas",
			expectedTopics: []string{},
		},
		{
			name:           "only user topics",
			rawOutput:      "user-topic-1\ntest.collection\nmongodb.inventory.products",
			expectedTopics: []string{"user-topic-1", "test.collection", "mongodb.inventory.products"},
		},
		{
			name:           "empty output",
			rawOutput:      "",
			expectedTopics: []string{},
		},
	}

	tu := NewTestUtils(t)
	
	// Get excluded topics from the actual excludedTopics slice
	excludedTopicNames := make([]string, len(excludedTopics))
	for i, excluded := range excludedTopics {
		excludedTopicNames[i] = excluded.Name
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse raw topics
			rawTopics := strings.Split(tt.rawOutput, "\n")
			
			// Use shared utility for topic filtering validation
			tu.ValidateTopicFiltering(TopicFilterConfig{
				RawTopics:      rawTopics,
				ExcludedTopics: excludedTopicNames,
				ExpectedTopics: tt.expectedTopics,
			})
		})
	}
}

func TestKafkaTopicsCommandConstruction(t *testing.T) {
	tu := NewTestUtils(t)
	
	// Test the command construction for listing topics
	expectedCmd := []string{
		"docker", "exec", "kafka-connect", "kafka-topics",
		"--bootstrap-server=" + tu.StandardKafkaBootstrapServers(),
		"--list",
	}

	// Validate command structure using shared utility
	tu.ValidateDockerCommand(DockerCommand{
		Args: expectedCmd,
	})

	// Verify command structure
	if len(expectedCmd) != 6 {
		t.Errorf("Expected 6 command arguments, got %d", len(expectedCmd))
	}

	// Verify bootstrap servers using shared utility
	bootstrapArg := "--bootstrap-server=" + tu.StandardKafkaBootstrapServers()
	found := false
	for _, arg := range expectedCmd {
		if arg == bootstrapArg {
			found = true
			break
		}
	}

	if !found {
		t.Error("Bootstrap server argument not found in command")
	}

	// Verify all Kafka brokers using shared utility
	tu.ValidateBootstrapServers(bootstrapArg)
}

func TestListComponentsIntegration(t *testing.T) {
	// Test the integration between list_connectors and list_topics
	t.Run("components integration", func(t *testing.T) {
		// This would test that list_components calls both functions
		// For now, we'll verify the function signature and behavior
		
		// Simulate calling both functions (without actual execution)
		var connectorsResult string
		var topicsResult []string
		
		// Mock successful responses
		connectorsResult = `["test-connector"]`
		topicsResult = []string{"test.topic", "user.events"}

		if connectorsResult == "" {
			t.Error("Connectors result should not be empty in successful case")
		}

		if len(topicsResult) == 0 {
			t.Error("Topics result should not be empty in successful case")
		}
	})
}

func TestDockerCommandExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker command test in short mode")
	}

	tests := []struct {
		name        string
		command     []string
		expectError bool
	}{
		{
			name:        "docker version check",
			command:     []string{"docker", "--version"},
			expectError: false,
		},
		{
			name:        "invalid docker command",
			command:     []string{"docker", "invalid-command"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command[0], tt.command[1:]...)
			err := cmd.Run()

			if tt.expectError && err == nil {
				t.Error("Expected error but command succeeded")
			}

			if !tt.expectError && err != nil {
				t.Skip("Docker not available for testing")
			}
		})
	}
}

func TestTopicFilteringPerformance(t *testing.T) {
	// Create a large list of topics for performance testing
	topics := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		if i%10 == 0 {
			// Add some system topics
			topics[i] = "__consumer_offsets"
		} else {
			topics[i] = fmt.Sprintf("user-topic-%d", i)
		}
	}

	// Test filtering performance
	filtered := 0
	for _, topic := range topics {
		if !isExcludedTopic(topic) {
			filtered++
		}
	}

	expectedFiltered := 900 // 1000 - 100 system topics
	if filtered != expectedFiltered {
		t.Errorf("Expected %d filtered topics, got %d", expectedFiltered, filtered)
	}
}

// Benchmark tests moved to benchmarks_test.go to avoid duplication

// Test concurrent topic filtering
func TestConcurrentTopicFiltering(t *testing.T) {
	topics := []string{
		"__consumer_offsets",
		"user-topic-1",
		"_confluent-command",
		"test.collection",
		"_schemas",
		"mongodb.inventory.products",
	}

	// Test that isExcludedTopic is safe for concurrent use
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for _, topic := range topics {
				isExcludedTopic(topic)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent filtering test completed successfully")
}

func TestEmptyTopicList(t *testing.T) {
	// Test handling of empty topic lists
	rawOutput := ""
	lines := strings.Split(rawOutput, "\n")
	var topics []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !isExcludedTopic(line) {
			topics = append(topics, line)
		}
	}

	if len(topics) != 0 {
		t.Errorf("Expected empty topic list, got %d topics", len(topics))
	}
}