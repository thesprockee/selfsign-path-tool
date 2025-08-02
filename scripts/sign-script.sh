#!/bin/bash
# sign-script.sh - Sign PowerShell script if certificates are available

set -e

VERSIONED_SCRIPT="${1:-${VERSIONED_SCRIPT}}"

if [ -z "$VERSIONED_SCRIPT" ]; then
    echo "❌ Error: VERSIONED_SCRIPT must be provided as argument or environment variable"
    exit 1
fi

# Check if signing certificate secrets are available
if [ -n "$SIGNING_CERT" ] && [ -n "$SIGNING_CERT_PASSWORD" ]; then
    echo "🔐 Signing certificate found, proceeding with script signing..."
    
    # Use PowerShell to handle the signing since it requires Windows-specific APIs
    powershell.exe -Command "
        \$versionedScript = '$VERSIONED_SCRIPT'
        \$env:SIGNING_CERT = '$SIGNING_CERT'
        \$env:SIGNING_CERT_PASSWORD = '$SIGNING_CERT_PASSWORD'
        
        # Decode base64 certificate
        \$certBytes = [System.Convert]::FromBase64String(\$env:SIGNING_CERT)
        \$certPath = 'signing-cert.pfx'
        [System.IO.File]::WriteAllBytes(\$certPath, \$certBytes)
        
        try {
            # Load certificate
    # Create a secure temporary file for the certificate
    CERT_TEMP_FILE=$(mktemp)
    chmod 600 "$CERT_TEMP_FILE"
    echo "$SIGNING_CERT" | base64 -d > "$CERT_TEMP_FILE"
    
    # Use PowerShell to handle the signing since it requires Windows-specific APIs
    powershell.exe -Command "
        \$versionedScript = '$VERSIONED_SCRIPT'
        \$certPath = '$CERT_TEMP_FILE'
        \$certPassword = '$SIGNING_CERT_PASSWORD'
        try {
            # Load certificate
            \$cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2(\$certPath, \$certPassword)
            
            # Sign the script
            \$result = Set-AuthenticodeSignature -FilePath \$versionedScript -Certificate \$cert -TimestampServer 'https://timestamp.digicert.com'
            
            if (\$result.Status -eq 'Valid') {
                Write-Host '✅ Successfully signed \$versionedScript' -ForegroundColor Green
                if (\$env:GITHUB_ENV) {
                    Add-Content -Path \$env:GITHUB_ENV -Value 'SCRIPT_SIGNED=true'
                }
            } else {
                Write-Warning 'Signing failed: \$(\$result.StatusMessage)'
                if (\$env:GITHUB_ENV) {
                    Add-Content -Path \$env:GITHUB_ENV -Value 'SCRIPT_SIGNED=false'
                }
            }
        }
        catch {
            Write-Error 'Error during signing: \$(\$_.Exception.Message)'
            if (\$env:GITHUB_ENV) {
                Add-Content -Path \$env:GITHUB_ENV -Value 'SCRIPT_SIGNED=false'
            }
        }
        finally {
            # Clean up certificate file
            if (Test-Path \$certPath) {
                Remove-Item \$certPath -Force
            }
        }
    "
else
    echo "⚠️ No signing certificate configured. Skipping script signing."
    echo "To enable script signing, add SIGNING_CERT (base64 encoded .pfx) and SIGNING_CERT_PASSWORD secrets to your repository."
    
    # Set environment variable for subsequent steps (GitHub Actions style)
    if [ -n "$GITHUB_ENV" ]; then
        echo "SCRIPT_SIGNED=false" >> "$GITHUB_ENV"
    fi
fi