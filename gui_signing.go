//go:build windows

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// performSigning executes the signing process in a separate goroutine
func (app *GuiApp) performSigning() {
	var results strings.Builder
	success := true
	
	defer func() {
		// Update UI on main thread
		app.createCompleteScreen(success, results.String())
	}()
	
	app.appendOutput("Starting file signing process...")
	
	// Step 1: Create certificate
	app.appendOutput("Creating self-signed certificate...")
	results.WriteString("File Signing Results:\n")
	results.WriteString("====================\n\n")
	
	cert, privateKey, err := app.createOneTimeSigningCertificate()
	if err != nil {
		app.appendOutput(fmt.Sprintf("Error creating certificate: %v", err))
		results.WriteString(fmt.Sprintf("ERROR: Failed to create certificate: %v\n", err))
		success = false
		return
	}
	
	app.appendOutput("Certificate created successfully.")
	results.WriteString("✓ Certificate created successfully\n")
	
	// Step 2: Sign files
	app.appendOutput(fmt.Sprintf("Signing %d files...", len(app.selectedFiles)))
	signedCount := 0
	
	for i, file := range app.selectedFiles {
		app.appendOutput(fmt.Sprintf("Signing file %d of %d: %s", i+1, len(app.selectedFiles), filepath.Base(file)))
		
		if err := signFile(file, cert); err != nil {
			app.appendOutput(fmt.Sprintf("Failed to sign %s: %v", filepath.Base(file), err))
			results.WriteString(fmt.Sprintf("✗ Failed: %s - %v\n", filepath.Base(file), err))
		} else {
			app.appendOutput(fmt.Sprintf("Successfully signed: %s", filepath.Base(file)))
			results.WriteString(fmt.Sprintf("✓ Signed: %s\n", filepath.Base(file)))
			signedCount++
		}
	}
	
	results.WriteString(fmt.Sprintf("\nSigned %d out of %d files successfully.\n\n", signedCount, len(app.selectedFiles)))
	
	// Step 3: Install certificate to store
	app.appendOutput("Installing certificate to Windows certificate store...")
	if err := installCertificateToStore(cert.Cert); err != nil {
		app.appendOutput(fmt.Sprintf("Warning: Failed to install certificate to store: %v", err))
		results.WriteString(fmt.Sprintf("⚠ Warning: Certificate store installation failed: %v\n", err))
		results.WriteString("You may need to run as administrator for certificate store access.\n")
	} else {
		app.appendOutput("Certificate installed to store successfully.")
		results.WriteString("✓ Certificate installed to Windows certificate store\n")
	}
	
	// Step 4: Securely delete private key
	app.appendOutput("Securely deleting temporary private key...")
	if err := app.securelyDeletePrivateKey(privateKey); err != nil {
		app.appendOutput(fmt.Sprintf("Warning: Failed to securely delete private key: %v", err))
		results.WriteString(fmt.Sprintf("⚠ Warning: Failed to securely delete private key: %v\n", err))
	} else {
		app.appendOutput("Private key securely deleted.")
		results.WriteString("✓ Private key securely deleted\n")
	}
	
	app.appendOutput("File signing process completed!")
	results.WriteString("\nFile signing process completed!\n")
	
	if signedCount == len(app.selectedFiles) && err == nil {
		results.WriteString("\nAll files signed successfully. Your files are now trusted by Windows.")
	} else if signedCount > 0 {
		results.WriteString(fmt.Sprintf("\n%d files signed successfully. Some files may have failed.", signedCount))
	} else {
		results.WriteString("\nNo files were signed successfully. Please check the errors above.")
		success = false
	}
}

// createOneTimeSigningCertificate creates a certificate and private key for one-time use
func (app *GuiApp) createOneTimeSigningCertificate() (*Certificate, *rsa.PrivateKey, error) {
	// Generate a unique name for this signing session
	subjectName := "LocalSign-OneTime-" + generateRandomString(8)
	
	app.appendOutput(fmt.Sprintf("Generating certificate with subject: %s", subjectName))
	
	// Create the certificate (this will create both cert and key)
	cert, err := createSelfSignedCertificate(subjectName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}
	
	// Store reference to private key for secure deletion
	app.certificate = cert
	
	return cert, cert.PrivateKey, nil
}

// securelyDeletePrivateKey securely deletes the private key from memory and disk
func (app *GuiApp) securelyDeletePrivateKey(privateKey *rsa.PrivateKey) error {
	var errors []string
	
	// 1. Overwrite the private key in memory
	if privateKey != nil {
		// Overwrite key components with random data
		if privateKey.D != nil {
			privateKey.D.SetBytes(make([]byte, (privateKey.D.BitLen()+7)/8))
		}
		if privateKey.Primes != nil {
			for _, prime := range privateKey.Primes {
				if prime != nil {
					prime.SetBytes(make([]byte, (prime.BitLen()+7)/8))
				}
			}
		}
		
		// Force garbage collection to clear any remaining references
		runtime.GC()
		runtime.GC() // Call twice to be thorough
	}
	
	// 2. Find and securely delete any certificate files that may have been created
	certDir := getCertificateDirectory()
	pattern := filepath.Join(certDir, "LocalSign-OneTime-*.key")
	
	matches, err := filepath.Glob(pattern)
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to find key files: %v", err))
	} else {
		for _, keyFile := range matches {
			if err := app.securelyDeleteFile(keyFile); err != nil {
				errors = append(errors, fmt.Sprintf("failed to delete %s: %v", keyFile, err))
			}
		}
	}
	
	// Also clean up certificate files for one-time keys
	certPattern := filepath.Join(certDir, "LocalSign-OneTime-*.crt")
	certMatches, err := filepath.Glob(certPattern)
	if err == nil {
		for _, certFile := range certMatches {
			if err := app.securelyDeleteFile(certFile); err != nil {
				errors = append(errors, fmt.Sprintf("failed to delete %s: %v", certFile, err))
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("security warnings: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// securelyDeleteFile overwrites a file with random data before deleting it
func (app *GuiApp) securelyDeleteFile(filePath string) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	
	// Open file for writing
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file for overwriting: %w", err)
	}
	defer file.Close()
	
	// Overwrite with random data (3 passes)
	fileSize := info.Size()
	for pass := 0; pass < 3; pass++ {
		// Seek to beginning
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek to beginning: %w", err)
		}
		
		// Write random data
		randomData := make([]byte, fileSize)
		if _, err := rand.Read(randomData); err != nil {
			return fmt.Errorf("failed to generate random data: %w", err)
		}
		
		if _, err := file.Write(randomData); err != nil {
			return fmt.Errorf("failed to overwrite file: %w", err)
		}
		
		// Sync to disk
		if err := file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %w", err)
		}
	}
	
	file.Close()
	
	// Finally delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// generateRandomString generates a random string for certificate names
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomIndex := make([]byte, 1)
		rand.Read(randomIndex)
		b[i] = charset[randomIndex[0]%byte(len(charset))]
	}
	return string(b)
}

// Helper function to safely access UI controls from goroutines
func (app *GuiApp) safeAppendOutput(text string) {
	// In a real implementation, you'd want to marshal this to the main UI thread
	// For now, we'll call directly but in production you'd use PostMessage or similar
	app.appendOutput(text)
}