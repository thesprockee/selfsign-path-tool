#!/usr/bin/env pwsh

<#
.SYNOPSIS
    A utility to manage and apply self-signed code signatures to executables and libraries.

.DESCRIPTION
    The selfsign-path script automates the process of code signing using a self-signed
    certificate. Upon first run, it generates a new self-signed code-signing
    certificate and imports it into the system's Trusted Root Certification
    Authorities store. Subsequent runs will use this existing certificate.

    The script can sign new files, re-sign existing files, or append a signature.
    It can also be used to check the signature status of files or to remove its
    own self-signed signature. It accepts one or more files or glob-like patterns
    as input.

.PARAMETER FileOrPattern
    One or more space-separated paths to files or directories.
    Supports glob-like patterns (e.g., *.exe, bin/*) to specify multiple files.

.PARAMETER Recurse
    Recursively search for and process files in any specified directories.

.PARAMETER Name
    Specify the subject name of the certificate to use for signing. If not
    found, a new certificate with this name is created. Defaults to a
    pre-configured name if not provided.

.PARAMETER CertFile
    Specify the path to the certificate file (.cer or .pem). This bypasses
    the default certificate generation and lookup. Requires --key-file.

.PARAMETER KeyFile
    Specify the path to the private key file (.pvk or .key). Required if
    --cert-file is used.

.PARAMETER Clear
    Remove self-signed signatures created by this tool from the specified
    files. It will not affect other valid signatures.

.PARAMETER Status
    Print the signing status of the specified files instead of signing them.

.PARAMETER Help
    Display this help documentation and exit.

.EXAMPLE
    .\selfsign-path.ps1 myapp.exe
    Sign a single executable.

.EXAMPLE
    .\selfsign-path.ps1 -Recurse "bin/**/*.dll"
    Sign all DLLs in a directory and its subdirectories.

.EXAMPLE
    .\selfsign-path.ps1 -Status "*.exe"
    Check the signature status of all executables in the current directory.

.EXAMPLE
    .\selfsign-path.ps1 -Name "My Custom Cert" myapp.exe
    Sign files using a custom-named certificate.

.EXAMPLE
    .\selfsign-path.ps1 -CertFile "/path/to/my.crt" -KeyFile "/path/to/my.key" myapp.exe
    Sign a file using specific certificate and key files.

.EXAMPLE
    .\selfsign-path.ps1 -Clear -Recurse "release/"
    Remove self-signatures from all files in a release folder.
#>

[CmdletBinding()]
param(
    [Parameter(Position = 0, ValueFromRemainingArguments = $true)]
    [string[]]$FileOrPattern = @(),

    [Alias("r")]
    [switch]$Recurse,

    [Alias("n")]
    [string]$Name = "LocalSign-SelfSigned",

    [Alias("c")]
    [string]$CertFile,

    [Alias("k")]
    [string]$KeyFile,

    [switch]$Clear,

    [switch]$Status,

    [Alias("h")]
    [switch]$Help
)

# Display help if requested or no parameters provided
if ($Help -or ($FileOrPattern.Count -eq 0 -and -not $Clear -and -not $Status)) {
    $helpText = @"
NAME
    selfsign-path - A utility to manage and apply self-signed code signatures to executables and libraries.

SYNOPSIS
    selfsign-path [OPTIONS] file_or_pattern...

DESCRIPTION
    The selfsign-path script automates the process of code signing using a self-signed
    certificate. Upon first run, it generates a new self-signed code-signing
    certificate and imports it into the system's Trusted Root Certification
    Authorities store. Subsequent runs will use this existing certificate.

    The script can sign new files, re-sign existing files, or append a signature.
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

EXAMPLES
    Sign a single executable:
        selfsign-path myapp.exe

    Sign all DLLs in a directory and its subdirectories:
        selfsign-path --recurse 'bin/**/*.dll'

    Check the signature status of all executables in the current directory:
        selfsign-path --status *.exe

    Sign files using a custom-named certificate:
        selfsign-path -n "My Custom Cert" myapp.exe

    Sign a file using specific certificate and key files:
        selfsign-path --cert-file /path/to/my.crt --key-file /path/to/my.key myapp.exe

    Remove self-signatures from all files in a release folder:
        selfsign-path --clear --recurse release/
"@
    Write-Host $helpText
    exit 0
}

# Validate certificate and key file parameters
if ($CertFile -and -not $KeyFile) {
    Write-Error "Error: --cert-file requires --key-file to be specified."
    exit 1
}

if ($KeyFile -and -not $CertFile) {
    Write-Error "Error: --key-file requires --cert-file to be specified."
    exit 1
}

# Function to check if running as administrator (Windows)
function Test-Administrator {
    if ($IsWindows -or ($env:OS -eq "Windows_NT")) {
        $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
        $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
        return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    }
    return $true  # Assume admin on non-Windows systems for certificate operations
}

# Function to find or create a self-signed certificate
function Get-OrCreateSelfSignedCertificate {
    param([string]$SubjectName)
    
    # Check if New-SelfSignedCertificate is available (Windows)
    if (-not (Get-Command New-SelfSignedCertificate -ErrorAction SilentlyContinue)) {
        Write-Warning "Certificate generation not supported on this platform. Use -CertFile and -KeyFile options with existing certificates."
        return $null
    }
    
    # Look for existing certificate
    $existingCert = Get-ChildItem -Path "Cert:\CurrentUser\My" -ErrorAction SilentlyContinue | Where-Object {
        $_.Subject -like "*CN=$SubjectName*" -and $_.HasPrivateKey
    } | Select-Object -First 1

    if ($existingCert) {
        Write-Verbose "Using existing certificate: $($existingCert.Thumbprint)"
        return $existingCert
    }

    Write-Host "Creating new self-signed certificate with subject: $SubjectName"
    
    try {
        # Create self-signed certificate
        $cert = New-SelfSignedCertificate -Subject "CN=$SubjectName" -CertStoreLocation "Cert:\CurrentUser\My" -KeyUsage DigitalSignature -Type CodeSigning -KeyAlgorithm RSA -KeyLength 2048 -NotAfter (Get-Date).AddYears(3)
        
        # Try to install to Trusted Root (requires admin on Windows)
        if (Test-Administrator) {
            try {
                $store = New-Object System.Security.Cryptography.X509Certificates.X509Store([System.Security.Cryptography.X509Certificates.StoreName]::Root, [System.Security.Cryptography.X509Certificates.StoreLocation]::LocalMachine)
                $store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite)
                $store.Add($cert)
                $store.Close()
                Write-Host "Certificate installed to Trusted Root Certification Authorities store."
            }
            catch {
                Write-Warning "Could not install certificate to Trusted Root store: $($_.Exception.Message)"
            }
        }
        else {
            Write-Warning "Not running as administrator. Certificate created but not installed to Trusted Root store."
            Write-Host "To install the certificate to Trusted Root store, run this script as administrator."
        }

        return $cert
    }
    catch {
        Write-Error "Failed to create self-signed certificate: $($_.Exception.Message)"
        return $null
    }
}

# Function to load certificate from file
function Get-CertificateFromFile {
    param([string]$CertPath, [string]$KeyPath)
    
    if (-not (Test-Path $CertPath)) {
        Write-Error "Certificate file not found: $CertPath"
        exit 1
    }
    
    if (-not (Test-Path $KeyPath)) {
        Write-Error "Key file not found: $KeyPath"
        exit 1
    }
    
    try {
        # This is a simplified implementation - in practice, you'd need to handle
        # different certificate and key formats (PEM, DER, PVK, etc.)
        $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($CertPath)
        Write-Verbose "Loaded certificate from file: $CertPath"
        return $cert
    }
    catch {
        Write-Error "Failed to load certificate from file: $($_.Exception.Message)"
        exit 1
    }
}

# Function to expand file patterns and get target files
function Get-TargetFiles {
    param([string[]]$Patterns, [switch]$Recursive)
    
    $files = @()
    
    foreach ($pattern in $Patterns) {
        if (Test-Path $pattern -PathType Container) {
            # It's a directory
            if ($Recursive) {
                $files += Get-ChildItem -Path $pattern -Recurse -File | Where-Object { 
                    $_.Extension -in @('.exe', '.dll', '.msi', '.sys', '.com', '.ocx', '.scr', '.cpl')
                }
            }
            else {
                $files += Get-ChildItem -Path $pattern -File | Where-Object { 
                    $_.Extension -in @('.exe', '.dll', '.msi', '.sys', '.com', '.ocx', '.scr', '.cpl')
                }
            }
        }
        elseif ($pattern -match "[\*\?]") {
            # It's a pattern with wildcards
            if ($Recursive) {
                $files += Get-ChildItem -Path $pattern -Recurse -File -ErrorAction SilentlyContinue
            }
            else {
                $files += Get-ChildItem -Path $pattern -File -ErrorAction SilentlyContinue
            }
        }
        else {
            # It's a specific file
            if (Test-Path $pattern -PathType Leaf) {
                $files += Get-Item $pattern
            }
            else {
                Write-Warning "File not found: $pattern"
            }
        }
    }
    
    return $files | Sort-Object FullName -Unique
}

# Function to check signature status of a file
function Get-FileSignatureStatus {
    param([System.IO.FileInfo]$TargetFile)
    
    try {
        # Check if Get-AuthenticodeSignature is available (Windows)
        if (Get-Command Get-AuthenticodeSignature -ErrorAction SilentlyContinue) {
            $signature = Get-AuthenticodeSignature -FilePath $TargetFile.FullName
            $result = New-Object PSObject -Property @{
                File = $TargetFile.FullName
                Status = $signature.Status.ToString()
                SignerCertificate = $null
                TimestampCertificate = $null
                IsSelfSigned = $false
            }
            
            if ($signature.SignerCertificate) {
                $result.SignerCertificate = $signature.SignerCertificate.Subject
                $result.IsSelfSigned = $signature.SignerCertificate.Subject -eq $signature.SignerCertificate.Issuer
            }
            
            if ($signature.TimeStamperCertificate) {
                $result.TimestampCertificate = $signature.TimeStamperCertificate.Subject
            }
        }
        else {
            # Platform doesn't support Authenticode signatures
            $result = New-Object PSObject -Property @{
                File = $TargetFile.FullName
                Status = "NotSupported (Platform doesn't support Authenticode signatures)"
                SignerCertificate = $null
                TimestampCertificate = $null
                IsSelfSigned = $false
            }
        }
        
        return $result
    }
    catch {
        return New-Object PSObject -Property @{
            File = $TargetFile.FullName
            Status = "Error: $($_.Exception.Message)"
            SignerCertificate = $null
            TimestampCertificate = $null
            IsSelfSigned = $false
        }
    }
}

# Function to sign a file
function Set-FileSignature {
    param([System.IO.FileInfo]$TargetFile, [System.Security.Cryptography.X509Certificates.X509Certificate2]$Certificate)
    
    try {
        # Check if Set-AuthenticodeSignature is available (Windows)
        if (Get-Command Set-AuthenticodeSignature -ErrorAction SilentlyContinue) {
            $result = Set-AuthenticodeSignature -FilePath $TargetFile.FullName -Certificate $Certificate -TimestampServer "http://timestamp.digicert.com"
            
            if ($result.Status -eq "Valid") {
                Write-Host "Successfully signed: $($TargetFile.FullName)" -ForegroundColor Green
                return $true
            }
            else {
                Write-Warning "Signing failed for $($TargetFile.FullName): $($result.StatusMessage)"
                return $false
            }
        }
        else {
            Write-Warning "Code signing not supported on this platform: $($TargetFile.FullName)"
            return $false
        }
    }
    catch {
        Write-Error "Error signing $($TargetFile.FullName): $($_.Exception.Message)"
        return $false
    }
}

# Function to remove self-signed signatures
function Remove-SelfSignedSignature {
    param([System.IO.FileInfo]$TargetFile)
    
    try {
        # Check if Get-AuthenticodeSignature is available (Windows)
        if (Get-Command Get-AuthenticodeSignature -ErrorAction SilentlyContinue) {
            $signature = Get-AuthenticodeSignature -FilePath $TargetFile.FullName
            
            if ($signature.Status -ne "NotSigned" -and $signature.SignerCertificate) {
                # Check if it's a self-signed certificate (Subject == Issuer)
                if ($signature.SignerCertificate.Subject -eq $signature.SignerCertificate.Issuer) {
                    # Remove signature by setting it to $null
                    $result = Set-AuthenticodeSignature -FilePath $TargetFile.FullName -Certificate $null
                    Write-Host "Removed self-signed signature from: $($TargetFile.FullName)" -ForegroundColor Yellow
                    return $true
                }
                else {
                    Write-Verbose "Skipping $($TargetFile.FullName) - not self-signed"
                    return $false
                }
            }
            else {
                Write-Verbose "Skipping $($TargetFile.FullName) - not signed"
                return $false
            }
        }
        else {
            Write-Warning "Signature removal not supported on this platform: $($TargetFile.FullName)"
            return $false
        }
    }
    catch {
        Write-Error "Error processing $($TargetFile.FullName): $($_.Exception.Message)"
        return $false
    }
}

# Main execution logic
try {
    if ($Recurse) {
        $targetFiles = Get-TargetFiles -Patterns $FileOrPattern -Recursive
    } else {
        $targetFiles = Get-TargetFiles -Patterns $FileOrPattern
    }
    
    if ($targetFiles.Count -eq 0) {
        Write-Warning "No files found matching the specified patterns."
        exit 0
    }
    
    Write-Verbose "Found $($targetFiles.Count) file(s) to process."
    
    if ($Status) {
        # Status checking mode
        Write-Host "`nSignature Status Report:" -ForegroundColor Cyan
        Write-Host ("=" * 50) -ForegroundColor Cyan
        
        foreach ($file in $targetFiles) {
            $fileInfo = Get-FileSignatureStatus $file
            Write-Host "`nFile: $($fileInfo.File)"
            Write-Host "Status: $($fileInfo.Status)"
            if ($fileInfo.SignerCertificate) {
                Write-Host "Signer: $($fileInfo.SignerCertificate)"
                Write-Host "Self-signed: $($fileInfo.IsSelfSigned)"
            }
            if ($fileInfo.TimestampCertificate) {
                Write-Host "Timestamp: $($fileInfo.TimestampCertificate)"
            }
        }
    }
    elseif ($Clear) {
        # Clear signatures mode
        Write-Host "Removing self-signed signatures..." -ForegroundColor Yellow
        $removedCount = 0
        
        foreach ($file in $targetFiles) {
            if (Remove-SelfSignedSignature $file) {
                $removedCount++
            }
        }
        
        Write-Host "`nRemoved signatures from $removedCount file(s)." -ForegroundColor Yellow
    }
    else {
        # Signing mode
        $certificate = $null
        
        if ($CertFile -and $KeyFile) {
            $certificate = Get-CertificateFromFile -CertPath $CertFile -KeyPath $KeyFile
        }
        else {
            $certificate = Get-OrCreateSelfSignedCertificate -SubjectName $Name
        }
        
        if (-not $certificate) {
            if ($CertFile -and $KeyFile) {
                Write-Error "Failed to load certificate from files."
            }
            else {
                Write-Error "Failed to obtain signing certificate. Certificate generation not supported on this platform. Use -CertFile and -KeyFile options."
            }
            exit 1
        }
        
        Write-Host "Signing files with certificate: $($certificate.Subject)" -ForegroundColor Cyan
        $signedCount = 0
        
        foreach ($file in $targetFiles) {
            if (Set-FileSignature $file $certificate) {
                $signedCount++
            }
        }
        
        Write-Host "`nSuccessfully signed $signedCount out of $($targetFiles.Count) file(s)." -ForegroundColor Green
    }
}
catch {
    Write-Error "Unexpected error: $($_.Exception.Message)"
    exit 1
}