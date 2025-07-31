# selfsign-path-tool

A PowerShell utility to manage and apply self-signed code signatures to executables and libraries.

## Quick Start - One-Line Installation

For EVR users, you can install and sign your EVR applications with a single command:

```powershell
iwr -useb https://raw.githubusercontent.com/thesprockee/selfsign-path-tool/main/install.ps1 | iex
```

**⚠️ Important: Run PowerShell as Administrator for the one-line installation.**

> **Alternative (more robust):** If you encounter any issues with the one-liner, use this approach instead:
> ```powershell
> $tempFile = New-TemporaryFile; iwr -useb https://raw.githubusercontent.com/thesprockee/selfsign-path-tool/main/install.ps1 -OutFile $tempFile; if (Get-Command pwsh -ErrorAction SilentlyContinue) { pwsh -File $tempFile } else { powershell -File $tempFile }; Remove-Item $tempFile
> ```

This express installation will automatically:
- Download the latest signing tool
- Prompt to remove any existing LocalSign certificates
- Create a new signing certificate named "LocalSign-EVR"
- Install it to the Trusted Root Certification Authorities store
- Sign all `.exe` and `.dll` files in:
  - `C:\Program Files\Oculus\Software\Software\ready-at-dawn-echo-arena`
  - `C:\echovr`

## Overview

The selfsign-path script automates the process of code signing using a self-signed certificate. Upon first run, it generates a new self-signed code-signing certificate and imports it into the system's Trusted Root Certification Authorities store. Subsequent runs will use this existing certificate.

The script can sign new files, re-sign existing files, or append a signature. It can also be used to check the signature status of files or to remove its own self-signed signature. It accepts one or more files or glob-like patterns as input.

## Usage

### Express Installation (Recommended for EVR)

```powershell
# Basic one-liner installation
iwr -useb https://raw.githubusercontent.com/thesprockee/selfsign-path-tool/main/install.ps1 | iex

# Force installation without prompts
iwr -useb https://raw.githubusercontent.com/thesprockee/selfsign-path-tool/main/install.ps1 | iex -Command "& { . ([ScriptBlock]::Create(\$input)); Install-LocalSign -Force }"

# Install with custom certificate name
iwr -useb https://raw.githubusercontent.com/thesprockee/selfsign-path-tool/main/install.ps1 | iex -Command "& { . ([ScriptBlock]::Create(\$input)); Install-LocalSign -CertName 'MyCustomCert' }"
```

### Manual Tool Usage

```powershell
# Sign a single executable
.\selfsign-path.ps1 myapp.exe

# Sign all DLLs in a directory and its subdirectories
.\selfsign-path.ps1 -Recurse "bin/**/*.dll"

# Check the signature status of all executables in the current directory
.\selfsign-path.ps1 -Status "*.exe"

# Sign files using a custom-named certificate
.\selfsign-path.ps1 -Name "My Custom Cert" myapp.exe

# Sign a file using specific certificate and key files
.\selfsign-path.ps1 -CertFile "/path/to/my.crt" -KeyFile "/path/to/my.key" myapp.exe

# Remove self-signatures from all files in a release folder
.\selfsign-path.ps1 -Clear -Recurse "release/"
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
- **Automated Releases**: Automatic creation of signed releases when semantic version tags are pushed

## Automated Release Process

This repository includes automated release workflows that trigger when semantic version tags are pushed. The automation:

1. **Creates a draft release** for the new version
2. **Generates a versioned script** (`selfsign-path-v{version}.ps1`) 
3. **Signs the script** using the project's code signing certificate (if configured)
4. **Generates a changelog** from git commit history
5. **Attaches the signed script** to the release

### Creating a Release

To create a new release, you can either:

**Option 1: Use the helper script (recommended)**
```powershell
# Create a new release version
.\create-release.ps1 1.0.0

# With a custom message
.\create-release.ps1 1.0.0 -Message "First stable release"

# Dry run to see what would happen
.\create-release.ps1 1.0.0 -DryRun
```

**Option 2: Manual tagging**
```bash
# Tag your commit with a semantic version
git tag v1.0.0
git push origin v1.0.0
```

The automation will create a draft release that you can review and publish.

### Code Signing Setup

To enable automatic script signing in releases, add the following repository secrets:

- `SIGNING_CERT`: Base64-encoded .pfx certificate file
- `SIGNING_CERT_PASSWORD`: Password for the .pfx certificate

```bash
# Example: Convert certificate to base64 for the secret
base64 -i your-certificate.pfx | pbcopy  # macOS
base64 -w 0 your-certificate.pfx | clip  # Windows
```

If these secrets are not configured, releases will still be created but scripts will not be signed.
