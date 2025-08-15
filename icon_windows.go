//go:build windows

package main

// Windows icon resource data (embedded)
// This would normally be a compiled resource file (.ico)
// For demonstration, we're just specifying the icon ID

const (
	// Icon resource ID
	IDI_MAIN_ICON = 101
)

// Icon embedding would be done through resource compilation
// For a full implementation, you would:
// 1. Create an .ico file
// 2. Create a .rc resource file
// 3. Compile with resource compiler (rc.exe)
// 4. Link with Go using -ldflags="-H windowsgui -extldflags=-static"
//
// Example .rc file content:
// IDI_MAIN_ICON ICON "signing_tool.ico"
//
// For now, we'll use the default application icon