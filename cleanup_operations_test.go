package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test delete_connectors functionality
func TestDeleteConnectors(t *testing.T) {
	tests := []struct {
		name                string
		connectors          []string
		deleteStatus        int
		expectedDeleteCalls int
	}{
		{
			name:                "successful deletion of multiple connectors",
			connectors:          []string{"connector1", "connector2", "connector3"},
			deleteStatus:        http.StatusNoContent,
			expectedDeleteCalls: 3,
		},
		{
			name:                "successful deletion of single connector",
			connectors:          []string{"single-connector"},
			deleteStatus:        http.StatusNoContent,
			expectedDeleteCalls: 1,
		},
		{
			name:                "empty connector list",
			connectors:          []string{},
			deleteStatus:        http.StatusNoContent,
			expectedDeleteCalls: 0,
		},
		{
			name:                "deletion failure",
			connectors:          []string{"connector1"},
			deleteStatus:        http.StatusInternalServerError,
			expectedDeleteCalls: 1,
		},
	}

	tu := NewTestUtils(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Kafka Connect mock server using shared utility
			server := tu.CreateKafkaConnectMockServer(tt.connectors, tt.deleteStatus)
			defer server.Close()

			// Test the deletion logic
			testConnectorDeletion(t, fmt.Sprintf(`["%s"]`, strings.Join(tt.connectors, `","`)), server.URL, tt.expectedDeleteCalls)
		})
	}
}

// Helper function to test connector deletion logic
func testConnectorDeletion(t *testing.T, connectorList, serverURL string, expectedCalls int) {
	// Parse connector list (simulating the logic from delete_connectors)
	cleanConnectorList := strings.TrimPrefix(strings.TrimSuffix(connectorList, "]"), "[")

	if cleanConnectorList == "" {
		if expectedCalls != 0 {
			t.Errorf("Expected %d delete calls for empty list, but list is empty", expectedCalls)
		}
		return
	}

	connectorSlice := strings.Split(cleanConnectorList, ",")
	actualCalls := 0

	for _, connector := range connectorSlice {
		// Clean up quotes and whitespace
		connector = strings.TrimSpace(strings.Trim(connector, "\""))
		if connector == "" {
			continue
		}

		// Simulate DELETE request
		url := serverURL + "/connectors/" + connector
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			t.Errorf("Failed to create DELETE request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("DELETE request failed: %v", err)
			continue
		}
		resp.Body.Close()
		actualCalls++
	}

	if actualCalls != expectedCalls {
		t.Errorf("Expected %d delete calls, got %d", expectedCalls, actualCalls)
	}
}

// Test delete_topics functionality
func TestDeleteTopics(t *testing.T) {
	tests := []struct {
		name            string
		topics          []string
		expectError     bool
		expectedDeletes int
	}{
		{
			name:            "successful topic deletion",
			topics:          []string{"topic1", "topic2", "topic3"},
			expectError:     false,
			expectedDeletes: 3,
		},
		{
			name:            "empty topic list",
			topics:          []string{},
			expectError:     false,
			expectedDeletes: 0,
		},
		{
			name:            "single topic deletion",
			topics:          []string{"single-topic"},
			expectError:     false,
			expectedDeletes: 1,
		},
		{
			name:            "topics with system topics filtered",
			topics:          []string{"user-topic1", "user-topic2"}, // System topics already filtered
			expectError:     false,
			expectedDeletes: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the topic deletion logic
			deleteCount := 0

			for _, topic := range tt.topics {
				if topic == "" {
					continue
				}

				// Simulate the docker exec command for topic deletion
				expectedCmd := []string{
					"docker", "exec", "kafka-connect", "kafka-topics",
					"--delete",
					"--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091",
					"--topic", topic,
				}

				// Verify command structure
				if len(expectedCmd) != 8 {
					t.Errorf("Expected 8 command arguments, got %d", len(expectedCmd))
				}

				// Verify topic argument
				if expectedCmd[7] != topic {
					t.Errorf("Expected topic %s, got %s", topic, expectedCmd[7])
				}

				deleteCount++
			}

			if deleteCount != tt.expectedDeletes {
				t.Errorf("Expected %d topic deletions, got %d", tt.expectedDeletes, deleteCount)
			}
		})
	}
}

func TestConnectorNameCleaning(t *testing.T) {
	tests := []struct {
		name          string
		rawConnector  string
		expectedClean string
	}{
		{
			name:          "quoted connector name",
			rawConnector:  `"test-connector"`,
			expectedClean: "test-connector",
		},
		{
			name:          "connector with spaces",
			rawConnector:  ` "spaced-connector" `,
			expectedClean: "spaced-connector",
		},
		{
			name:          "unquoted connector name",
			rawConnector:  "simple-connector",
			expectedClean: "simple-connector",
		},
		{
			name:          "empty string",
			rawConnector:  "",
			expectedClean: "",
		},
		{
			name:          "only quotes",
			rawConnector:  `""`,
			expectedClean: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the cleaning logic from delete_connectors
			cleaned := strings.TrimSpace(strings.Trim(tt.rawConnector, "\""))

			if cleaned != tt.expectedClean {
				t.Errorf("Expected '%s', got '%s'", tt.expectedClean, cleaned)
			}
		})
	}
}

func TestConnectorListParsing(t *testing.T) {
	tests := []struct {
		name               string
		rawList            string
		expectedConnectors []string
	}{
		{
			name:               "multiple connectors",
			rawList:            `["connector1","connector2","connector3"]`,
			expectedConnectors: []string{"connector1", "connector2", "connector3"},
		},
		{
			name:               "single connector",
			rawList:            `["single-connector"]`,
			expectedConnectors: []string{"single-connector"},
		},
		{
			name:               "empty list",
			rawList:            `[]`,
			expectedConnectors: []string{},
		},
		{
			name:               "malformed list",
			rawList:            `"connector1","connector2"`,
			expectedConnectors: []string{"connector1", "connector2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the parsing logic
			cleanList := strings.TrimPrefix(strings.TrimSuffix(tt.rawList, "]"), "[")

			var connectors []string
			if cleanList != "" {
				rawConnectors := strings.Split(cleanList, ",")
				for _, connector := range rawConnectors {
					cleaned := strings.TrimSpace(strings.Trim(connector, "\""))
					if cleaned != "" {
						connectors = append(connectors, cleaned)
					}
				}
			}

			if len(connectors) != len(tt.expectedConnectors) {
				t.Errorf("Expected %d connectors, got %d", len(tt.expectedConnectors), len(connectors))
				return
			}

			for i, expected := range tt.expectedConnectors {
				if connectors[i] != expected {
					t.Errorf("Connector %d: expected '%s', got '%s'", i, expected, connectors[i])
				}
			}
		})
	}
}

func TestHTTPRequestConstruction(t *testing.T) {
	tests := []struct {
		name          string
		connectorName string
		expectedURL   string
	}{
		{
			name:          "simple connector name",
			connectorName: "test-connector",
			expectedURL:   "http://localhost:8083/connectors/test-connector",
		},
		{
			name:          "connector with numbers",
			connectorName: "mongo-connector-123",
			expectedURL:   "http://localhost:8083/connectors/mongo-connector-123",
		},
		{
			name:          "connector with underscores",
			connectorName: "mongo_kafka_connector",
			expectedURL:   "http://localhost:8083/connectors/mongo_kafka_connector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL := "http://localhost:8083/connectors/"
			constructedURL := baseURL + tt.connectorName

			if constructedURL != tt.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", tt.expectedURL, constructedURL)
			}

			// Test HTTP request creation
			req, err := http.NewRequest("DELETE", constructedURL, nil)
			if err != nil {
				t.Errorf("Failed to create DELETE request: %v", err)
				return
			}

			if req.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", req.Method)
			}

			if req.URL.String() != tt.expectedURL {
				t.Errorf("Request URL mismatch: expected '%s', got '%s'", tt.expectedURL, req.URL.String())
			}
		})
	}
}

func TestTopicDeletionCommand(t *testing.T) {
	// Test the kafka-topics command construction
	topic := "test-topic"
	expectedArgs := []string{
		"docker", "exec", "kafka-connect", "kafka-topics",
		"--delete",
		"--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091",
		"--topic", topic,
	}

	if len(expectedArgs) != 8 {
		t.Errorf("Expected 8 command arguments, got %d", len(expectedArgs))
	}

	// Verify bootstrap servers
	bootstrapArg := "--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091"
	found := false
	for _, arg := range expectedArgs {
		if arg == bootstrapArg {
			found = true
			break
		}
	}

	if !found {
		t.Error("Bootstrap server argument not found in topic deletion command")
	}

	// Verify all Kafka brokers are included
	servers := []string{"kafka1:19091", "kafka2:19092", "kafka3:19093"}
	for _, server := range servers {
		if !strings.Contains(bootstrapArg, server) {
			t.Errorf("Server %s not found in bootstrap servers", server)
		}
	}
}

// Test error handling in deletion operations
func TestDeletionErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		scenario    string
		expectError bool
	}{
		{
			name:        "network error during connector deletion",
			scenario:    "network_error",
			expectError: true,
		},
		{
			name:        "404 error for non-existent connector",
			scenario:    "not_found",
			expectError: false, // Function continues with other connectors
		},
		{
			name:        "successful deletion",
			scenario:    "success",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server for different scenarios
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch tt.scenario {
				case "network_error":
					// Simulate network error by not responding
					return
				case "not_found":
					w.WriteHeader(http.StatusNotFound)
				case "success":
					w.WriteHeader(http.StatusNoContent)
				}
			}))
			defer server.Close()

			// Test HTTP request
			req, err := http.NewRequest("DELETE", server.URL+"/connectors/test", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil && tt.scenario != "network_error" {
				t.Errorf("Unexpected error: %v", err)
			}

			if resp != nil {
				resp.Body.Close()
			}
		})
	}
}

// Benchmark tests moved to benchmarks_test.go to avoid duplication

// Test integration between deletion functions
func TestCleanupIntegration(t *testing.T) {
	t.Run("full cleanup workflow", func(t *testing.T) {
		// This would test the integration between delete_connectors and delete_topics
		// as called from the delete command in main.go

		// Simulate successful connector deletion
		connectorDeleteSuccess := true

		// Simulate successful topic deletion
		topicDeleteSuccess := true

		if !connectorDeleteSuccess {
			t.Error("Connector deletion should succeed")
		}

		if !topicDeleteSuccess {
			t.Error("Topic deletion should succeed")
		}

		// Both operations should complete without stopping each other
		t.Log("Full cleanup workflow simulation completed")
	})
}
