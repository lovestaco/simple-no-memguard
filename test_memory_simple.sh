#!/bin/bash

# Simple memory extraction test script (no root required)
# This script tests if sensitive data can be extracted from memory using alternative methods

set -e

if [ "$EUID" -ne 0 ]; then
    echo -e "\033[0;31m[ERROR]\033[0m This script must be run as root (sudo)."
    exit 1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="liveapi-runner"
SERVER_URL="http://localhost:8080/data"
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

# Check if binary exists
if [ ! -f "$BINARY_NAME" ]; then
    print_error "Binary $BINARY_NAME not found. Please run 'make build' first."
    exit 1
fi

# Check if required tools are available
check_tools() {
    local missing_tools=()
    for tool in curl strings grep ps awk; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_status "Please install them using: sudo apt-get install curl binutils procps"
        exit 1
    fi
}

print_status "Checking required tools..."
check_tools
print_success "All required tools are available"

# Find the PID of the running liveapi-runner process
print_status "Searching for running $BINARY_NAME process..."
SERVER_PID=$(ps auxf | grep "$BINARY_NAME" | grep -v grep | awk '{print $2}' | head -n 1)

if [ -z "$SERVER_PID" ]; then
    print_error "No running $BINARY_NAME process found. Please start the server first."
    exit 1
fi

print_success "Found $BINARY_NAME process with PID $SERVER_PID"

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

# Method 1: Search in binary file itself
print_status "=== METHOD 1: Searching in binary file ==="
found_any=false
for term in "${SEARCH_TERMS[@]}"; do
    print_status "Searching for: '$term' in binary"
    if strings "$BINARY_NAME" | grep -i "$term" > /dev/null; then
        print_error "FOUND in binary: '$term'"
        found_any=true
        # Show context around the found term
        print_status "Context around '$term' in binary:"
        strings "$BINARY_NAME" | grep -i -A 3 -B 3 "$term" | head -10
        echo
    else
        print_success "NOT found in binary: '$term'"
    fi
done

# Method 2: Search in /proc/$PID/maps and try to read memory regions
print_status "=== METHOD 2: Searching in process memory maps ==="
if [ -r "/proc/$SERVER_PID/maps" ]; then
    print_status "Reading process memory maps..."
    # Try to read readable memory regions
    while IFS= read -r line; do
        if [[ $line =~ rw ]]; then  # Readable and writable regions
            addr_range=$(echo "$line" | awk '{print $1}')
            start_addr=$(echo "$addr_range" | cut -d'-' -f1)
            end_addr=$(echo "$addr_range" | cut -d'-' -f2)
            # Try to read this memory region
            if [ -r "/proc/$SERVER_PID/mem" ]; then
                # Convert hex addresses to decimal for dd
                start_dec=$(printf "%d" "0x$start_addr")
                end_dec=$(printf "%d" "0x$end_addr")
                size=$((end_dec - start_dec))
                if [ $size -gt 0 ] && [ $size -lt 10485760 ]; then  # Max 10MB
                    print_status "Reading memory region: $addr_range (size: $size bytes)"
                    # Create temporary file for this memory region
                    temp_mem="/tmp/mem_region_$$.bin"
                    dd if="/proc/$SERVER_PID/mem" of="$temp_mem" bs=1 skip=$start_dec count=$size 2>/dev/null || continue
                    # Search in this memory region
                    for term in "${SEARCH_TERMS[@]}"; do
                        if strings "$temp_mem" 2>/dev/null | grep -i "$term" > /dev/null; then
                            print_error "FOUND in memory region $addr_range: '$term'"
                            found_any=true
                            # Show context
                            print_status "Context around '$term' in memory:"
                            strings "$temp_mem" 2>/dev/null | grep -i -A 2 -B 2 "$term" | head -10
                            echo
                        fi
                    done
                    rm -f "$temp_mem"
                fi
            fi
        fi
    done < "/proc/$SERVER_PID/maps"
else
    print_warning "Cannot read process memory maps (requires root or ptrace permissions)"
fi

# Method 3: Search in environment and command line
print_status "=== METHOD 3: Searching in process environment ==="
if [ -r "/proc/$SERVER_PID/environ" ]; then
    print_status "Checking process environment..."
    for term in "${SEARCH_TERMS[@]}"; do
        if strings "/proc/$SERVER_PID/environ" 2>/dev/null | grep -i "$term" > /dev/null; then
            print_error "FOUND in environment: '$term'"
            found_any=true
        fi
    done
fi

if [ -r "/proc/$SERVER_PID/cmdline" ]; then
    print_status "Checking process command line..."
    for term in "${SEARCH_TERMS[@]}"; do
        if strings "/proc/$SERVER_PID/cmdline" 2>/dev/null | grep -i "$term" > /dev/null; then
            print_error "FOUND in command line: '$term'"
            found_any=true
        fi
    done
fi

# Summary
echo
print_status "=== SIMPLE MEMORY EXTRACTION TEST SUMMARY ==="
if [ "$found_any" = true ]; then
    print_error "❌ MEMORY EXTRACTION SUCCESSFUL - Sensitive data found!"
    print_error "This means the application is vulnerable to memory extraction attacks."
    print_error "Consider using memory protection techniques like memguard."
else
    print_success "✅ MEMORY EXTRACTION FAILED - No sensitive data found"
    print_success "The application appears to be protected against basic memory extraction."
    print_warning "Note: This is a basic test. Advanced memory forensics might still succeed."
fi

echo
print_status "Test completed."