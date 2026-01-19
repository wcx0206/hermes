#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <backup|cli> <linux|darwin|all>"
  exit 1
fi

component="$1"
target="$2"

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="$PROJECT_ROOT/bin"

case "$component" in
  backup)
    PACKAGE="$PROJECT_ROOT/cmd/backupd"
    BINARY_NAME="hermes-backup"
    ;;
  cli)
    PACKAGE="$PROJECT_ROOT/cmd/cli"
    BINARY_NAME="hermes-cli"
    ;;
  *)
    echo "Unsupported component: $component"
    exit 1
    ;;
esac

mkdir -p "$OUTPUT_DIR"

build_target() {
  local os="$1"
  GOOS="$os" GOARCH=amd64 go build \
    -o "$OUTPUT_DIR/${BINARY_NAME}-${os}" \
    "$PACKAGE"
  echo "Built ${BINARY_NAME}-${os}"
}

case "$target" in
  linux) build_target linux ;;
  darwin) build_target darwin ;;
  all)
    build_target linux
    build_target darwin
    ;;
  *)
    echo "Unsupported target: $target"
    exit 1
    ;;
esac