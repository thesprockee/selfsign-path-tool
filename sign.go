package main

// SignatureStatus represents the status of a file's signature
type SignatureStatus struct {
	Status               string
	SignerCertificate    string
	TimestampCertificate string
	IsSelfSigned         bool
}

// signFile signs a file with the given certificate
func signFile(filename string, cert *Certificate) error {
	return signFilePlatform(filename, cert)
}

// getFileSignatureStatus checks the signature status of a file
func getFileSignatureStatus(filename string) (*SignatureStatus, error) {
	return getFileSignatureStatusPlatform(filename)
}

// removeSelfSignedSignature removes self-signed signatures from a file
func removeSelfSignedSignature(filename string) (bool, error) {
	return removeSelfSignedSignaturePlatform(filename)
}

// installCertificateToStore installs the certificate to the system trust store
func installCertificateToStore(cert interface{}) error {
	return installCertificateToStorePlatform(cert)
}