//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Windows API constants for GUI
const (
	// Window styles
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_TABSTOP          = 0x00010000
	WS_BORDER           = 0x00800000
	
	// Button styles
	BS_DEFPUSHBUTTON = 0x00000001
	BS_PUSHBUTTON    = 0x00000000
	
	// Edit control styles
	ES_MULTILINE = 0x0004
	ES_READONLY  = 0x0800
	ES_AUTOVSCROLL = 0x0040
	
	// List box styles
	LBS_STANDARD = 0x00A00003
	
	// Window messages
	WM_COMMAND = 0x0111
	WM_CLOSE   = 0x0010
	WM_DESTROY = 0x0002
	BN_CLICKED = 0
	
	// Dialog box return codes
	IDOK     = 1
	IDCANCEL = 2
	
	// File dialog constants
	OFN_FILEMUSTEXIST     = 0x00001000
	OFN_PATHMUSTEXIST     = 0x00000800
	OFN_ALLOWMULTISELECT  = 0x00000200
	OFN_EXPLORER          = 0x00080000
	
	// Control IDs
	ID_BUTTON_NEXT    = 1001
	ID_BUTTON_BACK    = 1002
	ID_BUTTON_CANCEL  = 1003
	ID_BUTTON_BROWSE  = 1004
	ID_LISTBOX_FILES  = 1005
	ID_EDIT_OUTPUT    = 1006
)

// Windows API structures
type POINT struct {
	X, Y int32
}

type MSG struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type WNDCLASS struct {
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
}

type OPENFILENAME struct {
	LStructSize       uint32
	HwndOwner         syscall.Handle
	HInstance         syscall.Handle
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
}

// Windows API functions
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	
	procDefWindowProc     = user32.NewProc("DefWindowProcW")
	procRegisterClass     = user32.NewProc("RegisterClassW")
	procCreateWindowEx    = user32.NewProc("CreateWindowExW")
	procShowWindow        = user32.NewProc("ShowWindow")
	procUpdateWindow      = user32.NewProc("UpdateWindow")
	procGetMessage        = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessage   = user32.NewProc("DispatchMessageW")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procPostMessage       = user32.NewProc("PostMessageW")
	procLoadCursor        = user32.NewProc("LoadCursorW")
	procSetWindowText     = user32.NewProc("SetWindowTextW")
	procGetWindowText     = user32.NewProc("GetWindowTextW")
	procSendMessage       = user32.NewProc("SendMessageW")
	procMessageBox        = user32.NewProc("MessageBoxW")
	procGetOpenFileName   = comdlg32.NewProc("GetOpenFileNameW")
	procGetModuleHandle   = kernel32.NewProc("GetModuleHandleW")
	procIsUserAnAdmin     = shell32.NewProc("IsUserAnAdmin")
	procCreateIconFromResource = user32.NewProc("CreateIconFromResource")
)

// GUI state
type GuiApp struct {
	hwnd         syscall.Handle
	currentStep  int
	selectedFiles []string
	certificate  *Certificate
	
	// UI controls
	controls map[string]syscall.Handle
}

// Wizard steps
const (
	STEP_WELCOME = iota
	STEP_FILE_SELECTION
	STEP_CONFIRM
	STEP_PROCESSING
	STEP_COMPLETE
)

// runGUI starts the Windows GUI application
func runGUI() error {
	app := &GuiApp{
		currentStep: STEP_WELCOME,
		controls:    make(map[string]syscall.Handle),
	}
	
	// Register window class
	if err := app.registerWindowClass(); err != nil {
		return fmt.Errorf("failed to register window class: %w", err)
	}
	
	// Create main window
	if err := app.createMainWindow(); err != nil {
		return fmt.Errorf("failed to create main window: %w", err)
	}
	
	// Show window
	showWindow(app.hwnd, 1) // SW_SHOWNORMAL
	updateWindow(app.hwnd)
	
	// Create initial UI
	app.createWelcomeScreen()
	
	// Message loop
	return app.messageLoop()
}

// registerWindowClass registers the window class
func (app *GuiApp) registerWindowClass() error {
	hInstance, _, _ := procGetModuleHandle.Call(0)
	
	className := syscall.StringToUTF16Ptr("SelfSignGUIClass")
	
	wc := WNDCLASS{
		LpfnWndProc:   syscall.NewCallback(app.windowProc),
		HInstance:     syscall.Handle(hInstance),
		LpszClassName: className,
		HCursor:       loadCursor(0, 32512), // IDC_ARROW
		HIcon:         loadIconFromMemory(), // Use embedded icon
	}
	
	ret, _, _ := procRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		return fmt.Errorf("RegisterClassW failed")
	}
	
	return nil
}

// createMainWindow creates the main application window
func (app *GuiApp) createMainWindow() error {
	hInstance, _, _ := procGetModuleHandle.Call(0)
	
	className := syscall.StringToUTF16Ptr("SelfSignGUIClass")
	windowName := syscall.StringToUTF16Ptr("File Signing Tool")
	
	hwnd, _, _ := procCreateWindowEx.Call(
		0,                                    // dwExStyle
		uintptr(unsafe.Pointer(className)),  // lpClassName
		uintptr(unsafe.Pointer(windowName)), // lpWindowName
		WS_OVERLAPPEDWINDOW,                  // dwStyle
		100, 100,                             // x, y
		600, 500,                             // nWidth, nHeight
		0,                                    // hWndParent
		0,                                    // hMenu
		hInstance,                            // hInstance
		0,                                    // lpParam
	)
	
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed")
	}
	
	app.hwnd = syscall.Handle(hwnd)
	return nil
}

// windowProc handles window messages
func (app *GuiApp) windowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		return app.handleCommand(wParam, lParam)
	case WM_CLOSE:
		procPostQuitMessage.Call(0)
		return 0
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	
	ret, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

// handleCommand handles WM_COMMAND messages
func (app *GuiApp) handleCommand(wParam, lParam uintptr) uintptr {
	controlID := wParam & 0xFFFF
	notification := (wParam >> 16) & 0xFFFF
	
	if notification == BN_CLICKED {
		switch controlID {
		case ID_BUTTON_NEXT:
			app.nextStep()
		case ID_BUTTON_BACK:
			app.previousStep()
		case ID_BUTTON_CANCEL:
			procPostQuitMessage.Call(0)
		case ID_BUTTON_BROWSE:
			app.browseFiles()
		}
	}
	
	return 0
}

// Helper functions for Windows API calls
func showWindow(hwnd syscall.Handle, nCmdShow int) {
	procShowWindow.Call(uintptr(hwnd), uintptr(nCmdShow))
}

func updateWindow(hwnd syscall.Handle) {
	procUpdateWindow.Call(uintptr(hwnd))
}

func loadCursor(hInstance syscall.Handle, lpCursorName uintptr) syscall.Handle {
	ret, _, _ := procLoadCursor.Call(uintptr(hInstance), lpCursorName)
	return syscall.Handle(ret)
}

// loadIconFromMemory creates an icon from embedded data
func loadIconFromMemory() syscall.Handle {
	iconData := GetIconData()
	if len(iconData) == 0 {
		return 0
	}
	
	// CreateIconFromResource expects the icon data to be in the correct format
	// For ICO files, we need to skip the ICO header and use the actual icon data
	// ICO header is 6 bytes + 16 bytes per icon entry
	// For simplicity, we'll try to use the data as-is first
	ret, _, _ := procCreateIconFromResource.Call(
		uintptr(unsafe.Pointer(&iconData[0])),
		uintptr(len(iconData)),
		1, // fIcon (TRUE for icon, FALSE for cursor)
		0x00030000, // dwVersion (standard version)
	)
	
	return syscall.Handle(ret)
}

func setWindowText(hwnd syscall.Handle, text string) {
	textPtr := syscall.StringToUTF16Ptr(text)
	procSetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(textPtr)))
}

func createWindow(className, windowName string, style uint32, x, y, width, height int, parent syscall.Handle, menu uintptr, hInstance syscall.Handle) syscall.Handle {
	classNamePtr := syscall.StringToUTF16Ptr(className)
	windowNamePtr := syscall.StringToUTF16Ptr(windowName)
	
	ret, _, _ := procCreateWindowEx.Call(
		0,                                         // dwExStyle
		uintptr(unsafe.Pointer(classNamePtr)),    // lpClassName
		uintptr(unsafe.Pointer(windowNamePtr)),   // lpWindowName
		uintptr(style),                           // dwStyle
		uintptr(x), uintptr(y),                   // x, y
		uintptr(width), uintptr(height),          // nWidth, nHeight
		uintptr(parent),                          // hWndParent
		menu,                                     // hMenu
		uintptr(hInstance),                       // hInstance
		0,                                        // lpParam
	)
	
	return syscall.Handle(ret)
}

// messageLoop runs the main message loop
func (app *GuiApp) messageLoop() error {
	var msg MSG
	
	for {
		ret, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		
		if int32(ret) == -1 {
			return fmt.Errorf("GetMessage failed")
		} else if ret == 0 {
			break // WM_QUIT
		}
		
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
	
	return nil
}

// destroyControl safely destroys a control if it exists
func (app *GuiApp) destroyControl(name string) {
	if hwnd, exists := app.controls[name]; exists {
		// Destroy window
		user32.NewProc("DestroyWindow").Call(uintptr(hwnd))
		delete(app.controls, name)
	}
}

// clearAllControls removes all current UI controls
func (app *GuiApp) clearAllControls() {
	for name := range app.controls {
		app.destroyControl(name)
	}
}