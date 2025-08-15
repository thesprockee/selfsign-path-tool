//go:build windows

package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// createWelcomeScreen creates the welcome/introduction screen
func (app *GuiApp) createWelcomeScreen() {
	app.clearAllControls()
	
	hInstance, _, _ := procGetModuleHandle.Call(0)
	
	// Title
	titleHwnd := createWindow("STATIC", "File Signing Tool", 
		WS_VISIBLE|WS_CHILD, 50, 30, 500, 40, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["title"] = titleHwnd
	
	// Welcome text
	welcomeText := `Welcome to the File Signing Tool!

This wizard will guide you through the process of signing your executable files with a self-signed certificate.

The signing process includes:
• Selecting files to sign
• Creating a secure code signing certificate
• Signing your files
• Installing the certificate to the Windows certificate store
• Securely removing temporary keys

This helps Windows Defender and other antivirus software recognize your files as trusted, reducing false positive detections.

Click Next to begin selecting files to sign.`
	
	textHwnd := createWindow("STATIC", welcomeText,
		WS_VISIBLE|WS_CHILD, 50, 80, 500, 280, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["welcome_text"] = textHwnd
	
	// Next button
	nextHwnd := createWindow("BUTTON", "Next >",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_DEFPUSHBUTTON, 
		420, 400, 80, 30, app.hwnd, ID_BUTTON_NEXT, syscall.Handle(hInstance))
	app.controls["next"] = nextHwnd
	
	// Cancel button
	cancelHwnd := createWindow("BUTTON", "Cancel",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON,
		330, 400, 80, 30, app.hwnd, ID_BUTTON_CANCEL, syscall.Handle(hInstance))
	app.controls["cancel"] = cancelHwnd
}

// createFileSelectionScreen creates the file selection screen
func (app *GuiApp) createFileSelectionScreen() {
	app.clearAllControls()
	
	hInstance, _, _ := procGetModuleHandle.Call(0)
	
	// Title
	titleHwnd := createWindow("STATIC", "Select Files to Sign",
		WS_VISIBLE|WS_CHILD, 50, 30, 500, 30, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["title"] = titleHwnd
	
	// Instructions
	instructText := "Choose the executable files you want to sign. You can select multiple files."
	instructHwnd := createWindow("STATIC", instructText,
		WS_VISIBLE|WS_CHILD, 50, 70, 500, 30, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["instructions"] = instructHwnd
	
	// Browse button
	browseHwnd := createWindow("BUTTON", "Browse for Files...",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON,
		50, 110, 120, 30, app.hwnd, ID_BUTTON_BROWSE, syscall.Handle(hInstance))
	app.controls["browse"] = browseHwnd
	
	// File list
	listHwnd := createWindow("LISTBOX", "",
		WS_VISIBLE|WS_CHILD|WS_BORDER|LBS_STANDARD,
		50, 150, 500, 200, app.hwnd, ID_LISTBOX_FILES, syscall.Handle(hInstance))
	app.controls["file_list"] = listHwnd
	
	// Populate existing files if any
	app.updateFileList()
	
	// Back button
	backHwnd := createWindow("BUTTON", "< Back",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON,
		240, 400, 80, 30, app.hwnd, ID_BUTTON_BACK, syscall.Handle(hInstance))
	app.controls["back"] = backHwnd
	
	// Next button
	nextHwnd := createWindow("BUTTON", "Next >",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_DEFPUSHBUTTON,
		420, 400, 80, 30, app.hwnd, ID_BUTTON_NEXT, syscall.Handle(hInstance))
	app.controls["next"] = nextHwnd
	
	// Cancel button
	cancelHwnd := createWindow("BUTTON", "Cancel",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON,
		330, 400, 80, 30, app.hwnd, ID_BUTTON_CANCEL, syscall.Handle(hInstance))
	app.controls["cancel"] = cancelHwnd
}

// createConfirmScreen creates the confirmation screen
func (app *GuiApp) createConfirmScreen() {
	app.clearAllControls()
	
	hInstance, _, _ := procGetModuleHandle.Call(0)
	
	// Title
	titleHwnd := createWindow("STATIC", "Confirm File Signing",
		WS_VISIBLE|WS_CHILD, 50, 30, 500, 30, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["title"] = titleHwnd
	
	// Summary text
	fileCount := len(app.selectedFiles)
	summaryText := fmt.Sprintf("Ready to sign %d file(s):\n\n", fileCount)
	for i, file := range app.selectedFiles {
		if i < 10 { // Show first 10 files
			summaryText += fmt.Sprintf("• %s\n", filepath.Base(file))
		} else {
			summaryText += fmt.Sprintf("• ... and %d more files\n", fileCount-10)
			break
		}
	}
	summaryText += "\nThe signing process will:\n"
	summaryText += "• Create a new self-signed certificate\n"
	summaryText += "• Sign all selected files\n"
	summaryText += "• Install the certificate to Windows certificate store\n"
	summaryText += "• Securely delete temporary keys\n\n"
	summaryText += "Click Next to begin signing."
	
	textHwnd := createWindow("STATIC", summaryText,
		WS_VISIBLE|WS_CHILD, 50, 70, 500, 280, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["summary"] = textHwnd
	
	// Back button
	backHwnd := createWindow("BUTTON", "< Back",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON,
		240, 400, 80, 30, app.hwnd, ID_BUTTON_BACK, syscall.Handle(hInstance))
	app.controls["back"] = backHwnd
	
	// Next button (Sign button)
	nextHwnd := createWindow("BUTTON", "Sign Files",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_DEFPUSHBUTTON,
		420, 400, 80, 30, app.hwnd, ID_BUTTON_NEXT, syscall.Handle(hInstance))
	app.controls["next"] = nextHwnd
	
	// Cancel button
	cancelHwnd := createWindow("BUTTON", "Cancel",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON,
		330, 400, 80, 30, app.hwnd, ID_BUTTON_CANCEL, syscall.Handle(hInstance))
	app.controls["cancel"] = cancelHwnd
}

// createProcessingScreen creates the processing screen
func (app *GuiApp) createProcessingScreen() {
	app.clearAllControls()
	
	hInstance, _, _ := procGetModuleHandle.Call(0)
	
	// Title
	titleHwnd := createWindow("STATIC", "Signing Files...",
		WS_VISIBLE|WS_CHILD, 50, 30, 500, 30, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["title"] = titleHwnd
	
	// Status text
	statusHwnd := createWindow("STATIC", "Please wait while files are being signed...",
		WS_VISIBLE|WS_CHILD, 50, 70, 500, 30, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["status"] = statusHwnd
	
	// Output area
	outputHwnd := createWindow("EDIT", "",
		WS_VISIBLE|WS_CHILD|WS_BORDER|ES_MULTILINE|ES_READONLY|ES_AUTOVSCROLL,
		50, 110, 500, 240, app.hwnd, ID_EDIT_OUTPUT, syscall.Handle(hInstance))
	app.controls["output"] = outputHwnd
	
	// Cancel button (disabled during processing)
	cancelHwnd := createWindow("BUTTON", "Cancel",
		WS_VISIBLE|WS_CHILD|BS_PUSHBUTTON,
		330, 400, 80, 30, app.hwnd, ID_BUTTON_CANCEL, syscall.Handle(hInstance))
	app.controls["cancel"] = cancelHwnd
	// Disable cancel button during processing
	user32.NewProc("EnableWindow").Call(uintptr(cancelHwnd), 0)
	
	// Start the signing process
	go app.performSigning()
}

// createCompleteScreen creates the completion screen
func (app *GuiApp) createCompleteScreen(success bool, results string) {
	app.clearAllControls()
	
	hInstance, _, _ := procGetModuleHandle.Call(0)
	
	// Title
	var title string
	if success {
		title = "Signing Complete!"
	} else {
		title = "Signing Failed"
	}
	
	titleHwnd := createWindow("STATIC", title,
		WS_VISIBLE|WS_CHILD, 50, 30, 500, 30, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["title"] = titleHwnd
	
	// Results text
	resultsHwnd := createWindow("EDIT", results,
		WS_VISIBLE|WS_CHILD|WS_BORDER|ES_MULTILINE|ES_READONLY|ES_AUTOVSCROLL,
		50, 70, 500, 280, app.hwnd, 0, syscall.Handle(hInstance))
	app.controls["results"] = resultsHwnd
	
	// Finish button
	finishHwnd := createWindow("BUTTON", "Finish",
		WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_DEFPUSHBUTTON,
		420, 400, 80, 30, app.hwnd, ID_BUTTON_CANCEL, syscall.Handle(hInstance))
	app.controls["finish"] = finishHwnd
}

// nextStep moves to the next wizard step
func (app *GuiApp) nextStep() {
	switch app.currentStep {
	case STEP_WELCOME:
		app.currentStep = STEP_FILE_SELECTION
		app.createFileSelectionScreen()
	case STEP_FILE_SELECTION:
		if len(app.selectedFiles) == 0 {
			app.showMessage("Please select at least one file to sign.", "No Files Selected")
			return
		}
		app.currentStep = STEP_CONFIRM
		app.createConfirmScreen()
	case STEP_CONFIRM:
		app.currentStep = STEP_PROCESSING
		app.createProcessingScreen()
	}
}

// previousStep moves to the previous wizard step
func (app *GuiApp) previousStep() {
	switch app.currentStep {
	case STEP_FILE_SELECTION:
		app.currentStep = STEP_WELCOME
		app.createWelcomeScreen()
	case STEP_CONFIRM:
		app.currentStep = STEP_FILE_SELECTION
		app.createFileSelectionScreen()
	}
}

// browseFiles opens file selection dialog
func (app *GuiApp) browseFiles() {
	// Prepare file buffer - large enough for multiple files
	fileBuffer := make([]uint16, 32768)
	
	// File filter for executable files
	filter := "Executable Files\x00*.exe;*.dll;*.msi;*.sys;*.com;*.ocx;*.scr;*.cpl\x00All Files\x00*.*\x00\x00"
	filterPtr := syscall.StringToUTF16Ptr(filter)
	
	ofn := OPENFILENAME{
		LStructSize:  uint32(unsafe.Sizeof(OPENFILENAME{})),
		HwndOwner:    app.hwnd,
		LpstrFilter:  filterPtr,
		LpstrFile:    &fileBuffer[0],
		NMaxFile:     uint32(len(fileBuffer)),
		LpstrTitle:   syscall.StringToUTF16Ptr("Select Files to Sign"),
		Flags:        OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_ALLOWMULTISELECT | OFN_EXPLORER,
	}
	
	ret, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret != 0 {
		// Parse selected files
		files := app.parseMultiSelectFiles(fileBuffer)
		
		// Add to selected files (avoid duplicates)
		for _, file := range files {
			exists := false
			for _, existing := range app.selectedFiles {
				if strings.EqualFold(existing, file) {
					exists = true
					break
				}
			}
			if !exists {
				app.selectedFiles = append(app.selectedFiles, file)
			}
		}
		
		app.updateFileList()
	}
}

// parseMultiSelectFiles parses the multi-select file dialog result
func (app *GuiApp) parseMultiSelectFiles(buffer []uint16) []string {
	var files []string
	
	// Convert to string
	str := syscall.UTF16ToString(buffer)
	if str == "" {
		return files
	}
	
	// Find null separators
	parts := strings.Split(str, "\x00")
	if len(parts) <= 1 {
		// Single file selected
		files = append(files, str)
	} else {
		// Multiple files selected
		directory := parts[0]
		for i := 1; i < len(parts) && parts[i] != ""; i++ {
			fullPath := filepath.Join(directory, parts[i])
			files = append(files, fullPath)
		}
	}
	
	return files
}

// updateFileList updates the file list display
func (app *GuiApp) updateFileList() {
	if listHwnd, exists := app.controls["file_list"]; exists {
		// Clear the list
		procSendMessage.Call(uintptr(listHwnd), 0x0184, 0, 0) // LB_RESETCONTENT
		
		// Add files to list
		for _, file := range app.selectedFiles {
			fileName := filepath.Base(file)
			fileNamePtr := syscall.StringToUTF16Ptr(fileName)
			procSendMessage.Call(uintptr(listHwnd), 0x0180, 0, uintptr(unsafe.Pointer(fileNamePtr))) // LB_ADDSTRING
		}
	}
}

// showMessage displays a message box
func (app *GuiApp) showMessage(message, title string) {
	messagePtr := syscall.StringToUTF16Ptr(message)
	titlePtr := syscall.StringToUTF16Ptr(title)
	procMessageBox.Call(uintptr(app.hwnd), uintptr(unsafe.Pointer(messagePtr)), 
		uintptr(unsafe.Pointer(titlePtr)), 0x00000040) // MB_ICONINFORMATION
}

// appendOutput adds text to the output area
func (app *GuiApp) appendOutput(text string) {
	if outputHwnd, exists := app.controls["output"]; exists {
		// Get current text length
		length, _, _ := procSendMessage.Call(uintptr(outputHwnd), 0x000E, 0, 0) // WM_GETTEXTLENGTH
		
		// Set selection to end
		procSendMessage.Call(uintptr(outputHwnd), 0x00B1, length, length) // EM_SETSEL
		
		// Replace selection with new text
		textPtr := syscall.StringToUTF16Ptr(text + "\r\n")
		procSendMessage.Call(uintptr(outputHwnd), 0x00C2, 0, uintptr(unsafe.Pointer(textPtr))) // EM_REPLACESEL
	}
}