#!/usr/bin/env pwsh

<#
.SYNOPSIS
    Helper script to create new releases for the selfsign-path-tool project
.DESCRIPTION
    This script simplifies the process of creating new releases by automating
    version tagging and pushing to trigger the automated release workflow.
.PARAMETER Version
    The semantic version to release (e.g., 1.0.0, 2.1.3)
.PARAMETER Message
    Optional tag message. If not provided, a default message will be used.
.PARAMETER DryRun
    Show what would be done without actually creating the tag or pushing
.EXAMPLE
    .\create-release.ps1 1.0.0
    Create and push a v1.0.0 release tag
.EXAMPLE
    .\create-release.ps1 1.0.0 -Message "First stable release"
    Create a v1.0.0 tag with a custom message
.EXAMPLE
    .\create-release.ps1 2.1.0 -DryRun
    Show what would happen without actually creating the release
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$Version,
    
    [Parameter()]
    [string]$Message,
    
    [switch]$DryRun
)

# Validate version format (basic semantic versioning)
if ($Version -notmatch '^\d+\.\d+\.\d+(-[\w-]+(\.[\w-]+)*)?$') {
    Write-Error "Invalid version format. Please use semantic versioning (e.g., 1.0.0, 2.1.3, 1.0.0-beta.1)"
    exit 1
}

$tagName = "v$Version"

# Check if we're in a git repository
try {
    $null = git rev-parse --git-dir 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Not a git repository"
    }
} catch {
    Write-Error "This script must be run from within the selfsign-path-tool git repository"
    exit 1
}

# Check if tag already exists
$existingTag = git tag -l $tagName 2>$null
if ($existingTag) {
    Write-Error "Tag $tagName already exists. Use a different version number."
    exit 1
}

# Check for uncommitted changes
$status = git status --porcelain 2>$null
if ($status) {
    Write-Warning "There are uncommitted changes in the repository:"
    git status --short
    Write-Host ""
    $continue = Read-Host "Continue anyway? (y/N)"
    if ($continue -ne 'y' -and $continue -ne 'Y') {
        Write-Host "Release creation cancelled."
        exit 0
    }
}

# Default message if not provided
if (-not $Message) {
    $Message = "Release version $Version"
}

Write-Host "🏷️ Creating Release $tagName" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Version: $Version"
Write-Host "Tag: $tagName"
Write-Host "Message: $Message"
Write-Host ""

if ($DryRun) {
    Write-Host "🔍 DRY RUN - No changes will be made" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Would execute:"
    Write-Host "  git tag -a $tagName -m `"$Message`""
    Write-Host "  git push origin $tagName"
    Write-Host ""
    Write-Host "This would trigger the automated release workflow which will:"
    Write-Host "  • Create a draft release for $tagName"
    Write-Host "  • Generate selfsign-path-v$Version.ps1"
    Write-Host "  • Sign the script (if certificates are configured)"
    Write-Host "  • Create changelog from git history"
    Write-Host "  • Attach the script to the release"
    exit 0
}

# Create the tag
Write-Host "📝 Creating tag..." -ForegroundColor Green
try {
    git tag -a $tagName -m $Message
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create tag"
    }
    Write-Host "✅ Tag $tagName created successfully"
} catch {
    Write-Error "Failed to create tag: $($_.Exception.Message)"
    exit 1
}

# Push the tag
Write-Host "🚀 Pushing tag to trigger release workflow..." -ForegroundColor Green
try {
    git push origin $tagName
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to push tag"
    }
    Write-Host "✅ Tag pushed successfully"
} catch {
    Write-Error "Failed to push tag: $($_.Exception.Message)"
    Write-Host "You can manually push the tag later with: git push origin $tagName"
    exit 1
}

Write-Host ""
Write-Host "🎉 Release creation initiated!" -ForegroundColor Green
Write-Host ""
Write-Host "The automated release workflow is now running. It will:"
Write-Host "  • Create a draft release for $tagName"
Write-Host "  • Generate selfsign-path-v$Version.ps1"
Write-Host "  • Sign the script (if certificates are configured)"
Write-Host "  • Create changelog from git history"
Write-Host "  • Attach the script to the release"
Write-Host ""
Write-Host "You can monitor the workflow progress at:"
Write-Host "https://github.com/thesprockee/selfsign-path-tool/actions"
Write-Host ""
Write-Host "Once complete, review and publish the draft release at:"
Write-Host "https://github.com/thesprockee/selfsign-path-tool/releases"