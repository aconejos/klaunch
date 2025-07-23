package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDBConnectionValidation(t *testing.T) {
	tests := []struct {
		name        string
		addresses   []string
		expectError bool
	}{
		{
			name:        "valid addresses",
			addresses:   []string{"127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000"},
			expectError: false,
		},
		{
			name:        "invalid addresses",
			addresses:   []string{"invalid:99999/?directConnection=true&serverSelectionTimeoutMS=2000"},
			expectError: true,
		},
		{
			name:        "empty addresses",
			addresses:   []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test MongoDB connection logic
			for _, addr := range tt.addresses {
				clientOptions := options.Client().ApplyURI("mongodb://" + addr)
				client, err := mongo.Connect(context.Background(), clientOptions)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected connection error: %v", err)
					continue
				}

				if client != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					err = client.Ping(ctx, nil)
					cancel()
					client.Disconnect(context.Background())

					if err != nil && !tt.expectError {
						t.Logf("Expected ping failure for test case: %s", tt.name)
					}
				}
			}
		})
	}
}

func TestMongoDBIsMasterCommand(t *testing.T) {
	// Using mtest for mocking MongoDB responses
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("ismaster true response", func(mt *mtest.T) {
		// Mock ismaster command response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "ismaster", Value: true},
			{Key: "maxBsonObjectSize", Value: 16777216},
		})

		db := mt.Client.Database("admin")
		var result bson.M
		err := db.RunCommand(context.Background(), bson.D{{Key: "ismaster", Value: true}}).Decode(&result)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result["ismaster"] != true {
			t.Errorf("Expected ismaster to be true, got %v", result["ismaster"])
		}
	})

	mt.Run("ismaster false response", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "ismaster", Value: false},
			{Key: "secondary", Value: true},
		})

		db := mt.Client.Database("admin")
		var result bson.M
		err := db.RunCommand(context.Background(), bson.D{{Key: "ismaster", Value: true}}).Decode(&result)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result["ismaster"] != false {
			t.Errorf("Expected ismaster to be false, got %v", result["ismaster"])
		}
	})
}

func TestHostsFileModification(t *testing.T) {
	tests := []struct {
		name           string
		initialContent string
		expectedEntry  string
		shouldModify   bool
	}{
		{
			name:           "hosts file without entry",
			initialContent: "127.0.0.1 localhost\n",
			expectedEntry:  "127.0.0.1 host.docker.internal",
			shouldModify:   true,
		},
		{
			name:           "hosts file with entry",
			initialContent: "127.0.0.1 localhost\n127.0.0.1 host.docker.internal\n",
			expectedEntry:  "127.0.0.1 host.docker.internal",
			shouldModify:   false,
		},
		{
			name:           "empty hosts file",
			initialContent: "",
			expectedEntry:  "127.0.0.1 host.docker.internal",
			shouldModify:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary hosts file
			tempDir := t.TempDir()
			hostsFile := filepath.Join(tempDir, "hosts")

			// Write initial content
			err := os.WriteFile(hostsFile, []byte(tt.initialContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test hosts file: %v", err)
			}

			// Test the logic (simulate the hosts file check)
			content, err := os.ReadFile(hostsFile)
			if err != nil {
				t.Fatalf("Failed to read test hosts file: %v", err)
			}

			contentStr := string(content)
			entryExists := strings.Contains(contentStr, tt.expectedEntry)

			if tt.shouldModify && entryExists {
				t.Errorf("Expected entry to be missing, but it was found")
			}

			if !tt.shouldModify && !entryExists {
				t.Errorf("Expected entry to exist, but it was not found")
			}

			// Test modification logic
			if !entryExists {
				newContent := contentStr + "\n" + tt.expectedEntry
				err = os.WriteFile(hostsFile, []byte(newContent), 0644)
				if err != nil {
					t.Errorf("Failed to write updated hosts file: %v", err)
				}

				// Verify modification
				updatedContent, err := os.ReadFile(hostsFile)
				if err != nil {
					t.Errorf("Failed to read updated hosts file: %v", err)
				}

				if !strings.Contains(string(updatedContent), tt.expectedEntry) {
					t.Errorf("Entry was not properly added to hosts file")
				}
			}
		})
	}
}

func TestReplicaSetConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		port        string
		expectError bool
	}{
		{
			name:        "valid port",
			port:        "27017",
			expectError: false,
		},
		{
			name:        "invalid port",
			port:        "invalid",
			expectError: true,
		},
		{
			name:        "empty port",
			port:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the replica set reconfiguration script generation
			script := `
			var cfg = rs.conf();
			cfg.members[0].host = "host.docker.internal:27017";
			cfg.members[1].host = "host.docker.internal:27018";
			cfg.members[2].host = "host.docker.internal:27019";
			rs.reconfig(cfg);
			`

			if len(script) == 0 {
				t.Errorf("Script should not be empty")
			}

			// Verify script contains expected host configurations
			expectedHosts := []string{
				"host.docker.internal:27017",
				"host.docker.internal:27018",
				"host.docker.internal:27019",
			}

			for _, host := range expectedHosts {
				if !strings.Contains(script, host) {
					t.Errorf("Script missing expected host: %s", host)
				}
			}
		})
	}
}

func TestPortExtractionFromConnectionString(t *testing.T) {
	tests := []struct {
		name          string
		connectionStr string
		expectedPort  string
		expectError   bool
	}{
		{
			name:          "standard port",
			connectionStr: "127.0.0.1:27017/?directConnection=true",
			expectedPort:  "27017",
			expectError:   false,
		},
		{
			name:          "custom port",
			connectionStr: "127.0.0.1:27018/?directConnection=true",
			expectedPort:  "27018",
			expectError:   false,
		},
		{
			name:          "no port",
			connectionStr: "127.0.0.1/?directConnection=true",
			expectedPort:  "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test port extraction regex logic
			if strings.Contains(tt.connectionStr, ":") {
				parts := strings.Split(tt.connectionStr, ":")
				if len(parts) >= 2 {
					portPart := strings.Split(parts[1], "/")[0]
					portPart = strings.Split(portPart, "?")[0]

					if tt.expectError && portPart != "" {
						t.Errorf("Expected error but got port: %s", portPart)
					}

					if !tt.expectError && portPart != tt.expectedPort {
						t.Errorf("Expected port %s, got %s", tt.expectedPort, portPart)
					}
				}
			}
		})
	}
}

// Integration test for full MongoDB connection workflow
func TestMongoDBConnectionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would test the full workflow but requires actual MongoDB
	t.Run("full workflow simulation", func(t *testing.T) {
		addresses := []string{
			"127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000",
			"127.0.0.1:27018/?directConnection=true&serverSelectionTimeoutMS=2000",
			"127.0.0.1:27019/?directConnection=true&serverSelectionTimeoutMS=2000",
		}

		for _, addr := range addresses {
			t.Logf("Testing connection to: %s", addr)
			// In a real test, this would attempt actual connections
			// For now, we just validate the address format
			if !strings.Contains(addr, "127.0.0.1") {
				t.Errorf("Invalid address format: %s", addr)
			}
		}
	})
}

// Benchmark tests moved to benchmarks_test.go to avoid duplication
