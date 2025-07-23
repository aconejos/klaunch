package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BenchmarkDownloadFile benchmarks file download performance
func BenchmarkDownloadFile(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("benchmark test content"))
	}))
	defer server.Close()

	tempDir := b.TempDir()

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			tempFile := filepath.Join(tempDir, fmt.Sprintf("bench%d.jar", i))
			download_file(server.URL, tempFile)
		}
	})
}

// BenchmarkConfigFileReading benchmarks config file reading
func BenchmarkConfigFileReading(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	// Create temporary config file
	tempDir := b.TempDir()
	configFile := filepath.Join(tempDir, "benchmark-config.json")
	content := `{
		"name": "benchmark-connector",
		"config": {
			"connector.class": "com.mongodb.kafka.connect.MongoSourceConnector",
			"connection.uri": "mongodb://localhost:27017",
			"database": "benchmark",
			"collection": "test"
		}
	}`
	os.WriteFile(configFile, []byte(content), 0644)

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			os.ReadFile(configFile)
		}
	})
}

// BenchmarkConnectorDeletion benchmarks connector deletion
func BenchmarkConnectorDeletion(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	connectorNames := []string{"connector1", "connector2", "connector3"}

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			for _, name := range connectorNames {
				req, _ := http.NewRequest("DELETE", server.URL+"/connectors/"+name, nil)
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err == nil && resp != nil {
					resp.Body.Close()
				}
			}
		}
	})
}

// BenchmarkCLIStartup benchmarks CLI startup time
func BenchmarkCLIStartup(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	// Build once
	buildCmd := exec.Command("go", "build", "-o", "klaunch-benchmark")
	err := buildCmd.Run()
	if err != nil {
		b.Fatalf("Failed to build for benchmark: %v", err)
	}
	defer os.Remove("klaunch-benchmark")

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			cmd := exec.Command("./klaunch-benchmark", "--help")
			cmd.Run()
		}
	})
}

// BenchmarkConsumerConfig benchmarks Kafka consumer configuration
func BenchmarkConsumerConfig(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	broker := "host.docker.internal:9093,"
	group := "consumer-cluster-group"

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			config := &kafka.ConfigMap{
				"bootstrap.servers":               broker,
				"group.id":                        group,
				"go.application.rebalance.enable": true,
				"session.timeout.ms":              6000,
				"receive.message.max.bytes":       2147483647,
				"security.protocol":               "PLAINTEXT",
				"api.version.request":             1,
				"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
			}
			_ = config
		}
	})
}

// BenchmarkTopicFiltering benchmarks topic filtering
func BenchmarkTopicFiltering(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	topics := []string{
		"__consumer_offsets",
		"user-topic-1",
		"_confluent-command",
		"test.collection",
		"_schemas",
		"mongodb.inventory.products",
	}

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			for _, topic := range topics {
				isExcludedTopic(topic)
			}
		}
	})
}

// BenchmarkPortCheck benchmarks port availability checking
func BenchmarkPortCheck(b *testing.B) {
	bh := NewBenchmarkHelper(b)
	port := "8083"

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			listener, err := net.Listen("tcp", ":"+port)
			if err == nil {
				listener.Close()
			}
		}
	})
}

// BenchmarkMongoDBConnection benchmarks MongoDB connection performance
func BenchmarkMongoDBConnection(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	bh := NewBenchmarkHelper(b)
	addr := "127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=1000"
	clientOptions := options.Client().ApplyURI("mongodb://" + addr)

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			client, err := mongo.Connect(context.Background(), clientOptions)
			if err != nil {
				b.Skip("MongoDB not available for benchmark")
			}
			if client != nil {
				client.Disconnect(context.Background())
			}
		}
	})
}

// BenchmarkStringCleaning benchmarks connector name cleaning
func BenchmarkStringCleaning(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	testStrings := []string{
		`"test-connector"`,
		` "spaced-connector" `,
		"simple-connector",
		`""`,
	}

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			for _, str := range testStrings {
				// Simulate cleaning logic
				cleaned := strings.TrimSpace(strings.Trim(str, "\""))
				_ = cleaned
			}
		}
	})
}

// BenchmarkJSONValidation benchmarks JSON config validation
func BenchmarkJSONValidation(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	jsonContent := `{
		"name": "test-connector",
		"config": {
			"connector.class": "com.mongodb.kafka.connect.MongoSourceConnector",
			"connection.uri": "mongodb://localhost:27017"
		}
	}`

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			// Simulate JSON validation by checking basic structure
			valid := strings.Contains(jsonContent, "{") && strings.Contains(jsonContent, "}")
			_ = valid
		}
	})
}

// BenchmarkDockerCommandConstruction benchmarks Docker command building
func BenchmarkDockerCommandConstruction(b *testing.B) {
	bh := NewBenchmarkHelper(b)

	bh.RunWithReset(func() {
		for i := 0; i < b.N; i++ {
			// Simulate command construction
			cmd := []string{
				"docker", "exec", "kafka-connect", "kafka-topics",
				"--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091",
				"--list",
			}
			_ = cmd
		}
	})
}
