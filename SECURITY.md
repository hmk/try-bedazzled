# Security Policy

## Reporting a vulnerability

Please email **jake@costa.security** with details of any suspected security
issue. Expect an acknowledgement within 72 hours.

Do not open a public GitHub issue for security-sensitive reports — it exposes
the vulnerability before a fix is available.

When reporting, include:

- Affected version(s) (`try --version`)
- Operating system and terminal
- Steps to reproduce
- Any relevant logs, crash output, or proof-of-concept

## Supported versions

Only the latest minor release receives security fixes. Older versions are
patched on a best-effort basis.

## Scope

try-bedazzled is a local TUI for managing scratch directories. In-scope
security concerns include:

- Path traversal or arbitrary file access beyond the configured tries path
- Command injection via filenames, shell integration output, or config
- Credential leakage via config loading or logging

Out of scope: terminal-emulator bugs, dependencies' own vulnerabilities
(report those upstream; we track them via Dependabot).
