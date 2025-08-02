//go:build linux

package main

import (
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// signFilePlatform signs a file on Linux using a simulated approach
func signFilePlatform(filename string, cert *Certificate) error {
	// On Linux, there's no standard code signing like Windows Authenticode
	// We'll create a detached signature file similar to GPG signatures
	
	signatureFile := filename + ".sig"
	
	// Create a simple signature file indicating the file is signed
	sigContent := fmt.Sprintf("SIGNED_BY=%s\nTIMESTAMP=%s\nCERT_SUBJECT=%s\nPLATFORM=linux\n", 
		cert.Subject, 
		"2024-01-01T00:00:00Z",  // Simplified timestamp
		cert.Cert.Subject.String())
	
	if err := os.WriteFile(signatureFile, []byte(sigContent), 0644); err != nil {
		return fmt.Errorf("failed to create signature file: %w", err)
	}
	
	return nil
}

// getFileSignatureStatusPlatform checks signature status on Linux
func getFileSignatureStatusPlatform(filename string) (*SignatureStatus, error) {
	// Check if our simple signature file exists
	signatureFile := filename + ".sig"
	
	if _, err := os.Stat(signatureFile); os.IsNotExist(err) {
		return &SignatureStatus{
			Status: "NotSigned",
		}, nil
	}
	
	// Read signature file
	sigContent, err := os.ReadFile(signatureFile)
	if err != nil {
		return &SignatureStatus{
			Status: "Error reading signature",
		}, nil
	}
	
	lines := strings.Split(string(sigContent), "\n")
	status := &SignatureStatus{
		Status: "Valid",
		IsSelfSigned: true,
	}
	
	for _, line := range lines {
		if strings.HasPrefix(line, "SIGNED_BY=") {
			status.SignerCertificate = strings.TrimPrefix(line, "SIGNED_BY=")
		} else if strings.HasPrefix(line, "CERT_SUBJECT=") {
			subject := strings.TrimPrefix(line, "CERT_SUBJECT=")
			// Check if self-signed (simplified check)
			status.IsSelfSigned = strings.Contains(subject, "LocalSign")
		}
	}
	
	return status, nil
}

// removeSelfSignedSignaturePlatform removes self-signed signatures on Linux
func removeSelfSignedSignaturePlatform(filename string) (bool, error) {
	signatureFile := filename + ".sig"
	
	// Check if signature file exists and is self-signed
	status, err := getFileSignatureStatusPlatform(filename)
	if err != nil {
		return false, err
	}
	
	if status.Status == "NotSigned" {
		return false, nil
	}
	
	if status.IsSelfSigned {
		if err := os.Remove(signatureFile); err != nil {
			return false, fmt.Errorf("failed to remove signature file: %w", err)
		}
		return true, nil
	}
	
	return false, nil
}

// installCertificateToStorePlatform installs certificate to Linux certificate store
func installCertificateToStorePlatform(certInterface interface{}) error {
	cert, ok := certInterface.(*x509.Certificate)
	if !ok {
		return fmt.Errorf("invalid certificate type")
	}

	// Try to install to system certificate store
	// Different distributions have different locations and tools
	return installCertificateLinuxSystem(cert)
}

// installCertificateLinuxSystem tries to install certificate to system store
func installCertificateLinuxSystem(cert *x509.Certificate) error {
	// Common certificate directories on Linux
	certDirs := []string{
		"/usr/local/share/ca-certificates",
		"/etc/ssl/certs",
		"/etc/pki/ca-trust/source/anchors",
	}

	certName := fmt.Sprintf("selfsign-path-%s.crt", cert.Subject.CommonName)
	
	// Try each directory
	for _, certDir := range certDirs {
		if _, err := os.Stat(certDir); err != nil {
			continue // Directory doesn't exist
		}
		
		certPath := filepath.Join(certDir, certName)
		
		// Try to write certificate
		if err := os.WriteFile(certPath, cert.Raw, 0644); err != nil {
			continue // Can't write to this directory
		}
		
		// Try to update certificate store
		updateCertStore(certDir)
		
		fmt.Printf("Certificate installed to: %s\n", certPath)
		return nil
	}
	
	// If system installation fails, install to user directory
	return installCertificateLinuxUser(cert)
}

// installCertificateLinuxUser installs certificate to user certificate store
func installCertificateLinuxUser(cert *x509.Certificate) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	// Create user certificate directory
	certDir := filepath.Join(homeDir, ".local", "share", "ca-certificates")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create user certificate directory: %w", err)
	}
	
	certName := fmt.Sprintf("selfsign-path-%s.crt", cert.Subject.CommonName)
	certPath := filepath.Join(certDir, certName)
	
	if err := os.WriteFile(certPath, cert.Raw, 0644); err != nil {
		return fmt.Errorf("failed to write certificate to user directory: %w", err)
	}
	
	fmt.Printf("Certificate installed to user directory: %s\n", certPath)
	fmt.Printf("Note: Certificate may not be trusted system-wide without administrator privileges.\n")
	
	return nil
}

// updateCertStore runs commands to update the certificate store
func updateCertStore(certDir string) {
	// Try different update commands based on the certificate directory
	switch {
	case strings.Contains(certDir, "ca-certificates"):
		// Debian/Ubuntu
		exec.Command("update-ca-certificates").Run()
	case strings.Contains(certDir, "ca-trust"):
		// Red Hat/CentOS/Fedora
		exec.Command("update-ca-trust").Run()
	}
}