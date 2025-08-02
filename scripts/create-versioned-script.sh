#!/bin/bash
# create-versioned-script.sh - Create versioned PowerShell script

set -e

VERSION="${1:-${VERSION}}"

if [ -z "$VERSION" ]; then
    echo "❌ Error: VERSION must be provided as argument or environment variable"
    exit 1
fi

VERSIONED_SCRIPT="selfsign-path-v${VERSION}.ps1"

echo "Creating versioned script: $VERSIONED_SCRIPT"

# Copy the main script to the versioned name
if [ -f "selfsign-path.ps1" ]; then
    cp "selfsign-path.ps1" "$VERSIONED_SCRIPT"
    echo "✅ Created versioned script: $VERSIONED_SCRIPT"
    
    # Set environment variable for subsequent steps (GitHub Actions style)
    if [ -n "$GITHUB_ENV" ]; then
        echo "VERSIONED_SCRIPT=$VERSIONED_SCRIPT" >> "$GITHUB_ENV"
    fi
else
    echo "❌ Error: selfsign-path.ps1 not found"
    exit 1
fi