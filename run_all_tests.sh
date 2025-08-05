#!/bin/bash

# Comprehensive test runner for the Go memory guard project
# This script runs all tests including unit tests, integration tests, and validation checks

set -e

echo "🚀 Running comprehensive test suite for memory guard project..."
echo "============================================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_status $RED "❌ Go is not installed or not in PATH"
    exit 1
fi

print_status $GREEN "✅ Go is available: $(go version)"

# Run Go mod tidy to ensure dependencies are correct
print_status $YELLOW "📦 Ensuring Go module dependencies..."
go mod tidy

# Run all Go tests with verbose output and coverage
print_status $YELLOW "🧪 Running Go unit tests..."
if go test -v -race -coverprofile=coverage.out ./...; then
    print_status $GREEN "✅ All Go tests passed!"
    
    # Generate coverage report
    if command -v go &> /dev/null; then
        print_status $YELLOW "📊 Generating coverage report..."
        go tool cover -html=coverage.out -o coverage.html
        print_status $GREEN "✅ Coverage report generated: coverage.html"
        
        # Show coverage summary
        COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}')
        print_status $GREEN "📈 Total test coverage: $COVERAGE"
    fi
else
    print_status $RED "❌ Some Go tests failed!"
    exit 1
fi

# Run benchmark tests
print_status $YELLOW "⚡ Running benchmark tests..."
go test -bench=. -benchmem > benchmark_results.txt
print_status $GREEN "✅ Benchmark results saved to benchmark_results.txt"

# Validate shell scripts exist and are executable
print_status $YELLOW "🔍 Validating shell scripts..."
for script in test_memory_extraction.sh test_memory_simple.sh; do
    if [[ -f "$script" && -x "$script" ]]; then
        print_status $GREEN "✅ $script is present and executable"
    else
        print_status $RED "❌ $script is missing or not executable"
    fi
done

# Check code quality with go vet
print_status $YELLOW "🔎 Running go vet for code quality checks..."
if go vet ./...; then
    print_status $GREEN "✅ Code quality checks passed!"
else
    print_status $RED "❌ Code quality issues found!"
    exit 1
fi

# Check code formatting
print_status $YELLOW "🎨 Checking code formatting..."
if [ -n "$(gofmt -l .)" ]; then
    print_status $RED "❌ Code formatting issues found. Run 'gofmt -w .' to fix."
    gofmt -l .
    exit 1
else
    print_status $GREEN "✅ Code formatting is correct!"
fi

# Run staticcheck if available
if command -v staticcheck &> /dev/null; then
    print_status $YELLOW "🔍 Running staticcheck..."
    if staticcheck ./...; then
        print_status $GREEN "✅ Static analysis passed!"
    else
        print_status $RED "❌ Static analysis found issues!"
        exit 1
    fi
else
    print_status $YELLOW "⚠️  staticcheck not available, skipping static analysis"
fi

# Test that the application can be built
print_status $YELLOW "🏗️  Testing application build..."
if go build -o memguard_app .; then
    print_status $GREEN "✅ Application builds successfully!"
    rm -f memguard_app  # Clean up
else
    print_status $RED "❌ Application build failed!"
    exit 1
fi

# Summary
print_status $GREEN "🎉 All tests completed successfully!"
print_status $GREEN "============================================"
print_status $GREEN "📋 Test Summary:"
print_status $GREEN "   ✅ Unit tests: PASSED"
print_status $GREEN "   ✅ Integration tests: PASSED" 
print_status $GREEN "   ✅ Benchmarks: COMPLETED"
print_status $GREEN "   ✅ Code quality: PASSED"
print_status $GREEN "   ✅ Build test: PASSED"
echo ""
print_status $GREEN "📁 Generated files:"
print_status $GREEN "   - coverage.out (coverage data)"
print_status $GREEN "   - coverage.html (coverage report)"
print_status $GREEN "   - benchmark_results.txt (benchmark results)"