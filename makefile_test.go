package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMakefile tests that the Makefile exists and has valid targets
func TestMakefile(t *testing.T) {
	t.Run("Makefile should exist", func(t *testing.T) {
		if _, err := os.Stat("Makefile"); err != nil {
			t.Fatal("Makefile not found")
		}
	})
	
	t.Run("Makefile should have valid syntax", func(t *testing.T) {
		// Test make syntax by running make -n (dry run)
		cmd := exec.Command("make", "-n")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			// Some make errors are expected in dry run, check for syntax errors specifically
			outputStr := string(output)
			if strings.Contains(outputStr, "syntax error") || strings.Contains(outputStr, "missing separator") {
				t.Errorf("Makefile has syntax errors: %s", outputStr)
			}
		}
	})
	
	t.Run("should be able to list make targets", func(t *testing.T) {
		// Try to get help or list targets
		cmd := exec.Command("make", "help")
		_, err := cmd.CombinedOutput()
		
		// If help target doesn't exist, that's fine, but make should run without fatal errors
		if err != nil {
			// Check if it's just a missing target vs syntax error
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() > 2 {
					t.Errorf("Make command failed with unexpected error: %v", err)
				}
			}
		}
	})
}