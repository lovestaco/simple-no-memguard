package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/awnumar/memguard"
)

// TestGetData tests the getData function
func TestGetData(t *testing.T) {
	t.Run("should return the guarded data buffer", func(t *testing.T) {
		// Setup: Initialize memguard and create test data
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		testData := "test data for memory protection"
		guardedData = memguard.NewBufferFromBytes([]byte(testData))
		defer guardedData.Destroy()
		
		// Execute
		result := getData()
		
		// Assert
		if result == nil {
			t.Fatal("getData() returned nil, expected non-nil LockedBuffer")
		}
		
		if result != guardedData {
			t.Error("getData() did not return the expected guardedData reference")
		}
		
		// Verify data integrity
		if string(result.Bytes()) != testData {
			t.Errorf("getData() returned incorrect data. Expected %q, got %q", testData, string(result.Bytes()))
		}
	})
	
	t.Run("should handle nil guardedData gracefully", func(t *testing.T) {
		// Setup: Set guardedData to nil
		originalData := guardedData
		guardedData = nil
		defer func() { guardedData = originalData }()
		
		// Execute
		result := getData()
		
		// Assert
		if result != nil {
			t.Error("getData() should return nil when guardedData is nil")
		}
	})
}

// TestDataHandler tests the HTTP handler function
func TestDataHandler(t *testing.T) {
	t.Run("should serve guarded data with correct content type", func(t *testing.T) {
		// Setup: Initialize memguard and create test data
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		testData := "OpenAPI specification data for testing"
		guardedData = memguard.NewBufferFromBytes([]byte(testData))
		defer guardedData.Destroy()
		
		// Create request and response recorder
		req := httptest.NewRequest("GET", "/data", nil)
		w := httptest.NewRecorder()
		
		// Execute
		dataHandler(w, req)
		
		// Assert response status
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
		
		// Assert content type
		expectedContentType := "text/plain; charset=utf-8"
		if contentType := resp.Header.Get("Content-Type"); contentType != expectedContentType {
			t.Errorf("Expected Content-Type %q, got %q", expectedContentType, contentType)
		}
		
		// Assert response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		
		if string(body) != testData {
			t.Errorf("Expected response body %q, got %q", testData, string(body))
		}
	})
	
	t.Run("should handle different HTTP methods", func(t *testing.T) {
		// Setup test data
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		testData := "test data"
		guardedData = memguard.NewBufferFromBytes([]byte(testData))
		defer guardedData.Destroy()
		
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
		
		for _, method := range methods {
			t.Run("method_"+method, func(t *testing.T) {
				req := httptest.NewRequest(method, "/data", nil)
				w := httptest.NewRecorder()
				
				dataHandler(w, req)
				
				resp := w.Result()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Method %s: Expected status %d, got %d", method, http.StatusOK, resp.StatusCode)
				}
				
				body, _ := io.ReadAll(resp.Body)
				if string(body) != testData {
					t.Errorf("Method %s: Expected body %q, got %q", method, testData, string(body))
				}
			})
		}
	})
	
	t.Run("should handle concurrent requests safely", func(t *testing.T) {
		// Setup test data
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		testData := "concurrent access test data"
		guardedData = memguard.NewBufferFromBytes([]byte(testData))
		defer guardedData.Destroy()
		
		// Run multiple concurrent requests
		concurrency := 10
		done := make(chan bool, concurrency)
		
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				req := httptest.NewRequest("GET", "/data", nil)
				w := httptest.NewRecorder()
				
				dataHandler(w, req)
				
				resp := w.Result()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Concurrent request %d: Expected status %d, got %d", id, http.StatusOK, resp.StatusCode)
				}
				
				body, _ := io.ReadAll(resp.Body)
				if string(body) != testData {
					t.Errorf("Concurrent request %d: Expected body %q, got %q", id, testData, string(body))
				}
				
				done <- true
			}(i)
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
	
	t.Run("should handle nil guardedData gracefully", func(t *testing.T) {
		// Setup: Set guardedData to nil
		originalData := guardedData
		guardedData = nil
		defer func() { guardedData = originalData }()
		
		req := httptest.NewRequest("GET", "/data", nil)
		w := httptest.NewRecorder()
		
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("dataHandler panicked with nil guardedData: %v", r)
			}
		}()
		
		dataHandler(w, req)
		
		// The handler should handle this gracefully (though it may fail)
		// We're mainly testing that it doesn't crash the server
	})
}

// TestMemoryGuardIntegration tests the integration with memguard
func TestMemoryGuardIntegration(t *testing.T) {
	t.Run("should properly initialize and destroy guarded buffer", func(t *testing.T) {
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		testData := "sensitive data to be guarded"
		buffer := memguard.NewBufferFromBytes([]byte(testData))
		
		// Verify buffer is created correctly
		if buffer == nil {
			t.Fatal("Failed to create LockedBuffer")
		}
		
		// Verify data integrity
		if string(buffer.Bytes()) != testData {
			t.Errorf("Buffer data corrupted. Expected %q, got %q", testData, string(buffer.Bytes()))
		}
		
		// Test that buffer can be destroyed without panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Buffer.Destroy() panicked: %v", r)
			}
		}()
		
		buffer.Destroy()
	})
	
	t.Run("should handle OpenAPI specification data correctly", func(t *testing.T) {
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		// Use the actual OpenAPI data from main.go
		openAPIData := `
 # Role and Context  
You are an expert OpenAPI 3.0.0 specification generator with deep  

# Guidelines  
## General Formatting  

## Endpoint Specifications  

### Endpoint Path  
- The endpoint path must be unique across the entire specification.  
- Use the "parameters" attribute to list these details.  
`
		
		buffer := memguard.NewBufferFromBytes([]byte(openAPIData))
		defer buffer.Destroy()
		
		// Verify the OpenAPI data is preserved correctly
		retrievedData := string(buffer.Bytes())
		if retrievedData != openAPIData {
			t.Error("OpenAPI specification data was not preserved correctly in memory guard")
		}
		
		// Verify specific content exists
		expectedContent := []string{
			"Role and Context",
			"OpenAPI 3.0.0",
			"Guidelines",
			"General Formatting",
			"Endpoint Specifications",
			"parameters",
		}
		
		for _, content := range expectedContent {
			if !strings.Contains(retrievedData, content) {
				t.Errorf("Expected content %q not found in guarded data", content)
			}
		}
	})
}

// TestHTTPServerIntegration tests the full HTTP server integration
func TestHTTPServerIntegration(t *testing.T) {
	t.Run("should serve data endpoint correctly", func(t *testing.T) {
		// Setup test server
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		testData := "integration test data"
		guardedData = memguard.NewBufferFromBytes([]byte(testData))
		defer guardedData.Destroy()
		
		// Create test server
		mux := http.NewServeMux()
		mux.HandleFunc("/data", dataHandler)
		server := httptest.NewServer(mux)
		defer server.Close()
		
		// Make request to test server
		resp, err := http.Get(server.URL + "/data")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		
		// Assert response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		
		if string(body) != testData {
			t.Errorf("Expected response %q, got %q", testData, string(body))
		}
	})
	
	t.Run("should return 404 for unknown endpoints", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/data", dataHandler)
		server := httptest.NewServer(mux)
		defer server.Close()
		
		// Test unknown endpoint
		resp, err := http.Get(server.URL + "/unknown")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status %d for unknown endpoint, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

// BenchmarkDataHandler benchmarks the data handler performance
func BenchmarkDataHandler(b *testing.B) {
	memguard.CatchInterrupt()
	defer memguard.Purge()
	
	testData := strings.Repeat("benchmark test data ", 100)
	guardedData = memguard.NewBufferFromBytes([]byte(testData))
	defer guardedData.Destroy()
	
	req := httptest.NewRequest("GET", "/data", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		dataHandler(w, req)
	}
}

// BenchmarkGetData benchmarks the getData function
func BenchmarkGetData(b *testing.B) {
	memguard.CatchInterrupt()
	defer memguard.Purge()
	
	testData := strings.Repeat("benchmark data ", 1000)
	guardedData = memguard.NewBufferFromBytes([]byte(testData))
	defer guardedData.Destroy()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getData()
	}
}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("should handle empty data", func(t *testing.T) {
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		emptyData := ""
		guardedData = memguard.NewBufferFromBytes([]byte(emptyData))
		defer guardedData.Destroy()
		
		req := httptest.NewRequest("GET", "/data", nil)
		w := httptest.NewRecorder()
		
		dataHandler(w, req)
		
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d for empty data, got %d", http.StatusOK, resp.StatusCode)
		}
		
		body, _ := io.ReadAll(resp.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty response body, got %q", string(body))
		}
	})
	
	t.Run("should handle large data", func(t *testing.T) {
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		// Create large test data (1MB)
		largeData := strings.Repeat("A", 1024*1024)
		guardedData = memguard.NewBufferFromBytes([]byte(largeData))
		defer guardedData.Destroy()
		
		req := httptest.NewRequest("GET", "/data", nil)
		w := httptest.NewRecorder()
		
		dataHandler(w, req)
		
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d for large data, got %d", http.StatusOK, resp.StatusCode)
		}
		
		body, _ := io.ReadAll(resp.Body)
		if len(body) != len(largeData) {
			t.Errorf("Expected response length %d, got %d", len(largeData), len(body))
		}
	})
	
	t.Run("should handle special characters and unicode", func(t *testing.T) {
		memguard.CatchInterrupt()
		defer memguard.Purge()
		
		unicodeData := "Hello 世界 🌍 Special chars: !@#$%^&*()[]{}|;':\",./<>?"
		guardedData = memguard.NewBufferFromBytes([]byte(unicodeData))
		defer guardedData.Destroy()
		
		req := httptest.NewRequest("GET", "/data", nil)
		w := httptest.NewRecorder()
		
		dataHandler(w, req)
		
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)
		
		if string(body) != unicodeData {
			t.Errorf("Unicode data not preserved. Expected %q, got %q", unicodeData, string(body))
		}
	})
}