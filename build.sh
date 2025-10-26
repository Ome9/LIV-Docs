#!/bin/bash
# Build script for LIV Format project

set -e

echo "Building LIV Format Binaries..."
echo "================================"

# Build main CLI
echo "Building liv CLI..."
go build -o bin/liv ./cmd/cli/

# Build PDF operations tool
echo "Building liv-pdf..."
go build -o bin/liv-pdf ./cmd/liv-pdf/

# Build builder
echo "Building liv-builder..."
go build -o bin/liv-builder ./cmd/builder/

# Build viewer
echo "Building liv-viewer..."
go build -o bin/liv-viewer ./cmd/viewer/

# Build integrity checker
echo "Building liv-integrity..."
go build -o bin/liv-integrity ./cmd/liv-integrity/

# Build manifest validator
echo "Building liv-manifest-validator..."
go build -o bin/liv-manifest-validator ./cmd/manifest-validator/

# Build security admin
echo "Building liv-security-admin..."
go build -o bin/liv-security-admin ./cmd/security-admin/

# Build permission server
echo "Building liv-permission-server..."
go build -o bin/liv-permission-server ./cmd/permission-server/

# Build liv-pack
echo "Building liv-pack..."
go build -o bin/liv-pack ./cmd/liv-pack/

# Build liv-converter
echo "Building liv-converter..."
go build -o bin/liv-converter ./cmd/liv-converter/

echo ""
echo "✓ All binaries built successfully!"
echo "  Output: ./bin/"
echo ""
echo "Available commands:"
echo "  - liv: Main CLI for LIV documents"
echo "  - liv-pdf: PDF operations (extract, merge, split, etc.)"
echo "  - liv-converter: PDF to LIV converter (RECOMMENDED FOR PDF→LIV)"
echo "  - liv-builder: Document builder"
echo "  - liv-viewer: Document viewer"
echo "  - liv-integrity: Integrity checker"
echo "  - liv-manifest-validator: Manifest validator"
echo "  - liv-security-admin: Security administration"
echo "  - liv-permission-server: Permission server"
echo "  - liv-pack: Document packer"
