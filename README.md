# Secure Plugin Memory Protection System

## Overview

This project implements a secure system for storing and processing sensitive AI prompts while attempting to prevent memory dump attacks. The system uses encrypted plugins with memguard protection, but faces challenges with memory extraction.

## Architecture Flow

### 1. Source Code

- **`prompt.go`**: Contains sensitive AI prompts as string literals
- **`main.go`**: Main application that loads and processes encrypted plugins

### 2. Build Process

```
prompt.go (source with sensitive prompts)
    ↓ [compile]
prompt.so (shared library with embedded strings)
    ↓ [encrypt with AES-256-GCM]
prompt.so.enc (encrypted file)
```

### 3. Runtime Process

```
prompt.so.enc (encrypted at rest)
    ↓ [decrypt to temp file]
prompt.so (temporary decrypted plugin)
    ↓ [plugin.Open()]
Loaded plugin in memory
    ↓ [getDataFunc()]
Extract prompt string
    ↓ [memguard.NewBufferFromBytes()]
Protected memory buffer
    ↓ [process data]
Count characters, analyze content
```

## What We're Trying to Achieve

### Security Goals

- ✅ **Encrypted at rest**: Prompts stored securely in `prompt.so.enc`
- ✅ **Process data**: Count characters, analyze prompt content
- ✅ **API security**: Return "ok" instead of sensitive data
- ❌ **Memory dump protection**: Prevent extraction via memory dumps

### Use Case

- Store sensitive AI prompts securely
- Process prompts at runtime (character counting, analysis)
- Prevent unauthorized access to prompt content
- Maintain reasonable development workflow

## What We've Tried

### 1. Basic Memguard Protection

```go
// Load plugin and protect with memguard
rawData := getDataFunc()
protected := memguard.NewBufferFromBytes([]byte(rawData))
```

**Result**: ❌ Memory dumps still extract original prompts

### 2. Plugin Loading Prevention

```go
// Don't load plugin at all
if _, err := os.Stat("prompt.so.enc"); err == nil {
    return "ok"
}
```

**Result**: ✅ No memory leakage, but ❌ can't process data

### 3. Selective Data Loading

```go
// Load plugin but don't call GetData()
plugin, _ := plugin.Open(decryptedPath)
// Don't call getDataFunc() to avoid string extraction
```

**Result**: ❌ Plugin loading still puts embedded strings in memory

### 4. Memory Clearing Attempts

```go
// Try to clear original data after memguard protection
for i := range rawBytes {
    rawBytes[i] = 0
}
rawData = ""
```

**Result**: ❌ Original strings still extractable from plugin memory segments

## Root Cause Analysis

The fundamental issue is that when `plugin.Open()` loads the compiled `prompt.so`, it loads all embedded string literals from the original `prompt.go` into regular process memory. Even with memguard protecting a copy, the original data remains in unprotected memory segments where memory dumps can extract it.

## Testing Instructions

### Prerequisites

- Go 1.21+
- Linux system with sudo access
- Memory monitoring tools

### Terminal 1: Build and Run

```bash
make build-and-run
```

This will:

- Compile `prompt.go` → `prompt.so`
- Encrypt `prompt.so` → `prompt.so.enc`
- Start the HTTP server on port 8080

### Terminal 2: Monitor API

```bash
./monitor_api.sh
```

This script continuously hits the `/data` endpoint to trigger plugin loading and data processing.

### Terminal 3: Memory Extraction Test

```bash
sudo ./test_memory_simple.sh
```

This script:

- Attaches to the running process
- Scans memory regions for sensitive strings
- Reports any found prompt content

## Expected Results

### Current State

- ✅ API returns "ok" (no sensitive data in HTTP response)
- ✅ Prompts encrypted at rest
- ❌ Memory dumps can extract original prompts
- ❌ Can't process data without memory leakage

### Memory Dump Output Example

```
[ERROR] FOUND in memory region c000000000-c000800000: 'Role and Context'
[INFO] Context around 'Role and Context' in memory:
 # Role and Context
You are an expert OpenAPI 3.0.0 specification generator with deep
# Guidelines
## General Formatting
## Endpoint Specifications
```

## Security Challenge

**Question**: Is it fundamentally possible to load plugin data for processing while preventing memory dump extraction?

**Current Limitation**: Go plugins with embedded strings will always be extractable via memory dumps, regardless of memguard protection.

## Alternative Approaches to Explore

1. **Dynamic Prompt Generation**: Plugin reads from encrypted files instead of embedded strings
2. **Custom Encryption**: In-memory XOR encoding/decoding
3. **External Process**: Separate process for prompt handling
4. **Memory Obfuscation**: Code generation to obfuscate string storage
5. **Streaming Decryption**: Process data without full memory loading

## Files

- `main.go`: Main application with plugin loading and API
- `prompt.go`: Source file containing sensitive prompts
- `Makefile`: Build and encryption automation
- `monitor_api.sh`: API testing script
- `test_memory_simple.sh`: Memory extraction testing
- `prompt.so.enc`: Encrypted plugin file (generated)

## Dependencies

- `github.com/awnumar/memguard`: Memory protection library
- Standard Go libraries: `crypto/aes`, `crypto/cipher`, `plugin`

## Contributing

This project demonstrates the challenges of securing sensitive data in Go applications. Contributions focusing on:

1. Alternative plugin architectures
2. Better memory protection techniques
3. Secure data processing approaches
4. Memory dump prevention strategies

are welcome!
