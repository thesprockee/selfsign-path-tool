# Release Workflow Documentation

This document describes the Makefile-based release workflow for the selfsign-path-tool project.

## Overview

The release workflow has been refactored to use a Makefile that orchestrates the creation of release artifacts. This provides better maintainability and allows for easier local testing of the release process.

## Structure

### Makefile
- **Location**: `./Makefile`
- **Main Target**: `make dist` - Creates all release artifacts
- **Dependencies**: Requires `bash` and standard Unix tools (available on GitHub Actions Windows runners)

### Shell Scripts
All scripts are located in the `scripts/` directory:

- `create-versioned-script.sh` - Creates the versioned PowerShell script
- `sign-script.sh` - Signs the script if certificates are available
- `generate-changelog.sh` - Generates changelog and release notes

### GitHub Actions Workflow
- **Location**: `.github/workflows/release.yml`
- **Trigger**: Git tags matching `v*.*.*` pattern
- **Runner**: `windows-latest`
- **Key Change**: Now uses `make dist` instead of inline PowerShell commands

## Usage

### Local Testing
```bash
# Set required environment variables
export VERSION=1.0.0
export TAG_NAME=v1.0.0

# Create all release artifacts
make dist

# Clean up generated files
make clean

# Show help
make help
```

### Environment Variables
- `VERSION` - Version number (e.g., 1.0.0)
- `TAG_NAME` - Git tag name (e.g., v1.0.0)
- `SIGNING_CERT` - Base64 encoded signing certificate (optional)
- `SIGNING_CERT_PASSWORD` - Password for signing certificate (optional)

## Generated Artifacts
- `selfsign-path-v${VERSION}.ps1` - Versioned PowerShell script
- `RELEASE_NOTES.md` - Changelog and installation instructions

## Preserved Features
All existing functionality has been preserved:
- Version/tag extraction from GitHub events
- Conditional script signing (if certificates are available)
- Changelog generation from git history
- Release creation and artifact upload
- Error handling and logging

## Benefits
- **Modularity**: Individual tasks are now in separate, testable scripts
- **Maintainability**: Easier to modify and debug individual components
- **Local Testing**: Full release process can be tested locally
- **Platform Agnostic**: Makefile works on both Unix-like systems and Windows
- **Reduced Duplication**: Common patterns extracted into reusable scripts