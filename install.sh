#!/usr/bin/env bash
set -euo pipefail

REPO="wcx0206/hermes"
BIN_DIR="${HOME}/.local/bin"
CLI_NAME="hermes"
SERVER_NAME="hermes-backup"

mkdir -p "${BIN_DIR}"

download() {
  local bin="$1" os="$2" arch="$3"
  local url="https://github.com/${REPO}/releases/latest/download/${bin}-${os}-${arch}"
  echo "Downloading ${url}"
  curl -fsSL "${url}" -o "${BIN_DIR}/${bin}"
  chmod +x "${BIN_DIR}/${bin}"
}

os="$(uname | tr '[:upper:]' '[:lower:]')"
arch=$(uname -m)
echo "Downloading hermes from GitHub..."
download "${CLI_NAME}" "${os}" "${arch}"
download "${SERVER_NAME}" "${os}" "${arch}"

echo "Installed ${CLI_NAME} & ${SERVER_NAME} to ${BIN_DIR}"
echo "Add 'export PATH=\$PATH:${BIN_DIR}' to your shell profile if needed."
