package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Certificate represents a signing certificate
type Certificate struct {
	Subject    string
	Cert       *x509.Certificate
	PrivateKey *rsa.PrivateKey
}

// getCertificate obtains a certificate for signing - either from files or by creating one
func getCertificate() (*Certificate, error) {
	if *flagCertFile != "" && *flagKeyFile != "" {
		return loadCertificateFromFile(*flagCertFile, *flagKeyFile)
	}
	return getOrCreateSelfSignedCertificate(*flagName)
}

// loadCertificateFromFile loads a certificate and private key from files
func loadCertificateFromFile(certFile, keyFile string) (*Certificate, error) {
	// Load certificate file
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file %s: %w", certFile, err)
	}

	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate from %s", certFile)
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate from %s: %w", certFile, err)
	}

	// Load private key file
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file %s: %w", keyFile, err)
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM private key from %s", keyFile)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key from %s: %w", keyFile, err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key from %s is not an RSA key", keyFile)
		}
	}

	return &Certificate{
		Subject:    cert.Subject.CommonName,
		Cert:       cert,
		PrivateKey: privateKey,
	}, nil
}

// getOrCreateSelfSignedCertificate gets an existing certificate or creates a new one
func getOrCreateSelfSignedCertificate(subjectName string) (*Certificate, error) {
	// Try to load existing certificate
	certDir := getCertificateDirectory()
	certFile := filepath.Join(certDir, fmt.Sprintf("%s.crt", subjectName))
	keyFile := filepath.Join(certDir, fmt.Sprintf("%s.key", subjectName))

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			fmt.Printf("Using existing certificate: %s\n", subjectName)
			return loadCertificateFromFile(certFile, keyFile)
		}
	}

	// Create new certificate
	fmt.Printf("Creating new self-signed certificate with subject: %s\n", subjectName)
	return createSelfSignedCertificate(subjectName)
}

// createSelfSignedCertificate creates a new self-signed certificate
func createSelfSignedCertificate(subjectName string) (*Certificate, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: subjectName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(3, 0, 0), // Valid for 3 years
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created certificate: %w", err)
	}

	// Save certificate and key to files
	if err := saveCertificateFiles(subjectName, cert, privateKey); err != nil {
		fmt.Printf("Warning: Failed to save certificate to disk: %v\n", err)
	}

	// Try to install certificate to system store (platform-specific)
	if err := installCertificateToStore(cert); err != nil {
		fmt.Printf("Warning: Failed to install certificate to system store: %v\n", err)
		fmt.Printf("Certificate created but not installed to system trust store.\n")
	} else {
		fmt.Printf("Certificate installed to system trust store.\n")
	}

	return &Certificate{
		Subject:    subjectName,
		Cert:       cert,
		PrivateKey: privateKey,
	}, nil
}

// getCertificateDirectory returns the directory where certificates are stored
func getCertificateDirectory() string {
	var certDir string
	
	if runtime.GOOS == "windows" {
		// Use AppData on Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		certDir = filepath.Join(appData, "selfsign-path-tool", "certificates")
	} else {
		// Use ~/.local/share on Unix-like systems
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to /tmp
			certDir = "/tmp/selfsign-path-tool-certificates"
		} else {
			certDir = filepath.Join(homeDir, ".local", "share", "selfsign-path-tool", "certificates")
		}
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(certDir, 0700); err != nil {
		fmt.Printf("Warning: Failed to create certificate directory %s: %v\n", certDir, err)
	}

	return certDir
}

// saveCertificateFiles saves the certificate and private key to disk
func saveCertificateFiles(subjectName string, cert *x509.Certificate, privateKey *rsa.PrivateKey) error {
	certDir := getCertificateDirectory()
	
	// Save certificate
	certFile := filepath.Join(certDir, fmt.Sprintf("%s.crt", subjectName))
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Save private key
	keyFile := filepath.Join(certDir, fmt.Sprintf("%s.key", subjectName))
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	if err := keyOut.Chmod(0600); err != nil {
		fmt.Printf("Warning: Failed to set key file permissions: %v\n", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	fmt.Printf("Saved certificate files to: %s\n", certDir)
	return nil
}