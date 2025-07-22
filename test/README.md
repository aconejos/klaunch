# Test Suite Organization

This directory contains the comprehensive test suite for the Klaunch project. The tests are organized by functional areas and follow Go testing conventions.

## Test Files

### Core Functionality Tests
- `check_connector_updates_test.go` - Version management and Maven integration tests
- `check_mongodb_running_test.go` - MongoDB connectivity and replica set tests
- `create_kafka_task_test.go` - Task creation and config file selection tests
- `list_components_test.go` - Component monitoring and topic filtering tests
- `list_messages_test.go` - Kafka consumer and message handling tests
- `cleanup_operations_test.go` - Connector/topic deletion and cleanup tests

### Integration Tests
- `integration_test.go` - Full workflow and CLI integration tests
- `infrastructure_test.go` - Docker, networking, and service health tests

### Performance Tests
- `benchmarks_test.go` - Consolidated performance benchmarks for all components

### Test Utilities
- `test_utils.go` - Shared testing utilities and helper functions

## Test Categories

### Unit Tests
Test individual functions and components in isolation with mocking:
```bash
go test -short .
```

### Integration Tests
Test complete workflows and system interactions:
```bash
go test -run TestIntegration .
```

### Infrastructure Tests
Test Docker services, networking, and system health:
```bash
go test -run TestInfrastructure|TestDocker|TestPort|TestNetwork .
```

### Performance Benchmarks
Performance and memory usage benchmarks:
```bash
go test -bench=. .
```

## Test Execution

### Using Test Script
```bash
./run_tests.sh all          # Run all tests
./run_tests.sh unit         # Unit tests only
./run_tests.sh integration  # Integration tests only
./run_tests.sh benchmarks   # Performance benchmarks
```

### Using Makefile
```bash
make test           # Run all tests
make unit-tests     # Unit tests only
make benchmarks     # Performance benchmarks
make coverage       # Generate coverage report
```

### Direct Go Commands
```bash
go test -v .                    # Verbose unit tests
go test -short .                # Quick tests only
go test -race .                 # Race condition detection
go test -coverprofile=c.out .   # With coverage
```

## Test Organization Principles

1. **Shared Utilities**: All common testing patterns consolidated in `test_utils.go`
2. **No Duplication**: HTTP mocking, file creation, and validation utilities shared across tests
3. **Consistent Patterns**: Standardized approach to mocking, setup, and teardown
4. **Performance Focus**: All benchmarks consolidated in single file for easy execution
5. **Clear Separation**: Unit, integration, and infrastructure tests clearly categorized

## Coverage Goals

- **Function Coverage**: All core functions have corresponding test cases
- **Scenario Coverage**: Success paths, error conditions, and edge cases
- **Integration Coverage**: End-to-end workflows and CLI interactions
- **Performance Coverage**: Benchmarks for optimization tracking