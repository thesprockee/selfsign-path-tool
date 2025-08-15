package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIHelp(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "test-binary")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("test-binary")

	// Test help flag
	cmd = exec.Command("./test-binary", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}

	helpText := string(output)
	if !strings.Contains(helpText, "--gui") {
		t.Error("Help text should contain --gui flag")
	}
	if !strings.Contains(helpText, "Launch the graphical user interface") {
		t.Error("Help text should describe GUI functionality")
	}
}

func TestGUIFlag(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "test-binary")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("test-binary")

	// Test GUI flag on non-Windows (should fail)
	cmd = exec.Command("./test-binary", "--gui")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("GUI flag should fail on non-Windows platforms")
	}

	outputText := string(output)
	if !strings.Contains(outputText, "GUI mode is only supported on Windows") {
		t.Error("Should show Windows-only error message")
	}
}
