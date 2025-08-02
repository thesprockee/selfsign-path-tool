# Makefile for selfsign-path-tool release automation
# Supports both Unix-like systems and Windows (with make/bash available)

# Default shell for make
SHELL := bash

# Variables that can be overridden
VERSION ?= $(shell echo "${VERSION:-unknown}")
TAG_NAME ?= $(shell echo "${TAG_NAME:-v$(VERSION)}")
VERSIONED_SCRIPT ?= selfsign-path-v$(VERSION).ps1

# Default target
.PHONY: all
all: dist

# Main distribution target
.PHONY: dist
dist: create-versioned-script sign-script generate-changelog
	@echo "✅ Distribution artifacts created successfully!"
	@echo "📦 Version: $(VERSION)"
	@echo "📄 Script: $(VERSIONED_SCRIPT)"
	@echo "📝 Release notes: RELEASE_NOTES.md"

# Create versioned script
.PHONY: create-versioned-script
create-versioned-script:
	@echo "📝 Creating versioned script..."
	@./scripts/create-versioned-script.sh "$(VERSION)"

# Sign script if certificates are available
.PHONY: sign-script  
sign-script: create-versioned-script
	@echo "🔐 Signing script..."
	@./scripts/sign-script.sh "$(VERSIONED_SCRIPT)"

# Generate changelog and release notes
.PHONY: generate-changelog
generate-changelog:
	@echo "📋 Generating changelog..."
	@./scripts/generate-changelog.sh "$(VERSION)" "$(TAG_NAME)"

# Clean up generated files
.PHONY: clean
clean:
	@echo "🧹 Cleaning up generated files..."
	@rm -f selfsign-path-v*.ps1
	@rm -f RELEASE_NOTES.md
	@echo "✅ Cleanup completed"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  dist               - Create all distribution artifacts"
	@echo "  create-versioned-script - Create versioned PowerShell script"
	@echo "  sign-script        - Sign the script if certificates are available"
	@echo "  generate-changelog - Generate changelog and release notes"
	@echo "  clean              - Remove generated files"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION            - Version number (e.g., 1.0.0)"
	@echo "  TAG_NAME           - Git tag name (e.g., v1.0.0)"
	@echo "  SIGNING_CERT       - Base64 encoded signing certificate"
	@echo "  SIGNING_CERT_PASSWORD - Password for signing certificate"