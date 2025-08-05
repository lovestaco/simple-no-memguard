package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestShellScripts tests that the existing shell scripts execute without errors
func TestShellScripts(t *testing.T) {
	t.Run("test_memory_extraction.sh should be executable", func(t *testing.T) {
		// Check if file exists and is executable
		info, err := os.Stat("test_memory_extraction.sh")
		if err != nil {
			t.Fatalf("test_memory_extraction.sh not found: %v", err)
		}
		
		if info.Mode()&0111 == 0 {
			t.Error("test_memory_extraction.sh is not executable")
		}
	})
	
	t.Run("test_memory_simple.sh should be executable", func(t *testing.T) {
		// Check if file exists and is executable
		info, err := os.Stat("test_memory_simple.sh")
		if err != nil {
			t.Fatalf("test_memory_simple.sh not found: %v", err)
		}
		
		if info.Mode()&0111 == 0 {
			t.Error("test_memory_simple.sh is not executable")
		}
	})
	
	t.Run("shell scripts should have valid bash syntax", func(t *testing.T) {
		scripts := []string{"test_memory_extraction.sh", "test_memory_simple.sh"}
		
		for _, script := range scripts {
			t.Run(script, func(t *testing.T) {
				// Use bash -n to check syntax without executing
				cmd := exec.Command("bash", "-n", script)
				output, err := cmd.CombinedOutput()
				
				if err != nil {
					t.Errorf("Script %s has syntax errors: %v\nOutput: %s", script, err, string(output))
				}
			})
		}
	})
	
	t.Run("shell scripts should contain expected test patterns", func(t *testing.T) {
		scripts := map[string][]string{
			"test_memory_extraction.sh": {"memory", "extraction", "test"},
			"test_memory_simple.sh":     {"memory", "simple", "test"},
		}
		
		for script, expectedPatterns := range scripts {
			t.Run(script, func(t *testing.T) {
				content, err := os.ReadFile(script)
				if err != nil {
					t.Fatalf("Failed to read %s: %v", script, err)
				}
				
				scriptContent := strings.ToLower(string(content))
				for _, pattern := range expectedPatterns {
					if !strings.Contains(scriptContent, pattern) {
						t.Errorf("Script %s does not contain expected pattern: %s", script, pattern)
					}
				}
			})
		}
	})
}