# Contributing to InTunnel Client

Thank you for your interest in contributing to InTunnel Client! 

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/InTech-Solutions/intunnel-client/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Your OS and version

### Suggesting Features

1. Open an issue with the "feature request" label
2. Describe the feature and its use case
3. Be open to discussion

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Test thoroughly
5. Commit with clear messages
6. Push and create a Pull Request

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format code
- Add comments for exported functions
- Keep functions small and focused

## Development Setup

```bash
# Clone the repo
git clone https://github.com/InTech-Solutions/intunnel-client.git
cd intunnel-client

# Build CLI
cd cmd/cli
go build -o InTunnel-CLI main.go

# Build GUI
cd ../gui
go build -o InTunnel-GUI main.go
```

## Code of Conduct

Be respectful and inclusive. We follow the [Contributor Covenant](https://www.contributor-covenant.org/).

## Questions?

Open an issue or contact us at support@intech-eg.tech
