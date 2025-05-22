#!/usr/bin/env bash
# Remove TatuScan client from ~/.local/bin (or $PREFIX/bin).
# Does not remove a dpkg-installed package — use: sudo apt remove tatuscan
set -euo pipefail

BINARY_NAME="${BINARY_NAME:-tatuscan}"

user_home() {
  if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
    getent passwd "$SUDO_USER" | cut -d: -f6
    return
  fi
  echo "${HOME}"
}

USER_LOCAL="$(user_home)/.local/bin/${BINARY_NAME}"

targets=()
if [ -n "${PREFIX:-}" ]; then
  targets+=("${PREFIX%/}/bin/${BINARY_NAME}")
else
  targets+=("$USER_LOCAL")
fi

removed=0
for target in "${targets[@]}"; do
  if [ ! -e "$target" ]; then
    continue
  fi
  dir="$(dirname "$target")"
  if [ -w "$dir" ] 2>/dev/null && rm -f "$target" 2>/dev/null; then
    :
  else
    sudo rm -f "$target"
  fi
  echo "Removed: $target"
  removed=$((removed + 1))
done

if [ "$removed" -eq 0 ]; then
  echo "Not installed (checked: ${targets[*]})"
fi
