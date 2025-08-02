# selfsign-path-tool

A cross-platform code signing utility written in Go that manages and applies self-signed code signatures to executables and libraries.

## Features

- **Cross-platform**: Supports Windows and Linux
- **Self-contained**: Single binary executable with no external dependencies
- **Self-signed certificates**: Automatically generates and manages certificates
- **Pattern matching**: Supports glob patterns and recursive directory processing
- **Signature management**: Sign, verify, and remove signatures
- **Certificate store integration**: Installs certificates to system trust stores

## Quick Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/thesprockee/selfsign-path-tool/releases):

- **Linux (x86_64)**: `selfsign-path-tool-linux-amd64`
- **Windows (x86_64)**: `selfsign-path-tool-windows-amd64.exe`

### Make Executable (Linux/macOS)

```bash
chmod +x selfsign-path-tool-linux-amd64
```

### Build from Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/thesprockee/selfsign-path-tool.git
cd selfsign-path-tool
go build -o selfsign-path-tool
```

## Usage

### Basic Usage

```bash
# Sign a single executable
./selfsign-path-tool myapp.exe

# Sign all DLLs in a directory and its subdirectories  
./selfsign-path-tool -r "bin/*.dll"

# Check the signature status of files
./selfsign-path-tool --status *.exe

# Remove self-signed signatures
./selfsign-path-tool --clear -r release/
```

### Command Line Options

```
selfsign-path [OPTIONS] file_or_pattern...

OPTIONS:
    -r, --recurse               Recursively search directories
    -n, --name <CERT_NAME>      Certificate subject name (default: "LocalSign-SelfSigned")
    -c, --cert-file <FILE>      Use specific certificate file (.crt/.pem)
    -k, --key-file <FILE>       Use specific private key file (.key)
    --clear                     Remove self-signed signatures
    --status                    Check signature status
    -h, --help                  Show help
    --version                   Show version
```

### Examples

```bash
# Sign files with custom certificate name
./selfsign-path-tool -n "MyCompany-Dev" myapp.exe

# Use external certificate and key files
./selfsign-path-tool -c mycert.crt -k mykey.key myapp.exe

# Sign all executables in current directory
./selfsign-path-tool *.exe

# Recursively sign all binaries in a build directory
./selfsign-path-tool -r build/

# Check what's signed in a directory
./selfsign-path-tool --status -r release/

# Remove signatures from release builds
./selfsign-path-tool --clear -r release/
```

## How It Works

### Certificate Generation

On first run, the tool:

1. **Generates** a new RSA 2048-bit self-signed certificate
2. **Saves** the certificate and private key to a local directory:
   - **Windows**: `%APPDATA%\selfsign-path-tool\certificates\`
   - **Linux**: `~/.local/share/selfsign-path-tool/certificates/`
3. **Installs** the certificate to the system trust store:
   - **Windows**: Local Machine Trusted Root store (requires admin)
   - **Linux**: `/usr/local/share/ca-certificates/` or user directory

### File Signing

The signing process creates:

- **Windows**: `.sig` signature files alongside binaries (simplified Authenticode alternative)
- **Linux**: `.sig` detached signature files (similar to GPG signatures)

> **Note**: This implementation uses a simplified signing approach. For production code signing, consider using platform-specific tools like SignTool (Windows) or proper code signing certificates from Certificate Authorities.

### Supported File Types

The tool automatically detects and processes these file types:
- `.exe` - Executables
- `.dll` - Dynamic Link Libraries  
- `.msi` - Windows Installer packages
- `.sys` - System files
- `.com` - DOS executables
- `.ocx` - ActiveX controls
- `.scr` - Screen savers
- `.cpl` - Control Panel items

## Cross-Platform Differences

### Windows
- Uses PowerShell for certificate store operations
- Creates Authenticode-style signature files
- Supports Windows certificate store integration
- Requires administrator privileges for system certificate installation

### Linux
- Creates detached signature files
- Attempts to install certificates to system CA directories
- Falls back to user certificate directory if system installation fails
- Uses standard Linux certificate update tools when available

## Building and Development

### Prerequisites
- Go 1.21 or later
- Git

### Build Commands

```bash
# Build for current platform
go build -o selfsign-path-tool

# Build for specific platforms
GOOS=windows GOARCH=amd64 go build -o selfsign-path-tool-windows-amd64.exe
GOOS=linux GOARCH=amd64 go build -o selfsign-path-tool-linux-amd64

# Build optimized release binaries
go build -ldflags="-s -w" -o selfsign-path-tool
```

### Release Process

The project uses GitHub Actions for automated releases:

1. **Tag** a new version: `git tag v1.0.0 && git push origin v1.0.0`
2. **GitHub Actions** automatically builds cross-platform binaries
3. **Draft release** is created with binaries attached
4. **Review and publish** the release

## Migrating from PowerShell Version

This Go implementation replaces the previous PowerShell-based version with:

- ✅ **Better cross-platform support** - Single binary for Windows/Linux
- ✅ **No external dependencies** - No PowerShell or CMake required
- ✅ **Faster execution** - Compiled binary vs interpreted scripts
- ✅ **Simplified deployment** - Single file instead of multiple scripts
- ✅ **Same CLI interface** - Compatible command-line arguments

### Migration Steps

1. **Download** the appropriate binary for your platform
2. **Replace** existing PowerShell scripts with the binary
3. **Update** any automation scripts to use the new binary
4. **Test** with your existing workflows

The command-line interface is designed to be compatible with the PowerShell version.

## Troubleshooting

### Certificate Issues

**Problem**: Certificate not trusted
**Solution**: Run with administrator/root privileges to install to system store

**Problem**: Certificate generation fails
**Solution**: Check file system permissions for certificate directory

### Signing Issues  

**Problem**: File not found or pattern doesn't match
**Solution**: Use absolute paths or check current working directory

**Problem**: Permission denied
**Solution**: Ensure write access to target files and directories

### Cross-Platform Issues

**Problem**: Binary won't execute
**Solution**: Check execute permissions (`chmod +x`) and architecture compatibility

## License

This project is open source. See the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

- **Issues**: Report bugs and request features via GitHub Issues
- **Documentation**: This README and built-in help (`--help`)
- **Community**: GitHub Discussions for questions and support
