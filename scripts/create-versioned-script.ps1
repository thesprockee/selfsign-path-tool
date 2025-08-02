# create-versioned-script.ps1 - Create versioned PowerShell script

param (
    [string]$Version = $env:VERSION
)
# If VERSION is not provided, try to get the latest git tag
if (-not $Version) {
    # Try to get the latest semantic version tag from git
    $LatestTag = git tag --sort=-version:refname | Where-Object { $_ -match '^v?\d+\.\d+\.\d+' } | Select-Object -First 1
    
    if ($LatestTag) {
        # Remove 'v' prefix if present
        $Version = $LatestTag -replace '^v', ''
        Write-Host "No VERSION provided, using latest git tag: $LatestTag (version: $Version)"
    } else {
        Write-Error "❌ Error: VERSION must be provided as argument or environment variable, and no valid git tags found"
        exit 1
    }
}

$VersionedScript = "selfsign-path-v$Version.ps1"

Write-Host "Creating versioned script: $VersionedScript"

if (Test-Path "selfsign-path.ps1") {
    Copy-Item "selfsign-path.ps1" $VersionedScript -Force
    Write-Host "✅ Created versioned script: $VersionedScript"

    # Set environment variable for subsequent steps (GitHub Actions style)
    if ($env:GITHUB_ENV) {
        Add-Content -Path $env:GITHUB_ENV -Value "VERSIONED_SCRIPT=$VersionedScript"
    }
} else {
    Write-Error "❌ Error: selfsign-path.ps1 not found"
    exit 1
}
