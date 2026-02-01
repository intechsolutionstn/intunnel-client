# InTunnel Client

<p align="center">
  <img src="https://intunnel.cloud/static/logos/intunnel-logo-horizontal.svg" alt="InTunnel" width="300">
</p>

<p align="center">
  <strong>Expose your local services to the internet securely</strong>
</p>

<p align="center">
  <a href="https://github.com/InTech-Solutions/intunnel-client/releases"><img src="https://img.shields.io/github/v/release/InTech-Solutions/intunnel-client" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/InTech-Solutions/intunnel-client" alt="License"></a>
  <a href="https://intunnel.cloud"><img src="https://img.shields.io/badge/website-intunnel.cloud-blue" alt="Website"></a>
</p>

---

## What is InTunnel?

InTunnel is a self-hosted tunneling solution that allows you to expose your local services to the internet via secure HTTPS tunnels. Think of it as a self-hosted alternative to ngrok.

## Features

- 🔒 **Secure HTTPS tunnels** - All traffic is encrypted end-to-end
- 🌐 **Custom subdomains** - Get your own `yourapp.intunnel.cloud` URL
- 🖥️ **Cross-platform** - Works on Windows, macOS, and Linux
- 🎨 **GUI & CLI clients** - Choose your preferred interface
- 🔑 **Token-based auth** - Secure token authentication
- 🌍 **Multi-language** - English, French, and Arabic support

## Installation

### GUI Client (Recommended)

Download the GUI client for your platform from the [Releases](https://github.com/InTech-Solutions/intunnel-client/releases) page or directly from [intunnel.cloud](https://intunnel.cloud/#download).

| Platform | Download |
|----------|----------|
| Windows | [InTunnel-GUI-Windows.exe](https://intunnel.cloud/download/gui/windows) |
| macOS (Apple Silicon) | [InTunnel-GUI-macOS-ARM](https://intunnel.cloud/download/gui/macos-arm) |
| macOS (Intel) | [InTunnel-GUI-macOS-Intel](https://intunnel.cloud/download/gui/macos-intel) |
| Ubuntu/Debian | [InTunnel-GUI-Ubuntu](https://intunnel.cloud/download/gui/ubuntu) |
| Linux | [InTunnel-GUI-Linux](https://intunnel.cloud/download/gui/linux) |

### CLI Client

For advanced users who prefer the command line:

| Platform | Download |
|----------|----------|
| Windows | [InTunnel-CLI-Windows.exe](https://intunnel.cloud/download/client/windows) |
| macOS (Apple Silicon) | [InTunnel-CLI-macOS-ARM](https://intunnel.cloud/download/client/macos-arm) |
| macOS (Intel) | [InTunnel-CLI-macOS-Intel](https://intunnel.cloud/download/client/macos-intel) |
| Ubuntu/Debian | [InTunnel-CLI-Ubuntu](https://intunnel.cloud/download/client/ubuntu) |
| Linux | [InTunnel-CLI-Linux](https://intunnel.cloud/download/client/linux) |

## Quick Start

1. **Create an account** at [intunnel.cloud](https://intunnel.cloud/register)
2. **Create a tunnel** from your dashboard
3. **Copy your token** from the tunnel details
4. **Run the client** and paste your token

```bash
# CLI example
./InTunnel-Linux
# Enter your token when prompted
```

Your local service is now accessible at `https://yoursubdomain.intunnel.cloud`!

## Building from Source

### Prerequisites

- Go 1.21 or later
- FRP client binary (frpc)

### Build CLI Client

```bash
cd cmd/cli
go build -o InTunnel-CLI main.go
```

### Build GUI Client

```bash
cd cmd/gui
go build -o InTunnel-GUI main.go
```

## Verifying Downloads

All releases include SHA256 checksums. Verify your download:

**Windows (PowerShell):**
```powershell
Get-FileHash .\InTunnel-GUI-Windows.exe -Algorithm SHA256
```

**macOS/Linux:**
```bash
shasum -a 256 InTunnel-GUI-*
```

Compare the output with the checksums on the [download page](https://intunnel.cloud/#download).

## Code Signing Policy

Free code signing provided by [SignPath.io](https://about.signpath.io), certificate by [SignPath Foundation](https://signpath.org).

### Team Roles
- **Maintainers**: [InTech Solutions Team](https://github.com/orgs/InTech-Solutions/people)
- **Approvers**: Repository owners

### Privacy Policy
This program will not transfer any information to other networked systems unless specifically requested by the user or the person installing or operating it.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📧 Email: support@intech-eg.tech
- 🌐 Website: [intunnel.cloud](https://intunnel.cloud)
- 🐛 Issues: [GitHub Issues](https://github.com/InTech-Solutions/intunnel-client/issues)

---

<p align="center">
  Made with ❤️ by <a href="https://intech-eg.tech">InTech Solutions</a>
</p>
