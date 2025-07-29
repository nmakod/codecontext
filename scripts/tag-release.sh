#!/bin/bash
# Simple script to create a release tag
# Usage: ./scripts/tag-release.sh

set -euo pipefail

# Read version from VERSION file
VERSION=$(cat VERSION)

echo "🏷️  Creating release tag for v${VERSION}"

# Check if tag already exists
if git tag -l "v${VERSION}" | grep -q .; then
    echo "❌ Tag v${VERSION} already exists"
    exit 1
fi

# Check if CHANGELOG has entry for this version
if ! grep -q "## \[${VERSION}\]" CHANGELOG.md; then
    echo "⚠️  Warning: No changelog entry found for version ${VERSION}"
    echo "   Please update CHANGELOG.md before releasing"
    exit 1
fi

# Create and push tag
echo "Creating tag v${VERSION}..."
git tag -a "v${VERSION}" -m "Release v${VERSION}"

echo ""
echo "✅ Tag created successfully!"
echo ""
echo "To push the release:"
echo "  git push origin v${VERSION}"
echo ""
echo "This will trigger:"
echo "  - Security checks"
echo "  - Tests"
echo "  - Multi-platform builds"
echo "  - GitHub release with changelog"