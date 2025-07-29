#!/usr/bin/env pwsh

<#
.SYNOPSIS
    Express installation script for local-sign tool - signs Oculus VR applications.

.DESCRIPTION
    This script provides a one-liner installation experience for the local-sign tool.
    It automatically downloads the sign-tool.ps1 script from GitHub and performs
    the following operations:
    
    1. Prompts to delete existing certificates created by this script
    2. Creates a new signing certificate with a deterministic description
    3. Imports the certificate to the trusted CA in the certificate store
    4. Signs all executables and DLLs in the Oculus VR directories:
       - C:\Program Files\Oculus\Software\Software\ready-at-dawn-echo-arena
       - C:\echovr

.PARAMETER Force
    Skip confirmation prompts and proceed with installation automatically.

.PARAMETER CertName
    Override the default certificate name. Default is "LocalSign-OculusVR".

.EXAMPLE
    Install via one-liner (recommended):
    $tempFile = New-TemporaryFile; iwr -useb https://raw.githubusercontent.com/thesprockee/local-sign/main/install.ps1 -OutFile $tempFile; pwsh -File $tempFile; Remove-Item $tempFile

.EXAMPLE
    Install with custom certificate name:
    $tempFile = New-TemporaryFile; iwr -useb https://raw.githubusercontent.com/thesprockee/local-sign/main/install.ps1 -OutFile $tempFile; pwsh -File $tempFile -CertName 'MyCustomCert'; Remove-Item $tempFile

.EXAMPLE
    Force install without prompts:
    $tempFile = New-TemporaryFile; iwr -useb https://raw.githubusercontent.com/thesprockee/local-sign/main/install.ps1 -OutFile $tempFile; pwsh -File $tempFile -Force; Remove-Item $tempFile
#>

[CmdletBinding()]
param(
    [switch]$Force,
    [string]$CertName = "LocalSign-OculusVR"
)

function Install-LocalSign {
    [CmdletBinding()]
    param(
        [switch]$Force,
        [string]$CertName = "LocalSign-OculusVR"
    )

    # Script configuration
    $SignToolUrl = "https://raw.githubusercontent.com/thesprockee/local-sign/main/sign-tool.ps1"
    $OculusDirectories = @(
        "C:\Program Files\Oculus\Software\Software\ready-at-dawn-echo-arena",
        "C:\echovr"
    )

    Write-Host "=== Local-Sign Express Installation ===" -ForegroundColor Cyan
    Write-Host "This script will automatically sign Oculus VR applications." -ForegroundColor Yellow
    Write-Host ""

    # Check if running as administrator
    if (-not (Test-Administrator)) {
        Write-Error "This script must be run as Administrator to install certificates and sign system files."
        Write-Host "Please restart PowerShell as Administrator and run the command again." -ForegroundColor Red
        exit 1
    }

    # Step 1: Download sign-tool.ps1
    Write-Host "Step 1: Downloading sign-tool.ps1 from GitHub..." -ForegroundColor Green
    try {
        $tempSignTool = Join-Path $env:TEMP "sign-tool.ps1"
        Invoke-WebRequest -Uri $SignToolUrl -OutFile $tempSignTool -UseBasicParsing
        Write-Host "✓ Downloaded sign-tool.ps1 successfully" -ForegroundColor Green
    }
    catch {
        Write-Error "Failed to download sign-tool.ps1: $($_.Exception.Message)"
        exit 1
    }

    # Step 2: Handle existing certificates
    Write-Host "`nStep 2: Checking for existing certificates..." -ForegroundColor Green
    $existingCerts = Get-ChildItem -Path "Cert:\CurrentUser\My" -ErrorAction SilentlyContinue | Where-Object {
        $_.Subject -like "*CN=LocalSign*" -and $_.HasPrivateKey
    }

    if ($existingCerts.Count -gt 0) {
        Write-Host "Found $($existingCerts.Count) existing LocalSign certificate(s):" -ForegroundColor Yellow
        foreach ($cert in $existingCerts) {
            Write-Host "  - $($cert.Subject) (Thumbprint: $($cert.Thumbprint))" -ForegroundColor Yellow
        }
        
        if (-not $Force) {
            $response = Read-Host "`nDo you want to delete these existing certificates? [Y/n]"
            if ($response -eq "" -or $response -eq "Y" -or $response -eq "y") {
                $deleteCerts = $true
            }
            else {
                $deleteCerts = $false
                Write-Host "Keeping existing certificates. New certificate will be created alongside them." -ForegroundColor Yellow
            }
        }
        else {
            $deleteCerts = $true
            Write-Host "Force mode: Deleting existing certificates..." -ForegroundColor Yellow
        }

        if ($deleteCerts) {
            foreach ($cert in $existingCerts) {
                try {
                    # Remove from personal store
                    $personalStore = New-Object System.Security.Cryptography.X509Certificates.X509Store([System.Security.Cryptography.X509Certificates.StoreName]::My, [System.Security.Cryptography.X509Certificates.StoreLocation]::CurrentUser)
                    $personalStore.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite)
                    $personalStore.Remove($cert)
                    $personalStore.Close()

                    # Try to remove from trusted root store
                    try {
                        $rootStore = New-Object System.Security.Cryptography.X509Certificates.X509Store([System.Security.Cryptography.X509Certificates.StoreName]::Root, [System.Security.Cryptography.X509Certificates.StoreLocation]::LocalMachine)
                        $rootStore.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite)
                        $rootStore.Remove($cert)
                        $rootStore.Close()
                    }
                    catch {
                        # Ignore errors removing from root store (might not be there)
                    }

                    Write-Host "✓ Deleted certificate: $($cert.Subject)" -ForegroundColor Green
                }
                catch {
                    Write-Warning "Failed to delete certificate $($cert.Subject): $($_.Exception.Message)"
                }
            }
        }
    }
    else {
        Write-Host "✓ No existing LocalSign certificates found" -ForegroundColor Green
    }

    # Step 3: Create new certificate
    Write-Host "`nStep 3: Creating new signing certificate..." -ForegroundColor Green
    try {
        # Create a temporary dummy file to satisfy the signing tool's requirement for a file.
        $dummyFile = [System.IO.Path]::GetTempFileName()
        try {
            & $tempSignTool -Name $CertName $dummyFile 2>$null >$null
            Write-Host "✓ Created and installed certificate: $CertName" -ForegroundColor Green
        }
        finally {
            # Clean up the temporary dummy file.
            Remove-Item -Path $dummyFile -Force -ErrorAction SilentlyContinue
        }
    }
    catch {
        Write-Error "Failed to create certificate: $($_.Exception.Message)"
        exit 1
    }

    # Step 4: Sign Oculus directories
    Write-Host "`nStep 4: Signing Oculus VR applications..." -ForegroundColor Green
    $totalSigned = 0
    $totalFound = 0

    foreach ($directory in $OculusDirectories) {
        Write-Host "`nProcessing directory: $directory" -ForegroundColor Cyan
        
        if (-not (Test-Path $directory)) {
            Write-Host "  ⚠ Directory not found, skipping: $directory" -ForegroundColor Yellow
            continue
        }

        try {
            # Get all exe and dll files recursively
            $files = Get-ChildItem -Path $directory -Recurse -File | Where-Object { 
                $_.Extension -in @('.exe', '.dll') 
            }

            if ($files.Count -eq 0) {
                Write-Host "  ⚠ No .exe or .dll files found in: $directory" -ForegroundColor Yellow
                continue
            }

            Write-Host "  Found $($files.Count) files to sign..." -ForegroundColor Cyan
            $totalFound += $files.Count

            # Sign files using the sign-tool
            $signResult = & $tempSignTool -Name $CertName -Recurse $directory 2>&1
            
            # Count successful signatures from output
            $signedInDir = ($signResult | Select-String "Successfully signed:" | Measure-Object).Count
            $totalSigned += $signedInDir
            
            Write-Host "  ✓ Signed $signedInDir files in $directory" -ForegroundColor Green
        }
        catch {
            Write-Warning "Error processing directory $directory : $($_.Exception.Message)"
        }
    }

    # Step 5: Cleanup and summary
    Write-Host "`nStep 5: Cleanup..." -ForegroundColor Green
    try {
        Remove-Item $tempSignTool -Force
        Write-Host "✓ Cleaned up temporary files" -ForegroundColor Green
    }
    catch {
        Write-Warning "Could not cleanup temporary file: $tempSignTool"
    }

    # Final summary
    Write-Host "`n=== Installation Complete ===" -ForegroundColor Cyan
    Write-Host "✓ Certificate created and installed: $CertName" -ForegroundColor Green
    Write-Host "✓ Total files processed: $totalFound" -ForegroundColor Green
    Write-Host "✓ Total files signed: $totalSigned" -ForegroundColor Green
    
    if ($totalSigned -lt $totalFound) {
        Write-Host "⚠ Some files could not be signed (possibly already signed or in use)" -ForegroundColor Yellow
    }

    Write-Host "`nYour Oculus VR applications are now signed and should work without security warnings." -ForegroundColor Cyan
    Write-Host "If you encounter any issues, try restarting the Oculus software." -ForegroundColor Yellow
}

# Function to check if running as administrator (Windows)
function Test-Administrator {
    if ($IsWindows -or ($env:OS -eq "Windows_NT")) {
        $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
        $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
        return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    }
    return $true  # Assume admin on non-Windows systems
}

# Auto-execute if script is run directly (not dot-sourced)
if ($MyInvocation.InvocationName -ne '.') {
    Install-LocalSign -Force:$Force -CertName $CertName
}