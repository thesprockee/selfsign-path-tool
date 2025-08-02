# Release Workflow Documentation

This document describes the CMake-based release workflow for the selfsign-path-tool project.

## Overview

The release workflow uses a CMake build system that orchestrates the creation of release artifacts using PowerShell scripts. This provides Windows compatibility while maintaining cross-platform support and allows for easier local testing of the release process.

## Structure

### CMake Build System
- **Location**: `./CMakeLists.txt`
- **Main Target**: `cmake --build . --target dist` - Creates all release artifacts
- **Requirements**: CMake 3.15+ and PowerShell (pwsh or powershell)

### PowerShell Scripts
All scripts are located in the `scripts/` directory:

- `create-versioned-script.ps1` - Creates the versioned PowerShell script
- `sign-script.ps1` - Signs the script if certificates are available
- `generate-changelog.ps1` - Generates changelog and release notes

### GitHub Actions Workflow
- **Location**: `.github/workflows/release.yml`
- **Trigger**: Git tags matching `v*.*.*` pattern
- **Runner**: `windows-latest`
- **Key Change**: Now uses CMake with PowerShell instead of make with bash scripts

## Usage

### Local Testing
```powershell
# Set required environment variables (Windows)
$env:VERSION = "1.0.0"
$env:TAG_NAME = "v1.0.0"

# Or on Unix-like systems
export VERSION=1.0.0
export TAG_NAME=v1.0.0

# Create build directory and configure
mkdir build
cd build
cmake .. -DVERSION=$env:VERSION

# Create all release artifacts
cmake --build . --target dist

# Clean up generated files
cmake --build . --target clean-dist

# Show help
cmake --build . --target help-dist
```

### Environment Variables
- `VERSION` - Version number (e.g., 1.0.0)
- `TAG_NAME` - Git tag name (e.g., v1.0.0)
- `SIGNING_CERT` - Base64 encoded signing certificate (optional)
- `SIGNING_CERT_PASSWORD` - Password for signing certificate (optional)

## Generated Artifacts
- `selfsign-path-v${VERSION}.ps1` - Versioned PowerShell script
- `RELEASE_NOTES.md` - Changelog and installation instructions

## Available CMake Targets
- `dist` - Create all release artifacts (main target)
- `create-versioned-script` - Create versioned PowerShell script
- `generate-changelog` - Generate changelog and release notes
- `sign-script` - Sign the script (if certificates available)
- `clean-dist` - Clean generated distribution files
- `help-dist` - Show help information

## Preserved Features
All existing functionality has been preserved:
- Version/tag extraction from GitHub events
- Conditional script signing (if certificates are available)
- Changelog generation from git history
- Release creation and artifact upload
- Error handling and logging

## Benefits
- **Windows Compatible**: Uses CMake and PowerShell instead of make and bash
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **Modularity**: Individual tasks are now in separate, testable scripts
- **Maintainability**: Easier to modify and debug individual components
- **Local Testing**: Full release process can be tested locally
- **IDE Support**: CMake provides better IDE integration than Makefiles