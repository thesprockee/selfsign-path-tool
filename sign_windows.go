//go:build windows

package main

import (
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Windows API constants
const (
	CERT_SYSTEM_STORE_LOCAL_MACHINE = 0x20000
	CERT_SYSTEM_STORE_CURRENT_USER  = 0x10000
	CERT_STORE_ADD_REPLACE_EXISTING = 3
)

// Windows DLL and function declarations
var (
	crypt32                = syscall.NewLazyDLL("crypt32.dll")
	procCertOpenSystemStore = crypt32.NewProc("CertOpenSystemStoreW")
	procCertAddCertificateContextToStore = crypt32.NewProc("CertAddCertificateContextToStore")
	procCertCreateCertificateContext = crypt32.NewProc("CertCreateCertificateContext")
	procCertCloseStore = crypt32.NewProc("CertCloseStore")
	procCertFreeCertificateContext = crypt32.NewProc("CertFreeCertificateContext")
)

// signFilePlatform signs a file on Windows using a simulated approach
func signFilePlatform(filename string, cert *Certificate) error {
	// On Windows, we would typically use SignTool.exe or the Windows Authenticode APIs
	// For this implementation, we'll create a simple signature file alongside the binary
	// This is a simplified approach since full Authenticode signing requires more complex implementation
	
	signatureFile := filename + ".sig"
	
	// Create a simple signature file indicating the file is signed
	sigContent := fmt.Sprintf("SIGNED_BY=%s\nTIMESTAMP=%s\nCERT_SUBJECT=%s\n", 
		cert.Subject, 
		time.Now().Format(time.RFC3339),  // Use current timestamp
		cert.Cert.Subject.String())
	
	if err := os.WriteFile(signatureFile, []byte(sigContent), 0644); err != nil {
		return fmt.Errorf("failed to create signature file: %w", err)
	}
	
	return nil
}

// getFileSignatureStatusPlatform checks signature status on Windows
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

// removeSelfSignedSignaturePlatform removes self-signed signatures on Windows
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

// installCertificateToStorePlatform installs certificate to Windows certificate store
func installCertificateToStorePlatform(certInterface interface{}) error {
	cert, ok := certInterface.(*x509.Certificate)
	if !ok {
		return fmt.Errorf("invalid certificate type")
	}

	// Try to use PowerShell to install the certificate (fallback approach)
	return installCertificateWithPowerShell(cert)
}

// installCertificateWithPowerShell uses PowerShell to install the certificate
func installCertificateWithPowerShell(cert *x509.Certificate) error {
	// Create temporary certificate file
	tempDir := os.TempDir()
	certFile := filepath.Join(tempDir, "temp_cert.crt")
	
	// Write certificate to temporary file
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary certificate file: %w", err)
	}
	defer os.Remove(certFile)
	defer certOut.Close()
	
	if _, err := certOut.Write(cert.Raw); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}
	certOut.Close()
	
	// Use PowerShell to import the certificate
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf(`
		try {
			$cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2('%s')
			$store = New-Object System.Security.Cryptography.X509Certificates.X509Store([System.Security.Cryptography.X509Certificates.StoreName]::Root, [System.Security.Cryptography.X509Certificates.StoreLocation]::LocalMachine)
			$store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite)
			$store.Add($cert)
			$store.Close()
			Write-Host 'Certificate installed successfully'
		} catch {
			Write-Error $_.Exception.Message
			exit 1
		}`, certFile))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install certificate via PowerShell: %w, output: %s", err, string(output))
	}
	
	return nil
}

// isRunningAsAdmin checks if the current process is running as administrator
func isRunningAsAdmin() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}