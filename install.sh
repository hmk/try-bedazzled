#!/usr/bin/env sh
# install.sh — install try-bedazzled and wire up the shell function.
#
#   curl -fsSL https://raw.githubusercontent.com/hmk/try-bedazzled/main/install.sh | sh
#
# Env overrides:
#   TRY_INSTALL_DIR  install location for the binary  (default: /usr/local/bin)
#   TRY_VERSION      release tag to install            (default: latest)
#   TRY_NO_RC_EDIT   if set, skip editing the shell rc file

set -eu

REPO="hmk/try-bedazzled"
INSTALL_DIR="${TRY_INSTALL_DIR:-/usr/local/bin}"
VERSION="${TRY_VERSION:-latest}"

bold() { printf '\033[1m%s\033[0m\n' "$1"; }
info() { printf '  %s\n' "$1"; }
warn() { printf '\033[33m  %s\033[0m\n' "$1" >&2; }
die()  { printf '\033[31merror:\033[0m %s\n' "$1" >&2; exit 1; }

bold "try-bedazzled installer"

# --- detect platform ----------------------------------------------------------

uname_s=$(uname -s)
uname_m=$(uname -m)

case "$uname_s" in
  Darwin) os=darwin ;;
  Linux)  os=linux  ;;
  *) die "unsupported OS: $uname_s (try Homebrew or download from GitHub Releases)" ;;
esac

case "$uname_m" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) die "unsupported architecture: $uname_m" ;;
esac

info "platform: ${os}_${arch}"

# --- resolve download URL -----------------------------------------------------

if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/try-bedazzled_${os}_${arch}.tar.gz"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/try-bedazzled_${os}_${arch}.tar.gz"
fi

info "download:  $url"

# --- download + extract -------------------------------------------------------

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

if ! curl -fsSL -o "$tmp/try.tar.gz" "$url"; then
  die "download failed — check that ${VERSION} has a release asset for ${os}_${arch}"
fi

tar -xzf "$tmp/try.tar.gz" -C "$tmp"

if [ ! -x "$tmp/try" ]; then
  die "downloaded archive does not contain a 'try' binary"
fi

# --- install binary -----------------------------------------------------------

info "install:   $INSTALL_DIR/try"

if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "$tmp/try" "$INSTALL_DIR/try"
else
  if ! command -v sudo >/dev/null 2>&1; then
    die "$INSTALL_DIR is not writable and sudo is not available — set TRY_INSTALL_DIR to a writable path"
  fi
  warn "$INSTALL_DIR not writable — using sudo"
  sudo install -m 0755 "$tmp/try" "$INSTALL_DIR/try"
fi

# --- wire up shell function ---------------------------------------------------

if [ -n "${TRY_NO_RC_EDIT:-}" ]; then
  info "rc file:   skipped (TRY_NO_RC_EDIT set)"
  echo
  bold "Done."
  echo "  Add this line to your shell config to enable the 'try' function:"
  echo "    eval \"\$(try init)\"   # bash/zsh"
  echo "    try init | source     # fish"
  exit 0
fi

shell_name=$(basename "${SHELL:-/bin/sh}")

case "$shell_name" in
  bash) rc="$HOME/.bashrc";    line='eval "$(try init)"' ;;
  zsh)  rc="$HOME/.zshrc";     line='eval "$(try init)"' ;;
  fish) rc="$HOME/.config/fish/config.fish"; line='try init | source' ;;
  *)
    warn "unrecognized shell '$shell_name' — skipping rc edit"
    echo
    bold "Done."
    echo "  Add the appropriate line to your shell config:"
    echo "    eval \"\$(try init)\"   # bash/zsh"
    echo "    try init | source     # fish"
    exit 0
    ;;
esac

mkdir -p "$(dirname "$rc")"
[ -f "$rc" ] || : > "$rc"

if grep -Fq 'try init' "$rc" 2>/dev/null; then
  info "rc file:   $rc (already configured)"
else
  info "rc file:   $rc (appending)"
  {
    echo ''
    echo '# try-bedazzled (https://github.com/hmk/try-bedazzled)'
    echo "$line"
  } >> "$rc"
fi

# --- done ---------------------------------------------------------------------

echo
bold "Installed!"
echo "  Restart your shell or run:  source $rc"
echo "  Then try it:                try"
