# Install

The fast path is the install script in the [README](README.md). This file covers the alternatives.

## Install script options

The install script accepts a few env-var overrides:

```bash
# Install to a custom directory (no sudo needed if writable)
TRY_INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/hmk/try-bedazzled/main/install.sh | sh

# Pin to a specific version
TRY_VERSION=v0.1.3 curl -fsSL https://raw.githubusercontent.com/hmk/try-bedazzled/main/install.sh | sh

# Install the binary but skip editing your shell rc file
TRY_NO_RC_EDIT=1 curl -fsSL https://raw.githubusercontent.com/hmk/try-bedazzled/main/install.sh | sh
```

## With Go

```bash
go install github.com/hmk/try-bedazzled/cmd/try@latest
```

Then add the shell function manually (see "Manual shell setup" below).

## Manual download

Pick the tarball for your OS/arch from [Releases](https://github.com/hmk/try-bedazzled/releases/latest):

```bash
# macOS, Apple Silicon
curl -fsSL -o try.tar.gz https://github.com/hmk/try-bedazzled/releases/latest/download/try-bedazzled_darwin_arm64.tar.gz

# macOS, Intel
curl -fsSL -o try.tar.gz https://github.com/hmk/try-bedazzled/releases/latest/download/try-bedazzled_darwin_amd64.tar.gz

# Linux, amd64
curl -fsSL -o try.tar.gz https://github.com/hmk/try-bedazzled/releases/latest/download/try-bedazzled_linux_amd64.tar.gz

# Linux, arm64
curl -fsSL -o try.tar.gz https://github.com/hmk/try-bedazzled/releases/latest/download/try-bedazzled_linux_arm64.tar.gz

tar xzf try.tar.gz
sudo mv try /usr/local/bin/
```

Then add the shell function manually (see "Manual shell setup" below).

## Debian / RPM packages

Each release ships `.deb` and `.rpm` packages on the [Releases page](https://github.com/hmk/try-bedazzled/releases/latest).

```bash
# Debian/Ubuntu
sudo dpkg -i try-bedazzled_linux_amd64.deb

# RHEL/Fedora
sudo rpm -i try-bedazzled_linux_amd64.rpm
```

Then add the shell function manually (see "Manual shell setup" below).

## Manual shell setup

`try` needs a shell function wrapper to be able to `cd` your shell into the selected directory. The install script adds it for you; the other paths above don't.

**bash / zsh** (`.bashrc` or `.zshrc`):
```bash
eval "$(try init)"
```

**fish** (`config.fish`):
```fish
try init | source
```

Use a custom tries directory:
```bash
eval "$(try init ~/workspace/experiments)"
```
