#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Windows-compatible build script for selfsign-path-tool
.DESCRIPTION
    This script provides a Windows-compatible build system using PowerShell.
    It can be used as an alternative to the CMake build system.
.PARAMETER Target
    The build target to execute. Valid values: dist, clean, help
.PARAMETER Version
    The version number for the release (required for dist target)
.PARAMETER TagName
    The git tag name for the release (defaults to v{Version})
.EXAMPLE
    .\build.ps1 -Target dist -Version "1.0.0"
.EXAMPLE
    .\build.ps1 -Target help
.EXAMPLE
    .\build.ps1 -Target clean
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("dist", "clean", "help")]
    [string]$Target,
    
    [Parameter(Mandatory=$false)]
    [string]$Version = $env:VERSION,
    
    [Parameter(Mandatory=$false)]
    [string]$TagName = $env:TAG_NAME
)

$ErrorActionPreference = 'Stop'

function Show-Help {
    Write-Host "=== selfsign-path-tool Build Script ===" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\build.ps1 -Target <target> [-Version <version>] [-TagName <tag>]" -ForegroundColor White
    Write-Host ""
    Write-Host "Targets:" -ForegroundColor Yellow
    Write-Host "  dist    - Create all release artifacts (requires -Version)" -ForegroundColor White
    Write-Host "  clean   - Clean generated files" -ForegroundColor White
    Write-Host "  help    - Show this help message" -ForegroundColor White
    Write-Host ""
    Write-Host "Parameters:" -ForegroundColor Yellow
    Write-Host "  -Version    - Version number (e.g., '1.0.0')" -ForegroundColor White
    Write-Host "  -TagName    - Git tag name (defaults to 'v{Version}')" -ForegroundColor White
    Write-Host ""
    Write-Host "Environment Variables (optional):" -ForegroundColor Yellow
    Write-Host "  VERSION              - Version number" -ForegroundColor White
    Write-Host "  TAG_NAME             - Git tag name" -ForegroundColor White
    Write-Host "  SIGNING_CERT         - Base64 encoded signing certificate" -ForegroundColor White
    Write-Host "  SIGNING_CERT_PASSWORD - Password for signing certificate" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 -Target dist -Version '1.0.0'" -ForegroundColor White
    Write-Host "  .\build.ps1 -Target clean" -ForegroundColor White
    Write-Host "  `$env:VERSION='1.0.0'; .\build.ps1 -Target dist" -ForegroundColor White
}

function Invoke-Clean {
    Write-Host "🧹 Cleaning generated files..." -ForegroundColor Yellow
    
    $filesToRemove = @(
        "selfsign-path-v*.ps1",
        "RELEASE_NOTES.md"
    )
    
    $dirsToRemove = @(
        "dist",
        "build"
    )
    
    foreach ($pattern in $filesToRemove) {
        $files = Get-ChildItem -Path $pattern -ErrorAction SilentlyContinue
        foreach ($file in $files) {
            Remove-Item $file.FullName -Force
            Write-Host "  Removed: $($file.Name)" -ForegroundColor Gray
        }
    }
    
    foreach ($dir in $dirsToRemove) {
        if (Test-Path $dir) {
            Remove-Item $dir -Recurse -Force
            Write-Host "  Removed directory: $dir" -ForegroundColor Gray
        }
    }
    
    Write-Host "✅ Clean completed" -ForegroundColor Green
}

function Invoke-Dist {
    if (-not $Version) {
        Write-Error "❌ Version is required for dist target. Use -Version parameter or set VERSION environment variable."
        exit 1
    }
    
    if (-not $TagName) {
        $TagName = "v$Version"
    }
    
    Write-Host "📦 Creating distribution for version $Version (tag: $TagName)..." -ForegroundColor Cyan
    
    # Set environment variables for scripts
    $env:VERSION = $Version
    $env:TAG_NAME = $TagName
    
    try {
        # Step 1: Create versioned script
        Write-Host "  🔧 Creating versioned script..." -ForegroundColor Yellow
        & ".\scripts\create-versioned-script.ps1" -Version $Version
        if ($LASTEXITCODE -and $LASTEXITCODE -ne 0) {
            throw "Failed to create versioned script"
        }
        
        # Step 2: Generate changelog
        Write-Host "  📝 Generating changelog..." -ForegroundColor Yellow
        & ".\scripts\generate-changelog.ps1" -Version $Version -TagName $TagName
        if ($LASTEXITCODE -and $LASTEXITCODE -ne 0) {
            throw "Failed to generate changelog"
        }
        
        # Step 3: Sign script (optional)
        Write-Host "  🔐 Signing script..." -ForegroundColor Yellow
        & ".\scripts\sign-script.ps1" -VersionedScript "selfsign-path-v$Version.ps1"
        # Note: sign-script.ps1 doesn't exit with error if signing fails (it's optional)
        
        Write-Host "✅ Distribution created successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Generated files:" -ForegroundColor Yellow
        if (Test-Path "selfsign-path-v$Version.ps1") {
            Write-Host "  📄 selfsign-path-v$Version.ps1" -ForegroundColor White
        }
        if (Test-Path "RELEASE_NOTES.md") {
            Write-Host "  📄 RELEASE_NOTES.md" -ForegroundColor White
        }
        
    } catch {
        Write-Error "❌ Distribution build failed: $_"
        exit 1
    }
}

# Main execution
Write-Host "selfsign-path-tool Build Script" -ForegroundColor Cyan
Write-Host "Target: $Target" -ForegroundColor Gray

switch ($Target) {
    "help" { Show-Help }
    "clean" { Invoke-Clean }
    "dist" { Invoke-Dist }
    default { 
        Write-Error "Unknown target: $Target"
        Show-Help
        exit 1
    }
}