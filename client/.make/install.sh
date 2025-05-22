#!/usr/bin/env bash
# Install TatuScan client binary to ~/.local/bin (or $PREFIX/bin).
# System-wide + daemon: make package && sudo dpkg -i dist/*.deb
set -euo pipefail

BINARY_NAME="${BINARY_NAME:-tatuscan}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Prefer platform build; fall back to flat bin/ layout.
if [ -x "$ROOT_DIR/bin/linux/$BINARY_NAME" ]; then
  SRC="$ROOT_DIR/bin/linux/$BINARY_NAME"
elif [ -x "$ROOT_DIR/bin/$BINARY_NAME" ]; then
  SRC="$ROOT_DIR/bin/$BINARY_NAME"
else
  SRC="$ROOT_DIR/bin/linux/$BINARY_NAME"
fi

user_home() {
  if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
    getent passwd "$SUDO_USER" | cut -d: -f6
    return
  fi
  echo "${HOME}"
}

if [ -n "${PREFIX:-}" ]; then
  INSTALL_DIR="${PREFIX%/}/bin"
else
  INSTALL_DIR="$(user_home)/.local/bin"
fi

if [ ! -x "$SRC" ]; then
  echo "error: missing binary: $SRC" >&2
  echo "hint:  run make build first" >&2
  exit 1
fi

DEST="$INSTALL_DIR/$BINARY_NAME"
UH="$(user_home)"

if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
  case "$INSTALL_DIR" in
    "$UH" | "$UH"/*)
      sudo -u "$SUDO_USER" install -D -m 755 "$SRC" "$DEST"
      ;;
    *)
      install -D -m 755 "$SRC" "$DEST"
      ;;
  esac
elif [ -w "$(dirname "$INSTALL_DIR")" ] 2>/dev/null || [ -w "$INSTALL_DIR" ] 2>/dev/null; then
  install -D -m 755 "$SRC" "$DEST"
else
  echo "Destination not writable; using sudo for install only..."
  sudo install -D -m 755 "$SRC" "$DEST"
fi

echo "Installed: $DEST"
echo "Run daemon with: TATUSCAN_URL=http://host:8040 $DEST -d"
echo "System-wide + systemd: make package && sudo dpkg -i dist/*.deb"
