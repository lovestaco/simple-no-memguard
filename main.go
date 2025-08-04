package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// Encryption key - in production, this should be stored securely
var encryptionKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256

// decryptFile decrypts a file using AES-256-GCM
func decryptFile(inputPath, outputPath string) error {
	// Read the encrypted file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %v", err)
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

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %v", err)
	}

	// Write the decrypted data
	return os.WriteFile(outputPath, plaintext, 0644)
}

var loadedPlugin *plugin.Plugin
var pluginOnce sync.Once

// loadPlugin loads and decrypts the plugin (only once)
func loadPlugin() (*plugin.Plugin, error) {
	var loadErr error
	
	pluginOnce.Do(func() {
		// Path to the encrypted plugin
		encryptedPluginPath := "prompt.so.enc"
		
		// Check if encrypted plugin exists
		if _, err := os.Stat(encryptedPluginPath); os.IsNotExist(err) {
			loadErr = fmt.Errorf("encrypted plugin not found: %s. Please run 'make build-plugin' first", encryptedPluginPath)
			return
		}
		
		// Create temporary directory for decrypted plugin
		tempDir, err := os.MkdirTemp("", "plugin")
		if err != nil {
			loadErr = fmt.Errorf("failed to create temp directory: %v", err)
			return
		}
		
		// Decrypt the plugin to temporary location
		decryptedPluginPath := filepath.Join(tempDir, "prompt.so")
		if err := decryptFile(encryptedPluginPath, decryptedPluginPath); err != nil {
			loadErr = fmt.Errorf("failed to decrypt plugin: %v", err)
			return
		}
		
		// Load the plugin
		p, err := plugin.Open(decryptedPluginPath)
		if err != nil {
			loadErr = fmt.Errorf("failed to load plugin: %v (this often happens when plugin was built with different Go version or settings)", err)
			return
		}
		
		loadedPlugin = p
	})
	
	if loadErr != nil {
		return nil, loadErr
	}
	
	return loadedPlugin, nil
}

// getDataFromPlugin gets data from the loaded plugin
func getDataFromPlugin() (string, error) {
	// Load the plugin
	p, err := loadPlugin()
	if err != nil {
		return "", err
	}
	
	// Look up the GetData function
	getDataSym, err := p.Lookup("GetData")
	if err != nil {
		return "", fmt.Errorf("failed to lookup GetData function: %v", err)
	}
	
	// Type assert to function
	getDataFunc, ok := getDataSym.(func() string)
	if !ok {
		return "", fmt.Errorf("GetData is not a function returning string")
	}
	
	// Call the function
	return getDataFunc(), nil
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	
	data, err := getDataFromPlugin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading plugin: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Write([]byte(data))
}

func main() {
	http.HandleFunc("/data", dataHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
