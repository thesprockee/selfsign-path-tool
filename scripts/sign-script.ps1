# sign-script.ps1 - Sign PowerShell script if certificates are available

param (
    [string]$VersionedScript = $env:VERSIONED_SCRIPT
)

if (-not $VersionedScript) {
    Write-Error "❌ Error: VERSIONED_SCRIPT must be provided as argument or environment variable"
    exit 1
}

$signingCert = $env:SIGNING_CERT
$signingCertPassword = $env:SIGNING_CERT_PASSWORD

if ($signingCert -and $signingCertPassword) {
    Write-Host "🔐 Signing certificate found, proceeding with script signing..."

    # Decode base64 certificate to a secure temporary file
    $certTempFile = [System.IO.Path]::GetTempFileName()
    [System.IO.File]::WriteAllBytes($certTempFile, [Convert]::FromBase64String($signingCert))

    try {
        # Load certificate
        $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($certTempFile, $signingCertPassword)

        # Sign the script
        $result = Set-AuthenticodeSignature -FilePath $VersionedScript -Certificate $cert -TimestampServer 'https://timestamp.digicert.com'

        if ($result.Status -eq 'Valid') {
            Write-Host "✅ Successfully signed $VersionedScript" -ForegroundColor Green
            if ($env:GITHUB_ENV) {
                Add-Content -Path $env:GITHUB_ENV -Value 'SCRIPT_SIGNED=true'
            }
        } else {
            Write-Warning "Signing failed: $($result.StatusMessage)"
            if ($env:GITHUB_ENV) {
                Add-Content -Path $env:GITHUB_ENV -Value 'SCRIPT_SIGNED=false'
            }
        }
    }
    catch {
        Write-Error "Error during signing: $($_.Exception.Message)"
        if ($env:GITHUB_ENV) {
            Add-Content -Path $env:GITHUB_ENV -Value 'SCRIPT_SIGNED=false'
        }
    }
    finally {
        # Clean up certificate file
        if (Test-Path $certTempFile) {
            Remove-Item $certTempFile -Force
        }
    }
} else {
    Write-Warning "⚠️ No signing certificate configured. Skipping script signing."
    Write-Host "To enable script signing, add SIGNING_CERT (base64 encoded .pfx) and SIGNING_CERT_PASSWORD secrets to your repository."
    if ($env:GITHUB_ENV) {
        Add-Content -Path $env:GITHUB_ENV -Value 'SCRIPT_SIGNED=false'
    }
}
