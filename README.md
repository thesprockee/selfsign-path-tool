# local-sign

A PowerShell utility to manage and apply self-signed code signatures to executables and libraries.

## Overview

The sign-tool script automates the process of code signing using a self-signed certificate. Upon first run, it generates a new self-signed code-signing certificate and imports it into the system's Trusted Root Certification Authorities store. Subsequent runs will use this existing certificate.

The script can sign new files, re-sign existing files, or append a signature. It can also be used to check the signature status of files or to remove its own self-signed signature. It accepts one or more files or glob-like patterns as input.

## Usage

```powershell
# Sign a single executable
.\sign-tool.ps1 myapp.exe

# Sign all DLLs in a directory and its subdirectories
.\sign-tool.ps1 -Recurse "bin/**/*.dll"

# Check the signature status of all executables in the current directory
.\sign-tool.ps1 -Status "*.exe"

# Sign files using a custom-named certificate
.\sign-tool.ps1 -Name "My Custom Cert" myapp.exe

# Sign a file using specific certificate and key files
.\sign-tool.ps1 -CertFile "/path/to/my.crt" -KeyFile "/path/to/my.key" myapp.exe

# Remove self-signatures from all files in a release folder
.\sign-tool.ps1 -Clear -Recurse "release/"
```

## Requirements

- PowerShell 5.1 or later (PowerShell Core 6+ recommended)
- Windows (for certificate store operations) or cross-platform PowerShell
- Administrator privileges (recommended for certificate installation)

## Features

- **Automatic Certificate Generation**: Creates self-signed certificates on first run
- **Certificate Management**: Reuses existing certificates for subsequent signings
- **Pattern Support**: Supports file patterns and glob-like expressions
- **Recursive Operations**: Can process directories recursively
- **Status Checking**: View signature status of files
- **Signature Removal**: Remove self-signed signatures while preserving other signatures
- **Cross-platform**: Works on Windows, Linux, and macOS with PowerShell Core 
