# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| < 2.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do NOT** open a public issue
2. Email us at: security@intech-eg.tech
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested fixes

We will:
- Acknowledge receipt within 48 hours
- Provide an initial assessment within 7 days
- Work with you to understand and resolve the issue
- Credit you in the release notes (if desired)

## Security Measures

- All tunnel traffic is encrypted via HTTPS/TLS
- Token-based authentication
- No sensitive data stored locally
- Regular dependency updates

## Code Signing

Our releases are signed using certificates provided by [SignPath Foundation](https://signpath.org).

Verify downloads using the SHA256 checksums provided with each release.
