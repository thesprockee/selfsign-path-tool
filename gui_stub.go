//go:build !windows

package main

import "fmt"

// runGUI is a stub for non-Windows platforms
func runGUI() error {
	return fmt.Errorf("GUI mode is only supported on Windows")
}