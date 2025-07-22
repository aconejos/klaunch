# Test Organization

All test files are organized in the `/test` directory for better project structure:

## Test Directory Structure

```
test/
├── README.md                          # Test documentation
├── benchmarks_test.go                 # Performance benchmarks
├── check_connector_updates_test.go    # Version management tests
├── check_mongodb_running_test.go      # MongoDB connectivity tests
├── cleanup_operations_test.go         # Cleanup operation tests
├── create_kafka_task_test.go          # Task creation tests
├── infrastructure_test.go             # Infrastructure tests
├── integration_test.go                # Integration tests
├── list_components_test.go            # Component monitoring tests
├── list_messages_test.go              # Message consumer tests
└── test_utils.go                      # Shared test utilities
```

## Running Tests

### Using Test Scripts
```bash
./run_tests.sh all          # Run all tests
./run_tests.sh unit         # Unit tests only
./run_tests.sh benchmarks   # Performance tests
```

### Using Makefile
```bash
make test           # All tests
make unit-tests     # Unit tests only
make benchmarks     # Benchmarks
make coverage       # Coverage report
```

### Direct Go Commands
```bash
go test -v ./test                   # All tests in test directory
go test -short ./test               # Quick tests only
go test -bench=. ./test             # Benchmarks only
go test -coverprofile=c.out ./test  # With coverage
```

## Test Categories

- **Unit Tests**: Individual function testing with mocking
- **Integration Tests**: End-to-end CLI and workflow testing  
- **Infrastructure Tests**: Docker services and system validation
- **Performance Tests**: Benchmarks and memory profiling

All tests use shared utilities from `test_utils.go` for consistent patterns and reduced duplication.