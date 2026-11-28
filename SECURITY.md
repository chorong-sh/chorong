# Security Policy

## Supported Versions

Security updates are applied to the latest release and the `main` branch.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| main    | :white_check_mark: |
| older   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in this repository, please report it responsibly.

### How to Report

1. **Do not** open a public GitHub issue for security vulnerabilities.
2. Send a private report via [GitHub Security Advisories](https://github.com/chorong-sh/chorong/security/advisories/new) or contact the maintainer directly.
3. Include as much detail as possible:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: Within 7 days.
- **Assessment**: Evaluated and prioritized based on severity.
- **Resolution**: Fixed as soon as reasonably possible.
- **Disclosure**: Credit given in the release notes unless you prefer anonymity.

## Security Practices

This repository enforces:
- Hermetic static analysis and security scanning in Bazel (`nogo` with `vet`, `staticcheck`, `gosec`, `ineffassign`, `iface`, `godernize`; `shellcheck`; full-history `gitleaks`)
- Pre-commit secret scanning via `.githooks/pre-commit/gitleaks.sh`
- 40-character commit-SHA pinning for all GitHub Actions workflows
- Automated CI checks on pull requests and `main` (`.github/workflows/bazel-test.yml`)
