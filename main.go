package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "1.0.0"

// Command line flags
var (
	flagRecurse  = flag.Bool("r", false, "Recursively search for and process files in any specified directories")
	flagName     = flag.String("n", "LocalSign-SelfSigned", "Specify the subject name of the certificate to use for signing")
	flagCertFile = flag.String("c", "", "Specify the path to the certificate file (.cer or .pem)")
	flagKeyFile  = flag.String("k", "", "Specify the path to the private key file (.pvk or .key)")
	flagClear    = flag.Bool("clear", false, "Remove self-signed signatures created by this tool from the specified files")
	flagStatus   = flag.Bool("status", false, "Print the signing status of the specified files instead of signing them")
	flagHelp     = flag.Bool("h", false, "Display help documentation and exit")
	flagVersion  = flag.Bool("version", false, "Display version information and exit")
	flagGUI      = flag.Bool("gui", false, "Launch the graphical user interface (Windows only)")
)

func init() {
	// Set custom usage message
	flag.Usage = showHelp
}

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Printf("selfsign-path-tool version %s\n", version)
		fmt.Printf("Cross-platform code signing utility\n")
		os.Exit(0)
	}

	// Check for GUI mode (Windows only)
	if *flagGUI {
		if runtime.GOOS != "windows" {
			fmt.Fprintf(os.Stderr, "Error: GUI mode is only supported on Windows.\n")
			os.Exit(1)
		}
		
		fmt.Printf("Starting GUI mode...\n")
		if err := runGUI(); err != nil {
			fmt.Fprintf(os.Stderr, "GUI Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *flagHelp || (flag.NArg() == 0 && !*flagClear && !*flagStatus) {
		showHelp()
		os.Exit(0)
	}

	// Validate certificate and key file parameters
	if *flagCertFile != "" && *flagKeyFile == "" {
		fmt.Fprintf(os.Stderr, "Error: --cert-file requires --key-file to be specified.\n")
		os.Exit(1)
	}

	if *flagKeyFile != "" && *flagCertFile == "" {
		fmt.Fprintf(os.Stderr, "Error: --key-file requires --cert-file to be specified.\n")
		os.Exit(1)
	}

	// Get file patterns from remaining arguments
	patterns := flag.Args()
	if len(patterns) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No files or patterns specified.\n")
		os.Exit(1)
	}

	// Main execution logic
	if err := run(patterns); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(patterns []string) error {
	// Get target files from patterns
	files, err := getTargetFiles(patterns, *flagRecurse)
	if err != nil {
		return fmt.Errorf("failed to get target files: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("No files found matching the specified patterns.\n")
		return nil
	}

	fmt.Printf("Found %d file(s) to process.\n", len(files))

	if *flagStatus {
		return showStatus(files)
	} else if *flagClear {
		return clearSignatures(files)
	} else {
		return signFiles(files)
	}
}

func getTargetFiles(patterns []string, recursive bool) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Check if it's a directory
		if info, err := os.Stat(pattern); err == nil && info.IsDir() {
			dirFiles, err := getFilesFromDirectory(pattern, recursive)
			if err != nil {
				return nil, fmt.Errorf("failed to get files from directory %s: %w", pattern, err)
			}
			for _, file := range dirFiles {
				if !seen[file] {
					files = append(files, file)
					seen[file] = true
				}
			}
		} else if strings.ContainsAny(pattern, "*?[]") {
			// It's a glob pattern
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
			}
			for _, match := range matches {
				if info, err := os.Stat(match); err == nil && !info.IsDir() {
					if !seen[match] {
						files = append(files, match)
						seen[match] = true
					}
				}
			}
		} else {
			// It's a specific file
			if _, err := os.Stat(pattern); err == nil {
				if !seen[pattern] {
					files = append(files, pattern)
					seen[pattern] = true
				}
			} else {
				fmt.Printf("Warning: File not found: %s\n", pattern)
			}
		}
	}

	return files, nil
}

func getFilesFromDirectory(dir string, recursive bool) ([]string, error) {
	var files []string
	
	// Executable file extensions we care about
	extensions := map[string]bool{
		".exe": true, ".dll": true, ".msi": true, ".sys": true,
		".com": true, ".ocx": true, ".scr": true, ".cpl": true,
	}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has an extension we care about
		ext := strings.ToLower(filepath.Ext(path))
		if extensions[ext] {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(dir, walkFunc); err != nil {
		return nil, err
	}

	return files, nil
}

func showStatus(files []string) error {
	fmt.Printf("\nSignature Status Report:\n")
	fmt.Printf("========================================\n")

	for _, file := range files {
		status, err := getFileSignatureStatus(file)
		if err != nil {
			fmt.Printf("\nFile: %s\n", file)
			fmt.Printf("Status: Error - %v\n", err)
		} else {
			fmt.Printf("\nFile: %s\n", file)
			fmt.Printf("Status: %s\n", status.Status)
			if status.SignerCertificate != "" {
				fmt.Printf("Signer: %s\n", status.SignerCertificate)
				fmt.Printf("Self-signed: %t\n", status.IsSelfSigned)
			}
			if status.TimestampCertificate != "" {
				fmt.Printf("Timestamp: %s\n", status.TimestampCertificate)
			}
		}
	}

	return nil
}

func clearSignatures(files []string) error {
	fmt.Printf("Removing self-signed signatures...\n")
	removedCount := 0

	for _, file := range files {
		if removed, err := removeSelfSignedSignature(file); err != nil {
			fmt.Printf("Error processing %s: %v\n", file, err)
		} else if removed {
			fmt.Printf("Removed self-signed signature from: %s\n", file)
			removedCount++
		}
	}

	fmt.Printf("\nRemoved signatures from %d file(s).\n", removedCount)
	return nil
}

func signFiles(files []string) error {
	// Get or create certificate
	cert, err := getCertificate()
	if err != nil {
		return fmt.Errorf("failed to obtain signing certificate: %w", err)
	}

	fmt.Printf("Signing files with certificate: %s\n", cert.Subject)
	signedCount := 0

	for _, file := range files {
		if err := signFile(file, cert); err != nil {
			fmt.Printf("Warning: Failed to sign %s: %v\n", file, err)
		} else {
			fmt.Printf("Successfully signed: %s\n", file)
			signedCount++
		}
	}

	fmt.Printf("\nSuccessfully signed %d out of %d file(s).\n", signedCount, len(files))
	return nil
}

func showHelp() {
	fmt.Printf(`NAME
    selfsign-path - A utility to manage and apply self-signed code signatures to executables and libraries.

SYNOPSIS
    selfsign-path [OPTIONS] file_or_pattern...

DESCRIPTION
    The selfsign-path tool automates the process of code signing using a self-signed
    certificate. Upon first run, it generates a new self-signed code-signing
    certificate and installs it into the system's certificate store. Subsequent runs 
    will use this existing certificate.

    The tool can sign new files, re-sign existing files, or append a signature.
    It can also be used to check the signature status of files or to remove its
    own self-signed signature. It accepts one or more files or glob-like patterns
    as input.

ARGUMENTS
    file_or_pattern
        One or more space-separated paths to files or directories.
        Supports glob-like patterns (e.g., *.exe, bin/*) to specify multiple files.

OPTIONS
    -r, --recurse
        Recursively search for and process files in any specified directories.

    -n <CERT_NAME>, --name <CERT_NAME>
        Specify the subject name of the certificate to use for signing. If not
        found, a new certificate with this name is created. Defaults to a
        pre-configured name if not provided.

    -c <CERT_FILE>, --cert-file <CERT_FILE>
        Specify the path to the certificate file (.cer or .pem). This bypasses
        the default certificate generation and lookup. Requires --key-file.

    -k <KEY_FILE>, --key-file <KEY_FILE>
        Specify the path to the private key file (.pvk or .key). Required if
        --cert-file is used.

    --clear
        Remove self-signed signatures created by this tool from the specified
        files. It will not affect other valid signatures.

    --status
        Print the signing status of the specified files instead of signing them.

    -h, --help
        Display this help documentation and exit.

    --version
        Display version information and exit.

    --gui
        Launch the graphical user interface (Windows only).

EXAMPLES
    Sign a single executable:
        selfsign-path myapp.exe

    Sign all DLLs in a directory and its subdirectories:
        selfsign-path -r 'bin/*.dll'

    Check the signature status of all executables in the current directory:
        selfsign-path --status *.exe

    Sign files using a custom-named certificate:
        selfsign-path -n "My Custom Cert" myapp.exe

    Sign a file using specific certificate and key files:
        selfsign-path --cert-file /path/to/my.crt --key-file /path/to/my.key myapp.exe

    Remove self-signatures from all files in a release folder:
        selfsign-path --clear -r release/

    Launch the graphical user interface (Windows only):
        selfsign-path --gui

`)
}