# Klaunch Project Analysis

## Project Overview
**Klaunch** is a Go-based CLI tool for MongoDB Kafka Connect testing and reproduction scenarios. It provides a complete Docker-based infrastructure stack for development and testing of MongoDB Kafka connectors.

## Key Classes & Components

### 1. Main CLI Application (`main.go:14`)
- **Framework**: Cobra CLI with 6 main commands
- **Commands**: start, stop, create, delete, show, logs
- **Responsibilities**: Command orchestration, Docker management, user interaction

### 2. Version Manager (`check_connector_updates.go:15`)
- **Functions**: `check_connector_updates()`, `download_file()`
- **Responsibilities**: Maven repository integration, JAR download, semantic versioning
- **Features**: Automatic latest version detection, manual version override

### 3. MongoDB Integration (`check_mogodb_running.go:16`)
- **Function**: `check_mongodb_running()`
- **Responsibilities**: Replica set validation, network configuration, host file management
- **Ports**: 27017, 27018, 27019 (3-node replica set)

### 4. Task Manager (`create_kafka_task.go:10`)
- **Function**: `create_kafka_task()`
- **Responsibilities**: Interactive connector creation, JSON validation, REST API integration
- **Default Config**: `./case_configs/default_topic.json`

### 5. Component Monitor (`list_components.go:27`)
- **Types**: `ExcludedTopic` struct
- **Functions**: `list_components()`, `list_connectors()`, `list_topics()`
- **Features**: System topic filtering, connector enumeration

### 6. Message Consumer (`list_messages.go:11`)
- **Function**: `list_messages()`
- **Dependencies**: Confluent Kafka Go client
- **Features**: Real-time consumption, signal handling, offset management

### 7. Cleanup Utilities
- **Files**: `delete_connectors.go:10`, `delete_topics.go:8`
- **Functions**: `delete_connectors()`, `delete_topics()`
- **Features**: Batch operations, error handling

## Infrastructure Architecture

### Docker Services (docker-compose.yaml)
- **Kafka Cluster**: 3-node setup (kafka1:9091, kafka2:9092, kafka3:9093)
- **Zookeeper**: Single node coordination (port 2181)
- **Kafka Connect**: REST API (port 8083) with MongoDB connector
- **Schema Registry**: Port 8081
- **CMAK**: Web UI (port 9000)

### Configuration Management
- **Config Directory**: `case_configs/` with pre-built templates
- **Environment**: `.env` file with `MONGO_KAFKA_CONNECT_VERSION`
- **Network**: `host.docker.internal` mapping for MongoDB access

## Key Features

### Infrastructure Management
- Automated Docker orchestration
- Version management (MongoDB Kafka Connect)
- Service health checks
- Network configuration

### Connector Operations
- Dynamic connector creation with JSON configs
- Lifecycle management (create, list, delete)
- Topic management with intelligent filtering
- Real-time message monitoring

### Development Tools
- Configuration templates
- Centralized logging
- Interactive CLI
- Web interfaces (CMAK)

## Dependencies (go.mod:5)
- `github.com/spf13/cobra`: CLI framework
- `github.com/confluentinc/confluent-kafka-go/v2`: Kafka client
- `go.mongodb.org/mongo-driver`: MongoDB driver
- `golang.org/x/mod`: Semantic versioning

## Platform Support
- **Primary**: macOS (fully tested)
- **Secondary**: Linux (Ubuntu 24.04+)
- **Containers**: Cross-platform via Docker

## Typical Workflow
1. `./klaunch start [version]` - Setup infrastructure
2. `./klaunch create` - Create connector with config
3. `./klaunch show messages` - Monitor data flow
4. `./klaunch logs` - Extract debug logs
5. `./klaunch delete` - Cleanup connectors/topics
6. `./klaunch stop` - Teardown environment

## Web Access Points
- CMAK: http://localhost:9000 (cluster: kafka-connect, zk: zookeeper1:2181)
- Kafka Connect API: http://localhost:8083
- Schema Registry: http://localhost:8081

## Comprehensive Test Suite

### Test Categories Created:
1. **Unit Tests** - Individual function testing for all core components
2. **Integration Tests** - End-to-end CLI command testing and workflow validation
3. **Infrastructure Tests** - Docker services, networking, and system validation
4. **Performance Tests** - Benchmarks and memory usage analysis

### Test Files Created:
- `check_connector_updates_test.go` - Version management and Maven integration tests
- `check_mongodb_running_test.go` - MongoDB connectivity and replica set tests
- `create_kafka_task_test.go` - Task creation and config file selection tests
- `list_components_test.go` - Component monitoring and topic filtering tests  
- `list_messages_test.go` - Kafka consumer and message handling tests
- `cleanup_operations_test.go` - Connector/topic deletion and cleanup tests
- `integration_test.go` - Full workflow and CLI integration tests
- `infrastructure_test.go` - Docker, networking, and service health tests

### Test Execution Tools:
- `run_tests.sh` - Comprehensive test runner with categories and reporting
- `Makefile` - Build automation with test targets and development workflows

### Test Coverage:
- **Function Coverage**: All 7 core Go files have corresponding test files
- **Scenario Coverage**: Success/failure paths, error handling, edge cases
- **Integration Coverage**: CLI commands, Docker services, network connectivity
- **Performance Coverage**: Benchmarks, memory usage, concurrent operations

### Usage Examples:
```bash
# Run all tests
./run_tests.sh all

# Run specific categories
./run_tests.sh unit
./run_tests.sh integration
./run_tests.sh infrastructure

# Using Makefile
make test           # All tests
make unit-tests     # Unit tests only
make benchmarks     # Performance tests
make coverage       # Generate coverage report
```

### Test Results:
- All unit tests passing with proper mocking and validation
- Integration tests cover CLI functionality and workflows
- Infrastructure tests validate Docker setup and service health
- Performance benchmarks established for optimization tracking

## Test Organization Structure

### Final Test Directory Layout:
All test files organized in dedicated `/test` directory:

```
test/
├── README.md                          # Comprehensive test documentation
├── benchmarks_test.go                 # All performance benchmarks (8 functions)
├── check_connector_updates_test.go    # Version management tests
├── check_mongodb_running_test.go      # MongoDB connectivity tests
├── cleanup_operations_test.go         # Cleanup operation tests
├── create_kafka_task_test.go          # Task creation tests  
├── infrastructure_test.go             # Infrastructure validation tests
├── integration_test.go                # End-to-end integration tests
├── list_components_test.go            # Component monitoring tests
├── list_messages_test.go              # Message consumer tests
└── test_utils.go                      # Shared utilities (200+ lines)
```

### Test Execution Methods:
1. **Test Scripts**: `./run_tests.sh [category]` - Comprehensive test runner
2. **Makefile**: `make test`, `make unit-tests`, `make benchmarks`  
3. **Direct Go**: `go test -v ./test` - Standard Go testing

### Organization Benefits:
- **Clear Separation**: Tests isolated from main source code
- **Easy Navigation**: All test-related files in single directory
- **Consistent Structure**: Following Go project organization best practices
- **Comprehensive Documentation**: README.md with usage examples and patterns
- **Maintained Functionality**: All tests work with `./test` directory structure
