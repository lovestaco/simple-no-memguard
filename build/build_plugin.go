package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Encryption key - must match the one in main.go
var encryptionKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256

// encryptFile encrypts a file using AES-256-GCM
func encryptFile(inputPath, outputPath string) error {
	// Read the input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	// Create cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %v", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to create nonce: %v", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Write the encrypted data
	return os.WriteFile(outputPath, ciphertext, 0644)
}

func main() {
	fmt.Println("Building plugin...")

	// Change to the parent directory to build the plugin
	if err := os.Chdir(".."); err != nil {
		fmt.Printf("Failed to change directory: %v\n", err)
		os.Exit(1)
	}

	// Build the plugin with the same Go version and settings
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", "prompt.so", "prompt.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to build plugin: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Plugin built successfully!")

	// Encrypt the plugin
	fmt.Println("Encrypting plugin...")
	if err := encryptFile("prompt.so", "prompt.so.enc"); err != nil {
		fmt.Printf("Failed to encrypt plugin: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Plugin encrypted successfully!")

	// Clean up the unencrypted plugin
	if err := os.Remove("prompt.so"); err != nil {
		fmt.Printf("Warning: failed to remove unencrypted plugin: %v\n", err)
	}

	fmt.Println("Build complete! Plugin is now encrypted as prompt.so.enc")
} 