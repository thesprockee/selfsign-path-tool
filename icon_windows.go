//go:build windows

package main

import (
	_ "embed"
)

// Embed the icon file data directly into the binary
//go:embed signing_tool.ico
var iconData []byte

// Windows icon resource data (embedded)
const (
	// Icon resource ID
	IDI_MAIN_ICON = 101
)

// GetIconData returns the embedded icon data
func GetIconData() []byte {
	return iconData
}

// GetIconSize returns the size of the embedded icon data
func GetIconSize() int {
	return len(iconData)
}