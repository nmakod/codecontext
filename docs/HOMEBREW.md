# Homebrew Publishing Guide for CodeContext

## Overview

This guide covers how to publish CodeContext to Homebrew for easy installation on macOS.

## Current Status

✅ **Homebrew formula created**: `Formula/codecontext.rb`  
✅ **Build system ready**: Makefile with cross-platform support  
✅ **Version management**: Build-time version injection  
✅ **Source tarball**: Ready for GitHub releases  
✅ **SHA256 checksum**: Generated and updated in formula  

## Formula Details

**Location**: `Formula/codecontext.rb`  
**Package name**: `codecontext`  
**Dependencies**: Go (build-time only)  
**License**: MIT  
**Platforms**: macOS (Intel + Apple Silicon)  

## Publishing Options

### Option 1: Official Homebrew Core (Recommended for popular tools)

1. **Meet requirements**:
   - Stable, notable project with significant usage
   - No duplicate functionality with existing formulas
   - Actively maintained with regular releases

2. **Submit formula**:
   ```bash
   # Fork homebrew-core
   git clone https://github.com/Homebrew/homebrew-core.git
   cd homebrew-core
   
   # Copy our formula
   cp path/to/codecontext/Formula/codecontext.rb Formula/
   
   # Test locally
   brew install --build-from-source Formula/codecontext.rb
   brew test codecontext
   brew audit --new-formula codecontext
   
   # Submit PR
   git add Formula/codecontext.rb
   git commit -m "Add codecontext formula"
   git push origin main
   # Create PR to Homebrew/homebrew-core
   ```

### Option 2: Custom Tap (Immediate availability)

1. **Create a tap repository**:
   ```bash
   # Create repository: homebrew-codecontext
   git clone https://github.com/nmakod/homebrew-codecontext.git
   cd homebrew-codecontext
   
   # Copy formula
   cp path/to/codecontext/Formula/codecontext.rb .
   
   # Commit and push
   git add codecontext.rb
   git commit -m "Add CodeContext formula"
   git push origin main
   ```

2. **Users install via**:
   ```bash
   brew tap nmakod/codecontext
   brew install codecontext
   ```

### Option 3: Direct Formula Installation

Users can install directly from our repository:
```bash
brew install --build-from-source https://raw.githubusercontent.com/nmakod/codecontext/main/Formula/codecontext.rb
```

## Release Process

### 1. Prepare Release

```bash
# Use our automated script
./scripts/prepare-release.sh 2.0.0

# Or manually:
make release VERSION=2.0.0
tar --exclude='.git' -czf codecontext-2.0.0.tar.gz .
shasum -a 256 codecontext-2.0.0.tar.gz
```

### 2. Create GitHub Release

1. **Push changes**:
   ```bash
   git add .
   git commit -m "Release v2.0.0"
   git tag v2.0.0
   git push origin main --tags
   ```

2. **Create GitHub release**:
   - Go to GitHub releases page
   - Click "Create a new release"
   - Tag: `v2.0.0`
   - Title: `CodeContext v2.0.0`
   - Upload artifacts from `dist/` directory
   - Publish release

### 3. Update Formula

The formula automatically pulls from GitHub releases:
```ruby
url "https://github.com/nmakod/codecontext/archive/v2.0.0.tar.gz"
sha256 "72f79124718fe1d5f9787673ac62c4871168a8927f948ca99156d02c16da89c9"
```

### 4. Test Installation

```bash
# Test local formula
brew install --build-from-source Formula/codecontext.rb

# Test functionality
codecontext --version
codecontext --help

# Run formula tests
brew test codecontext

# Audit formula
brew audit codecontext
```

## Formula Features

### Build Configuration

- **Go build**: Uses standard `go build` with ldflags for version info
- **Universal binary**: Supports both Intel and Apple Silicon
- **Version injection**: Build-time version, date, and commit info

### Tests Included

- ✅ Binary execution test
- ✅ Version command verification  
- ✅ Help command functionality
- ✅ Basic file analysis test
- ✅ Configuration file handling

### Dependencies

- **Build-time**: Go 1.19+
- **Runtime**: None (statically linked)

## Maintenance

### Updating Formula

For new releases:
1. Update version in formula
2. Generate new SHA256 checksum
3. Test locally
4. Submit to tap/homebrew-core

### Version Management

Our formula supports:
- **Stable releases**: From GitHub tags
- **HEAD installation**: Latest main branch
- **Version pinning**: Specific version installation

## Best Practices

### Security
- ✅ SHA256 checksum verification
- ✅ Official GitHub releases only
- ✅ No external dependencies at runtime

### User Experience
- ✅ Clear description and homepage
- ✅ Comprehensive tests
- ✅ Proper error handling
- ✅ Version information available

### Development
- ✅ Automated release preparation
- ✅ Cross-platform build support
- ✅ CI/CD friendly process

## Troubleshooting

### Common Issues

**Build failures**: Check Go version compatibility
**SHA256 mismatch**: Regenerate checksum for exact tarball
**Formula syntax**: Use `brew audit` to validate

### Testing Locally

```bash
# Install from local formula
brew install --build-from-source Formula/codecontext.rb

# Uninstall
brew uninstall codecontext

# Test specific functionality
brew test codecontext
```

## Next Steps

1. **✅ Create GitHub release v2.0.0**
2. **🔄 Choose publishing approach** (tap vs homebrew-core)
3. **🔄 Test installation process**
4. **🔄 Submit to chosen platform**
5. **🔄 Update documentation with installation instructions**

---

*This guide ensures CodeContext can be easily installed by developers worldwide via Homebrew's trusted package manager.*