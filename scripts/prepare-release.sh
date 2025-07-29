#!/bin/bash

# CodeContext Release Preparation Script
# Usage: ./scripts/prepare-release.sh <version>

set -e

VERSION=${1:-"2.4.0"}
BINARY_NAME="codecontext"
BUILD_DIR="dist"

echo "🚀 Preparing CodeContext release v${VERSION}"

# Verify we're in the right directory
if [[ ! -f "go.mod" ]]; then
    echo "❌ Error: Must be run from project root directory"
    exit 1
fi

# Verify version format
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ Error: Version must be in format X.Y.Z (e.g., 2.0.0)"
    exit 1
fi

echo "📋 Release checklist:"
echo "  ✅ Version: ${VERSION}"
echo "  ✅ Binary: ${BINARY_NAME}"
echo "  ✅ Build dir: ${BUILD_DIR}"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
make clean

# Run tests
echo "🧪 Running tests..."
go test ./... || {
    echo "❌ Tests failed. Please fix before releasing."
    exit 1
}

# Format and lint
echo "🎨 Formatting and linting code..."
go fmt ./...

# Build for all platforms
echo "🔨 Building for all platforms..."
make build-all VERSION=${VERSION}

# Create release tarballs
echo "📦 Creating release artifacts..."
make release VERSION=${VERSION}

# Generate checksums
echo "🔐 Generating checksums..."
make checksums

# Create source tarball for Homebrew
echo "🍺 Creating Homebrew source tarball..."
tar --exclude='.git' --exclude='dist' --exclude='node_modules' --exclude='*.tar.gz' \
    -czf ${BINARY_NAME}-${VERSION}.tar.gz .

# Generate SHA256 for Homebrew formula
echo "📝 Generating SHA256 for Homebrew..."
HOMEBREW_SHA256=$(shasum -a 256 ${BINARY_NAME}-${VERSION}.tar.gz | cut -d' ' -f1)
echo "Homebrew SHA256: ${HOMEBREW_SHA256}"

# Update Homebrew formula
echo "📋 Updating Homebrew formula..."
sed -i.bak "s/sha256 \".*\"/sha256 \"${HOMEBREW_SHA256}\"/" Formula/codecontext.rb
sed -i.bak "s/v[0-9]\+\.[0-9]\+\.[0-9]\+/v${VERSION}/g" Formula/codecontext.rb
rm Formula/codecontext.rb.bak

echo "✅ Release preparation complete!"
echo ""
echo "📋 Next steps:"
echo "  1. Update CHANGELOG.md with release notes for v${VERSION}"
echo "  2. Review generated files in ${BUILD_DIR}/"
echo "  3. Test the binaries locally"
echo "  4. Commit changes: git add . && git commit -m 'chore: prepare release v${VERSION}'"
echo "  5. Create git tag: git tag -a v${VERSION} -m 'Release v${VERSION}'"
echo "  6. Push to GitHub: git push origin main"
echo "  7. Push tag to trigger release: git push origin v${VERSION}"
echo ""
echo "🎉 GitHub Actions will automatically create the release when the tag is pushed!"
echo ""
echo "Note: The release will include:"
echo "  - Binaries for all platforms (Linux, macOS, Windows)"
echo "  - Checksums file"
echo "  - Release notes from CHANGELOG.md"
echo "  - Proper version tags in binaries"