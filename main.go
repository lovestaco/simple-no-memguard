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

	"github.com/awnumar/memguard"
)

// Encryption key - in production, this should be stored securely
var encryptionKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256

// Global protected buffer for sensitive data
var protectedData *memguard.LockedBuffer
var dataOnce sync.Once

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

// secureStringWriter writes data directly from protected memory to HTTP response
type secureStringWriter struct {
	buffer *memguard.LockedBuffer
	writer http.ResponseWriter
}

func (sw *secureStringWriter) Write(data []byte) (int, error) {
	// Write directly from protected memory without creating intermediate strings
	return sw.writer.Write(data)
}

// getDataFromPlugin gets data from the loaded plugin and protects it with memguard
// NOTE: This function is kept for compatibility but not used in the current API
func getDataFromPlugin() (*memguard.LockedBuffer, error) {
	var loadErr error
	
	dataOnce.Do(func() {
		// Load the plugin
		p, err := loadPlugin()
		if err != nil {
			loadErr = err
			return
		}
		
		// Look up the GetData function
		getDataSym, err := p.Lookup("GetData")
		if err != nil {
			loadErr = fmt.Errorf("failed to lookup GetData function: %v", err)
			return
		}
		
		// Type assert to function
		getDataFunc, ok := getDataSym.(func() string)
		if !ok {
			loadErr = fmt.Errorf("GetData is not a function returning string")
			return
		}
		
		// Get the raw data
		rawData := getDataFunc()
		
		// Create a protected buffer for the sensitive data
		protected := memguard.NewBufferFromBytes([]byte(rawData))
		protectedData = protected
	})
	
	if loadErr != nil {
		return nil, loadErr
	}
	
	return protectedData, nil
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	
	// Just verify the encrypted plugin file exists without loading it
	encryptedPluginPath := "prompt.so.enc"
	if _, err := os.Stat(encryptedPluginPath); os.IsNotExist(err) {
		http.Error(w, "Plugin not found", http.StatusInternalServerError)
		return
	}
	
	// Don't load the plugin at all - just return "ok"
	w.Write([]byte("ok"))
}

func main() {
	// Initialize memguard
	memguard.CatchInterrupt()
	defer memguard.Purge()
	
	http.HandleFunc("/data", dataHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
