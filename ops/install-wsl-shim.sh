#!/usr/bin/env bash

set -euo pipefail

readonly REPO_URL="${REPO_URL:-https://github.com/yoonhyunwoo/ipc-bridge.git}"
readonly INSTALL_ROOT="${INSTALL_ROOT:-/opt/leapp-ipc-bridge}"
readonly REPO_DIR="${REPO_DIR:-$INSTALL_ROOT/repo}"
readonly SHIM_DIR="${SHIM_DIR:-$REPO_DIR/shim}"
readonly ENV_FILE="${ENV_FILE:-/etc/default/leapp-ipc-shim}"
readonly SERVICE_PATH="${SERVICE_PATH:-/etc/systemd/system/leapp-ipc-shim.service}"
readonly NODE_BIN="${NODE_BIN:-/usr/bin/node}"
readonly SERVER_ID="${LEAPP_SHIM_SERVER_ID:-leapp_da}"
readonly TCP_HOST="${LEAPP_SHIM_TCP_HOST:-127.0.0.1}"
readonly TCP_PORT="${LEAPP_SHIM_TCP_PORT:-43827}"
readonly LOG_LEVEL="${LEAPP_SHIM_LOG_LEVEL:-info}"

log() {
  printf '[leapp-ipc-shim] %s\n' "$*"
}

fail() {
  printf '[leapp-ipc-shim] ERROR %s\n' "$*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    fail "run this installer as root"
  fi
}

check_node() {
  [[ -x "$NODE_BIN" ]] || fail "node binary not found at $NODE_BIN"

  local version major
  version="$("$NODE_BIN" --version)"
  major="${version#v}"
  major="${major%%.*}"

  if [[ "$major" -lt 22 ]]; then
    fail "Node 22 or newer is required; found $version at $NODE_BIN"
  fi
}

sync_repo() {
  mkdir -p "$INSTALL_ROOT"

  if [[ -d "$REPO_DIR/.git" ]]; then
    log "updating existing ipc-bridge checkout in $REPO_DIR"
    git -C "$REPO_DIR" fetch --depth 1 origin main
    git -C "$REPO_DIR" reset --hard origin/main
  else
    log "cloning ipc-bridge into $REPO_DIR"
    rm -rf "$REPO_DIR"
    git clone --depth 1 "$REPO_URL" "$REPO_DIR"
  fi
}

install_dependencies() {
  log "installing shim dependencies"
  cd "$SHIM_DIR"
  npm ci --omit=dev
}

write_env_file() {
  log "writing environment file $ENV_FILE"
  install -d "$(dirname "$ENV_FILE")"
  cat >"$ENV_FILE" <<EOF
LEAPP_SHIM_SERVER_ID=$SERVER_ID
LEAPP_SHIM_TCP_HOST=$TCP_HOST
LEAPP_SHIM_TCP_PORT=$TCP_PORT
LEAPP_SHIM_LOG_LEVEL=$LOG_LEVEL
EOF
}

write_service() {
  log "writing systemd unit $SERVICE_PATH"
  cat >"$SERVICE_PATH" <<EOF
[Unit]
Description=Leapp WSL IPC shim
After=network.target
Wants=network.target

[Service]
Type=simple
EnvironmentFile=$ENV_FILE
WorkingDirectory=$SHIM_DIR
ExecStart=$NODE_BIN $SHIM_DIR/index.js
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
}

enable_service() {
  log "reloading systemd and enabling service"
  systemctl daemon-reload
  systemctl enable leapp-ipc-shim.service
  systemctl restart leapp-ipc-shim.service
}

main() {
  require_root
  require_command git
  require_command npm
  require_command systemctl
  check_node

  sync_repo
  install_dependencies
  write_env_file
  write_service
  enable_service

  log "installation complete"
  systemctl --no-pager --full status leapp-ipc-shim.service || true
}

main "$@"
