#!/bin/bash
# generate-changelog.sh - Generate changelog and release notes

set -e

VERSION="${1:-${VERSION}}"
TAG_NAME="${2:-${TAG_NAME}}"

if [ -z "$VERSION" ]; then
    echo "❌ Error: VERSION must be provided as argument or environment variable"
    exit 1
fi

if [ -z "$TAG_NAME" ]; then
    TAG_NAME="v${VERSION}"
fi

echo "Generating changelog for version $VERSION (tag: $TAG_NAME)..."

# Get the previous tag
PREVIOUS_TAG=$(git tag --sort=-version:refname | head -1)
if [ -z "$PREVIOUS_TAG" ]; then
    # If no previous tag, use first commit
    PREVIOUS_TAG=$(git rev-list --max-parents=0 HEAD)
fi

echo "Previous tag/commit: $PREVIOUS_TAG"
echo "Current tag: $TAG_NAME"

# Generate changelog from git log
GIT_LOG=$(git log --pretty=format:"- %s (%h)" "${PREVIOUS_TAG}..HEAD" 2>/dev/null || true)

# Determine if script is signed (check environment variable)
SCRIPT_SIGNED="${SCRIPT_SIGNED:-false}"

if [ -n "$GIT_LOG" ]; then
    cat > RELEASE_NOTES.md << EOF
## Changes in $TAG_NAME

$GIT_LOG

## Installation

Download the attached \`selfsign-path-v$VERSION.ps1\` script and run it with PowerShell.

\`\`\`powershell
# Make the script executable and run it
.\selfsign-path-v$VERSION.ps1 --help
\`\`\`

## Script Verification
EOF

    if [ "$SCRIPT_SIGNED" = "true" ]; then
        cat >> RELEASE_NOTES.md << EOF

✅ **This script has been digitally signed** for security and authenticity.

You can verify the signature using:
\`\`\`powershell
Get-AuthenticodeSignature .\selfsign-path-v$VERSION.ps1
\`\`\`
EOF
    else
        cat >> RELEASE_NOTES.md << EOF

⚠️ **This script is not digitally signed.** Please verify the source and integrity before use.
EOF
    fi
else
    cat > RELEASE_NOTES.md << EOF
## Changes in $TAG_NAME

Initial release of the selfsign-path-tool utility.

## Installation

Download the attached \`selfsign-path-v$VERSION.ps1\` script and run it with PowerShell.

\`\`\`powershell
# Make the script executable and run it
.\selfsign-path-v$VERSION.ps1 --help
\`\`\`
EOF
fi

echo "📋 Changelog generated:"
cat RELEASE_NOTES.md