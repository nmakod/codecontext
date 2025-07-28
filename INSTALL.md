# Installation Guide for CodeContext

## Prerequisites

- **Go 1.19+**: Required for building from source
- **macOS/Linux/Windows**: Cross-platform support
- **Git**: For cloning and development

## Installation Methods

### 1. Homebrew (macOS - Recommended)

```bash
# Install from custom tap (once published)
brew tap nmakod/codecontext
brew install codecontext

# Or install directly from formula
brew install --build-from-source Formula/codecontext.rb
```

### 2. Pre-built Binaries

Download the latest release from GitHub:

```bash
# macOS (Intel)
curl -L https://github.com/nmakod/codecontext/releases/download/v2.0.0/codecontext-2.0.0-darwin-amd64.tar.gz | tar xz

# macOS (Apple Silicon)
curl -L https://github.com/nmakod/codecontext/releases/download/v2.0.0/codecontext-2.0.0-darwin-arm64.tar.gz | tar xz

# Linux (Intel)
curl -L https://github.com/nmakod/codecontext/releases/download/v2.0.0/codecontext-2.0.0-linux-amd64.tar.gz | tar xz

# Linux (ARM)
curl -L https://github.com/nmakod/codecontext/releases/download/v2.0.0/codecontext-2.0.0-linux-arm64.tar.gz | tar xz
```

Move the binary to your PATH:
```bash
sudo mv codecontext /usr/local/bin/
```

### 3. Build from Source

```bash
# Clone the repository
git clone https://github.com/nmakod/codecontext.git
cd codecontext

# Build for your platform
make build

# Install locally
make install
```

### 4. Go Install (Development)

```bash
go install github.com/nmakod/codecontext/cmd/codecontext@latest
```

## Verification

After installation, verify CodeContext is working:

```bash
# Check version
codecontext --version

# View help
codecontext --help

# Initialize a project
cd your-project
codecontext init
```

## Next Steps

1. **Initialize your project**: `codecontext init`
2. **Generate context map**: `codecontext generate`
3. **Enable watch mode**: `codecontext watch`
4. **Optimize with compaction**: `codecontext compact`

## Troubleshooting

### Common Issues

**Command not found**: Ensure the binary is in your PATH
```bash
export PATH=$PATH:/usr/local/bin
```

**Permission denied**: Make the binary executable
```bash
chmod +x codecontext
```

**Build errors**: Ensure Go 1.19+ is installed
```bash
go version
```

### Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/nmakod/codecontext/issues)
- **Documentation**: Check the README and inline help
- **Development**: See CONTRIBUTING.md for development setup

## Uninstallation

### Homebrew
```bash
brew uninstall codecontext
```

### Manual Installation
```bash
rm /usr/local/bin/codecontext
rm -rf ~/.codecontext  # Optional: remove config directory
```