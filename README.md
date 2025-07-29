# local-sign

A PowerShell utility to manage and apply self-signed code signatures to executables and libraries.

## Quick Start - One-Line Installation

For Oculus VR users, you can install and sign your Oculus applications with a single command:

```powershell
iwr -useb https://raw.githubusercontent.com/thesprockee/local-sign/main/install.ps1 | iex
```

**⚠️ Important: Run PowerShell as Administrator for the one-line installation.**

This express installation will automatically:
- Download the latest signing tool
- Prompt to remove any existing LocalSign certificates
- Create a new signing certificate named "LocalSign-OculusVR"
- Install it to the Trusted Root Certification Authorities store
- Sign all `.exe` and `.dll` files in:
  - `C:\Program Files\Oculus\Software\Software\ready-at-dawn-echo-arena`
  - `C:\echovr`

## Overview

The sign-tool script automates the process of code signing using a self-signed certificate. Upon first run, it generates a new self-signed code-signing certificate and imports it into the system's Trusted Root Certification Authorities store. Subsequent runs will use this existing certificate.

The script can sign new files, re-sign existing files, or append a signature. It can also be used to check the signature status of files or to remove its own self-signed signature. It accepts one or more files or glob-like patterns as input.

## Usage

### Express Installation (Recommended for Oculus VR)

```powershell
# Basic one-liner installation
iwr -useb https://raw.githubusercontent.com/thesprockee/local-sign/main/install.ps1 | iex

# Force installation without prompts
iwr -useb https://raw.githubusercontent.com/thesprockee/local-sign/main/install.ps1 | iex -Command "& { . ([ScriptBlock]::Create(\$input)); Install-LocalSign -Force }"

# Install with custom certificate name
iwr -useb https://raw.githubusercontent.com/thesprockee/local-sign/main/install.ps1 | iex -Command "& { . ([ScriptBlock]::Create(\$input)); Install-LocalSign -CertName 'MyCustomCert' }"
```

### Manual Tool Usage

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
