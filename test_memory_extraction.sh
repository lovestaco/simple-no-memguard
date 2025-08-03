#!/bin/bash

# Memory extraction test script for simple-no-memguard
# This script tests if sensitive data can be extracted from memory dumps

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="liveapi-runner"
SERVER_URL="http://localhost:8080/data"
PID_FILE="/tmp/liveapi-runner.pid"
MEMORY_DUMP_FILE="/tmp/memory_dump.bin"
SEARCH_TERMS=("Role and Context" "Specifications" "OpenAPI" "Guidelines")

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    
    # Kill the server if it's running
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            print_status "Stopping server (PID: $PID)..."
            kill "$PID"
            sleep 2
            if kill -0 "$PID" 2>/dev/null; then
                print_warning "Server still running, force killing..."
                kill -9 "$PID" 2>/dev/null || true
            fi
        fi
        rm -f "$PID_FILE"
    fi
    
    # Remove memory dump file
    if [ -f "$MEMORY_DUMP_FILE" ]; then
        rm -f "$MEMORY_DUMP_FILE"
    fi
    
    print_success "Cleanup completed"
}

# Set up trap to cleanup on exit
trap cleanup EXIT

# Check if binary exists
if [ ! -f "$BINARY_NAME" ]; then
    print_error "Binary $BINARY_NAME not found. Please run 'make build' first."
    exit 1
fi

# Check if required tools are available
check_tools() {
    local missing_tools=()
    
    for tool in curl gdb strings grep; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_status "Please install them using: sudo apt-get install curl gdb binutils"
        exit 1
    fi
}

print_status "Checking required tools..."
check_tools
print_success "All required tools are available"

# Start the server
print_status "Starting server..."
./"$BINARY_NAME" &
SERVER_PID=$!
echo "$SERVER_PID" > "$PID_FILE"

# Wait for server to start
print_status "Waiting for server to start..."
sleep 3

# Check if server is running
if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    print_error "Server failed to start"
    exit 1
fi

# Test if server is responding
print_status "Testing server response..."
if curl -s "$SERVER_URL" > /dev/null; then
    print_success "Server is responding"
else
    print_error "Server is not responding"
    exit 1
fi

# Get the actual response to verify content
print_status "Fetching server response..."
SERVER_RESPONSE=$(curl -s "$SERVER_URL")
if echo "$SERVER_RESPONSE" | grep -q "Role and Context"; then
    print_success "Server contains expected sensitive data"
else
    print_warning "Server response doesn't contain expected sensitive data"
fi

# Take memory dump using gdb
print_status "Taking memory dump using gdb..."
gdb -p "$SERVER_PID" -batch -ex "dump memory $MEMORY_DUMP_FILE 0x0 0x7fffffffffff" 2>/dev/null || {
    print_warning "GDB memory dump failed, trying alternative method..."
    # Alternative: use /proc/$PID/mem if available
    if [ -r "/proc/$SERVER_PID/mem" ]; then
        print_status "Using /proc/$SERVER_PID/mem for memory dump..."
        dd if="/proc/$SERVER_PID/mem" of="$MEMORY_DUMP_FILE" bs=1M count=100 2>/dev/null || true
    else
        print_error "Cannot access process memory"
        exit 1
    fi
}

# Check if memory dump was created
if [ ! -f "$MEMORY_DUMP_FILE" ] || [ ! -s "$MEMORY_DUMP_FILE" ]; then
    print_error "Memory dump failed or is empty"
    exit 1
fi

print_success "Memory dump created: $MEMORY_DUMP_FILE ($(du -h "$MEMORY_DUMP_FILE" | cut -f1))"

# Search for sensitive content in memory dump
print_status "Searching for sensitive content in memory dump..."

found_any=false
for term in "${SEARCH_TERMS[@]}"; do
    print_status "Searching for: '$term'"
    
    # Use strings to extract readable text from binary
    if strings "$MEMORY_DUMP_FILE" | grep -i "$term" > /dev/null; then
        print_error "FOUND sensitive content: '$term'"
        found_any=true
        
        # Show context around the found term
        print_status "Context around '$term':"
        strings "$MEMORY_DUMP_FILE" | grep -i -A 5 -B 5 "$term" | head -20
        echo
    else
        print_success "NOT found: '$term'"
    fi
done

# Additional search using hex dump
print_status "Performing hex dump search..."
for term in "${SEARCH_TERMS[@]}"; do
    # Convert string to hex and search
    hex_term=$(echo -n "$term" | xxd -p)
    if hexdump -C "$MEMORY_DUMP_FILE" | grep -i "$hex_term" > /dev/null; then
        print_error "FOUND in hex dump: '$term'"
        found_any=true
    fi
done

# Summary
echo
print_status "=== MEMORY EXTRACTION TEST SUMMARY ==="
if [ "$found_any" = true ]; then
    print_error "❌ MEMORY EXTRACTION SUCCESSFUL - Sensitive data found in memory dump!"
    print_error "This means the application is vulnerable to memory extraction attacks."
    print_error "Consider using memory protection techniques like memguard."
else
    print_success "✅ MEMORY EXTRACTION FAILED - No sensitive data found in memory dump"
    print_success "The application appears to be protected against memory extraction."
fi

# Show memory dump file info
print_status "Memory dump file: $MEMORY_DUMP_FILE"
print_status "File size: $(du -h "$MEMORY_DUMP_FILE" | cut -f1)"
print_status "To manually inspect: strings $MEMORY_DUMP_FILE | grep -i 'sensitive'"

echo
print_status "Test completed. Server will be stopped automatically." 