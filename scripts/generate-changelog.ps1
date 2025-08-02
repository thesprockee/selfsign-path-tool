<#
.SYNOPSIS
    Generates a changelog and release notes from git history.
.DESCRIPTION
    This script creates a RELEASE_NOTES.md file by comparing the latest git tag
    with the previous one. It populates the file with commit messages from that range.
.PARAMETER Version
    The version number for the release (e.g., "1.2.0"). This is mandatory and can
    also be supplied via the $env:VERSION environment variable.
.PARAMETER TagName
    The git tag name for the release (e.g., "v1.2.0"). If not provided, it will be
    prefixed with "v" from the Version. Can also be supplied via $env:TAG_NAME.
.EXAMPLE
    .\Generate-Changelog.ps1 -Version "1.2.0"
.EXAMPLE
    $env:VERSION="1.2.0"; $env:TAG_NAME="v1.2.0"; .\Generate-Changelog.ps1
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$Version = $env:VERSION,

    [Parameter(Position=1)]
    [string]$TagName = $env:TAG_NAME
)

# Equivalent of 'set -e' in Bash. Stops the script on any terminating error.
$ErrorActionPreference = 'Stop'

# If TagName is not provided, derive it from Version.
if ([string]::IsNullOrEmpty($TagName)) {
    $TagName = "v$Version"
}

Write-Host "Generating changelog for version $Version (tag: $TagName)..."

# Get the previous tag. `Select-Object -Index 1` is the PowerShell equivalent of `sed -n '2p'`.
$PreviousTag = (git tag --sort=-version:refname | Select-Object -Index 1)

if ([string]::IsNullOrEmpty($PreviousTag)) {
    # If no previous tag, use the first commit in the repository's history.
    Write-Host "No previous tag found. Using the first commit as a baseline."
    $PreviousTag = (git rev-list --max-parents=0 HEAD)
}

Write-Host "Previous tag/commit: $PreviousTag"
Write-Host "Current tag: $TagName"

# Generate changelog from git log.
# `-ErrorAction SilentlyContinue` mimics bash's `2>/dev/null || true` to prevent errors if the commit range is empty.
$GitLog = git log --pretty=format:"- %s (%h)" "${PreviousTag}..HEAD" -ErrorAction SilentlyContinue

# Join the array of lines from git log into a single multi-line string.
$GitLogContent = $GitLog -join [System.Environment]::NewLine

# Determine if the script is signed by checking an environment variable.
$IsScriptSigned = $env:SCRIPT_SIGNED -eq 'true'

# Create the release notes content. A PowerShell Here-String (@"..."@) is used for multi-line content.
if (-not [string]::IsNullOrEmpty($GitLogContent)) {
    # If there are changes, create the full changelog.
    $ReleaseNotes = @"
## Changes in $TagName

$GitLogContent

## Installation

Download the attached `selfsign-path-v$Version.ps1` script and run it with PowerShell.

```powershell
# Run the script
./selfsign-path-v$Version.ps1 --help
```
## Signing Status
Script signing is $($IsScriptSigned ? 'enabled' : 'disabled'). Ensure you have
the necessary signing certificate configured in your environment variables.
"@
} else {
    # If no changes, create a minimal release note.
    $ReleaseNotes = @"
## Release Notes for $TagName
No changes since the last release.
## Installation
Download the attached `selfsign-path-v$Version.ps1` script and run it with PowerShell.
```powershell
# Run the script
./selfsign-path-v$Version.ps1 --help
```
## Signing Status
Script signing is $($IsScriptSigned ? 'enabled' : 'disabled'). Ensure you have
the necessary signing certificate configured in your environment variables.
"@
}
# Write the release notes to a file.
$ReleaseNotesFile = "RELEASE_NOTES.md"
Write-Host "Writing release notes to $ReleaseNotesFile..."
Set-Content -Path $ReleaseNotesFile -Value $ReleaseNotes
Write-Host "✅ Changelog generated successfully: $ReleaseNotesFile"
# If running in a GitHub Actions environment, set the output variable.
if ($env:GITHUB_ENV) {
    Add-Content -Path $env:GITHUB_ENV -Value "RELEASE_NOTES_FILE=$ReleaseNotesFile"
}


