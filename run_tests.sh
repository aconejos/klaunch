#!/bin/bash

# Klaunch Test Execution Script
# Comprehensive test runner for all Klaunch functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_TIMEOUT="10m"
COVERAGE_DIR="coverage"
TEST_RESULTS_DIR="test-results"
CONTINUE_ON_FAILURE=false
FAILED_TESTS=""

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Function to handle test failures consistently
handle_test_failure() {
    local test_name="$1"
    local error_message="$2"
    
    print_error "$error_message"
    
    if [ "$CONTINUE_ON_FAILURE" = "true" ]; then
        FAILED_TESTS="$FAILED_TESTS $test_name"
        print_warning "Continuing due to --continue flag"
        return 0
    else
        print_error "Stopping execution due to test failure"
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Go version
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+')
    print_status "Found Go version: $GO_VERSION"
    
    # Check Docker (optional)
    if command -v docker &> /dev/null; then
        print_status "Docker is available"
        DOCKER_AVAILABLE=true
    else
        print_warning "Docker not available - skipping Docker-related tests"
        DOCKER_AVAILABLE=false
    fi
    
    # Check Docker Compose (optional)
    if command -v docker-compose &> /dev/null || docker compose version &> /dev/null 2>&1; then
        print_status "Docker Compose is available"
        COMPOSE_AVAILABLE=true
    else
        print_warning "Docker Compose not available - skipping infrastructure tests"
        COMPOSE_AVAILABLE=false
    fi
    
    print_success "Prerequisites check completed"
}

# Function to setup test environment
setup_test_environment() {
    print_status "Setting up test environment..."
    
    # Create test directories
    mkdir -p "$COVERAGE_DIR"
    mkdir -p "$TEST_RESULTS_DIR"
    
    # Clean up previous test artifacts
    rm -f klaunch-test klaunch-integration-test klaunch-benchmark
    rm -f "$COVERAGE_DIR"/*.out "$COVERAGE_DIR"/*.html
    rm -f "$TEST_RESULTS_DIR"/*.xml "$TEST_RESULTS_DIR"/*.json
    
    # Ensure dependencies are up to date
    print_status "Updating Go modules..."
    go mod tidy
    go mod download
    
    print_success "Test environment setup completed"
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    local UNIT_TEST_PACKAGES=(
        "."
    )
    
    # Run tests with coverage
    for package in "${UNIT_TEST_PACKAGES[@]}"; do
        print_status "Testing package: $package"
        
        go test -v -timeout="$TEST_TIMEOUT" \
            -coverprofile="$COVERAGE_DIR/unit_$(basename $package).out" \
            -covermode=atomic \
            -race \
            "$package" | tee "$TEST_RESULTS_DIR/unit_$(basename $package).log"
        
        if [ $? -eq 0 ]; then
            print_success "Unit tests passed for package: $package"
        else
            handle_test_failure "unit-tests" "Unit tests failed for package: $package"
            return 1
        fi
    done
    
    # Generate combined coverage report
    print_status "Generating coverage report..."
    go tool cover -html="$COVERAGE_DIR/unit_test.out" -o "$COVERAGE_DIR/unit_coverage.html" 2>/dev/null || true
    
    print_success "Unit tests completed"
}

# Function to run specific test categories
run_test_category() {
    local category=$1
    local pattern=$2
    
    print_status "Running $category tests..."
    
    go test -v -timeout="$TEST_TIMEOUT" \
        -run "$pattern" \
        -coverprofile="$COVERAGE_DIR/${category}_coverage.out" \
        -covermode=atomic \
        . | tee "$TEST_RESULTS_DIR/${category}_tests.log"
    
    if [ $? -eq 0 ]; then
        print_success "$category tests passed"
    else
        handle_test_failure "$category" "$category tests failed"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Build test binary
    go build -o klaunch-integration-test
    
    if [ $? -ne 0 ]; then
        handle_test_failure "integration" "Failed to build integration test binary"
        return 1
    fi
    
    # Run integration tests
    go test -v -timeout="$TEST_TIMEOUT" \
        -tags=integration \
        -coverprofile="$COVERAGE_DIR/integration_coverage.out" \
        -covermode=atomic \
        . | tee "$TEST_RESULTS_DIR/integration_tests.log"
    
    local exit_code=$?
    
    # Cleanup
    rm -f klaunch-integration-test
    
    if [ $exit_code -eq 0 ]; then
        print_success "Integration tests passed"
    else
        handle_test_failure "integration" "Integration tests failed"
        return 1
    fi
}

# Function to run infrastructure tests
run_infrastructure_tests() {
    if [ "$DOCKER_AVAILABLE" != "true" ]; then
        print_warning "Skipping infrastructure tests - Docker not available"
        return 0
    fi
    
    print_status "Running infrastructure tests..."
    
    go test -v -timeout="$TEST_TIMEOUT" \
        -run "TestInfrastructure|TestDocker|TestPort|TestNetwork" \
        -coverprofile="$COVERAGE_DIR/infrastructure_coverage.out" \
        -covermode=atomic \
        . | tee "$TEST_RESULTS_DIR/infrastructure_tests.log"
    
    if [ $? -eq 0 ]; then
        print_success "Infrastructure tests passed"
    else
        handle_test_failure "infrastructure" "Infrastructure tests failed (services may not be running)"
        return 1
    fi
}

# Function to run performance benchmarks
run_benchmarks() {
    print_status "Running performance benchmarks..."
    
    go test -v -timeout="$TEST_TIMEOUT" \
        -bench=. \
        -benchmem \
        -cpuprofile="$TEST_RESULTS_DIR/cpu.prof" \
        -memprofile="$TEST_RESULTS_DIR/mem.prof" \
        . | tee "$TEST_RESULTS_DIR/benchmarks.log"
    
    if [ $? -eq 0 ]; then
        print_success "Benchmarks completed"
    else
        handle_test_failure "benchmarks" "Some benchmarks failed"
        return 1
    fi
}

# Function to run linting and static analysis
run_static_analysis() {
    print_status "Running static analysis..."
    
    # Format check
    if ! go fmt ./...; then
        handle_test_failure "static-analysis" "Code formatting issues found"
        return 1
    fi
    
    # Vet check
    if ! go vet ./...; then
        handle_test_failure "static-analysis" "Go vet found issues"
        return 1
    fi
    
    # Check for golint if available
    if command -v golint &> /dev/null; then
        print_status "Running golint..."
        golint ./... | tee "$TEST_RESULTS_DIR/lint.log"
    fi
    
    # Check for staticcheck if available
    if command -v staticcheck &> /dev/null; then
        print_status "Running staticcheck..."
        staticcheck ./... | tee "$TEST_RESULTS_DIR/staticcheck.log"
    fi
    
    print_success "Static analysis completed"
}

# Function to generate final test report
generate_test_report() {
    print_status "Generating test report..."
    
    local report_file="$TEST_RESULTS_DIR/test_summary.md"
    
    cat > "$report_file" << EOF
# Klaunch Test Report

**Generated:** $(date)
**Go Version:** $(go version)
**Test Environment:** $(uname -a)

## Test Results

EOF
    
    # Check individual test results
    local total_tests=0
    local passed_tests=0
    
    for log_file in "$TEST_RESULTS_DIR"/*.log; do
        if [ -f "$log_file" ]; then
            local test_name=$(basename "$log_file" .log)
            local test_passed=$(grep -c "PASS" "$log_file" 2>/dev/null || echo "0")
            local test_failed=$(grep -c "FAIL" "$log_file" 2>/dev/null || echo "0")
            
            # Ensure variables are numeric
            test_passed=${test_passed:-0}
            test_failed=${test_failed:-0}
            
            # Additional validation to ensure they're numbers
            [[ "$test_passed" =~ ^[0-9]+$ ]] || test_passed=0
            [[ "$test_failed" =~ ^[0-9]+$ ]] || test_failed=0
            
            echo "### $test_name" >> "$report_file"
            echo "- Passed: $test_passed" >> "$report_file"
            echo "- Failed: $test_failed" >> "$report_file"
            echo "" >> "$report_file"
            
            total_tests=$((total_tests + test_passed + test_failed))
            passed_tests=$((passed_tests + test_passed))
        fi
    done
    
    # Add coverage information
    echo "## Coverage" >> "$report_file"
    for coverage_file in "$COVERAGE_DIR"/*.out; do
        if [ -f "$coverage_file" ]; then
            local coverage_name=$(basename "$coverage_file" .out)
            local coverage_percent=$(go tool cover -func="$coverage_file" 2>/dev/null | tail -1 | awk '{print $3}' || echo "N/A")
            echo "- $coverage_name: $coverage_percent" >> "$report_file"
        fi
    done
    
    echo "" >> "$report_file"
    echo "**Total Tests:** $total_tests" >> "$report_file"
    echo "**Passed:** $passed_tests" >> "$report_file"
    echo "**Failed:** $((total_tests - passed_tests))" >> "$report_file"
    
    print_success "Test report generated: $report_file"
}

# Function to cleanup test environment
cleanup_test_environment() {
    print_status "Cleaning up test environment..."
    
    # Remove test binaries
    rm -f klaunch-test klaunch-integration-test klaunch-benchmark
    
    # Keep results and coverage for analysis
    print_status "Test artifacts preserved in:"
    print_status "  - Coverage: $COVERAGE_DIR/"
    print_status "  - Results: $TEST_RESULTS_DIR/"
    
    print_success "Cleanup completed"
}

# Function to parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --continue)
                CONTINUE_ON_FAILURE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                if [[ -z "$TEST_TYPE" ]]; then
                    TEST_TYPE="$1"
                else
                    print_error "Unknown argument: $1"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    TEST_TYPE="${TEST_TYPE:-all}"
}

# Function to show help
show_help() {
    cat << EOF
Usage: $0 [OPTIONS] [TEST_TYPE]

Test Types:
    unit            Run unit tests only
    integration     Run integration tests only  
    infrastructure  Run infrastructure tests only
    benchmarks      Run performance benchmarks only
    static          Run static analysis only
    all             Run all tests (default)

Options:
    --continue      Continue running tests even if some fail
    -h, --help      Show this help message

Examples:
    $0                          # Run all tests, stop on first failure
    $0 --continue all           # Run all tests, continue on failures
    $0 unit                     # Run only unit tests
    $0 --continue infrastructure # Run infrastructure tests, continue on failure
EOF
}

# Main execution function
main() {
    local start_time=$(date +%s)
    
    # Parse command line arguments
    parse_args "$@"
    
    print_status "Starting Klaunch test suite..."
    print_status "Test type: $TEST_TYPE"
    if [ "$CONTINUE_ON_FAILURE" = "true" ]; then
        print_status "Mode: Continue on failure"
    else
        print_status "Mode: Stop on first failure"
    fi
    
    # Always run prerequisites and setup
    check_prerequisites
    setup_test_environment
    
    # Run tests based on type
    case "$TEST_TYPE" in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "infrastructure")
            run_infrastructure_tests
            ;;
        "benchmarks")
            run_benchmarks
            ;;
        "static")
            run_static_analysis
            ;;
        "all")
            run_static_analysis
            run_unit_tests
            run_integration_tests
            run_infrastructure_tests  
            run_benchmarks
            ;;
        *)
            print_error "Unknown test type: $TEST_TYPE"
            show_help
            exit 1
            ;;
    esac
    
    # Always generate report and cleanup
    generate_test_report
    cleanup_test_environment
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Check if we had failures in continue mode
    if [ "$CONTINUE_ON_FAILURE" = "true" ] && [ -n "$FAILED_TESTS" ]; then
        print_error "Test suite completed with failures in: $FAILED_TESTS"
        print_error "Total duration: ${duration}s"
        exit 1
    else
        print_success "Test suite completed in ${duration}s"
    fi
}

# Script execution
if [ "$0" = "${BASH_SOURCE[0]}" ]; then
    main "$@"
fi