#!/usr/bin/env bash
# Build a .deb for the TatuScan agent: /usr/bin + systemd unit + enable --now.
set -euo pipefail

BINARY_NAME="${BINARY_NAME:-tatuscan}"
PACKAGE_NAME="${PACKAGE_NAME:-$BINARY_NAME}"
SERVICE_NAME="${SERVICE_NAME:-tatuscan}"
SERVICE_USER="${SERVICE_USER:-tatuscan}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
STAGE="$DIST_DIR/deb-root"

if [ -x "$ROOT_DIR/bin/linux/$BINARY_NAME" ]; then
  SRC="$ROOT_DIR/bin/linux/$BINARY_NAME"
elif [ -x "$ROOT_DIR/bin/$BINARY_NAME" ]; then
  SRC="$ROOT_DIR/bin/$BINARY_NAME"
else
  SRC="$ROOT_DIR/bin/linux/$BINARY_NAME"
fi

VERSION="${VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo 0.0.0)}"
VERSION="${VERSION#v}"
ARCH="$(dpkg --print-architecture 2>/dev/null || echo amd64)"
MAINTAINER="${MAINTAINER:-$(git -C "$ROOT_DIR" config user.email 2>/dev/null || echo unknown@localhost)}"
MAINTAINER_NAME="${MAINTAINER_NAME:-$(git -C "$ROOT_DIR" config user.name 2>/dev/null || echo Unknown)}"
DESCRIPTION="${DESCRIPTION:-TatuScan monitoring agent}"

if [ ! -x "$SRC" ]; then
  echo "error: missing binary: $SRC" >&2
  echo "hint:  run make build first" >&2
  exit 1
fi

if ! command -v dpkg-deb >/dev/null 2>&1; then
  echo "error: dpkg-deb not found (install dpkg)" >&2
  exit 1
fi

rm -rf "$STAGE"
install -D -m 755 "$SRC" "$STAGE/usr/bin/$BINARY_NAME"

install -d -m 755 "$STAGE/etc/tatuscan"
cat >"$STAGE/etc/tatuscan/agent.env" <<EOF
# Conffile — edit after install if needed
TATUSCAN_URL=${TATUSCAN_URL:-http://127.0.0.1:8040}
TATUSCAN_INTERVAL=${TATUSCAN_INTERVAL:-60s}
EOF

install -d -m 755 "$STAGE/lib/systemd/system"
cat >"$STAGE/lib/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=TatuScan monitoring agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
EnvironmentFile=-/etc/tatuscan/agent.env
ExecStart=/usr/bin/$BINARY_NAME -d
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

mkdir -p "$STAGE/DEBIAN"
cat >"$STAGE/DEBIAN/control" <<EOF
Package: ${PACKAGE_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${MAINTAINER_NAME} <${MAINTAINER}>
Depends: systemd
Description: ${DESCRIPTION}
 TatuScan host inventory agent. Installs /usr/bin/${BINARY_NAME}
 and a systemd unit that runs the agent as a daemon.
EOF

cat >"$STAGE/DEBIAN/conffiles" <<EOF
/etc/tatuscan/agent.env
EOF

cat >"$STAGE/DEBIAN/postinst" <<EOF
#!/bin/sh
set -e
SERVICE_USER="$SERVICE_USER"
SERVICE_NAME="$SERVICE_NAME"

if ! getent passwd "\$SERVICE_USER" >/dev/null 2>&1; then
  useradd --system --shell /bin/false --home-dir /var/lib/tatuscan \\
    --create-home "\$SERVICE_USER" || true
fi

chown root:"\$SERVICE_USER" /etc/tatuscan/agent.env 2>/dev/null || true
chmod 0640 /etc/tatuscan/agent.env 2>/dev/null || chmod 0644 /etc/tatuscan/agent.env

if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload
  systemctl enable --now "\$SERVICE_NAME.service" || true
fi

exit 0
EOF
chmod 0755 "$STAGE/DEBIAN/postinst"

cat >"$STAGE/DEBIAN/prerm" <<EOF
#!/bin/sh
set -e
SERVICE_NAME="$SERVICE_NAME"
if [ "\$1" = "remove" ] && command -v systemctl >/dev/null 2>&1; then
  systemctl stop "\$SERVICE_NAME.service" 2>/dev/null || true
  systemctl disable "\$SERVICE_NAME.service" 2>/dev/null || true
fi
exit 0
EOF
chmod 0755 "$STAGE/DEBIAN/prerm"

cat >"$STAGE/DEBIAN/postrm" <<EOF
#!/bin/sh
set -e
if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload || true
fi
exit 0
EOF
chmod 0755 "$STAGE/DEBIAN/postrm"

OUT="${DIST_DIR}/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
mkdir -p "$DIST_DIR"
dpkg-deb --root-owner-group --build "$STAGE" "$OUT"
rm -rf "$STAGE"

echo "Package ready: $OUT"
echo "Install with: sudo dpkg -i $OUT"
