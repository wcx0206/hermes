#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <linux|darwin|all>"
  exit 1
fi

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="$PROJECT_ROOT/bin"
BINARY_NAME="hermes"

mkdir -p "$OUTPUT_DIR"

build_target() {
  local os="$1"
  GOOS="$os" GOARCH=amd64 go build \
    -o "$OUTPUT_DIR/${BINARY_NAME}-${os}" \
    "$PROJECT_ROOT/cmd"
  echo "Built ${BINARY_NAME}-${os}"
}

case "$1" in
  linux) build_target linux ;;
  darwin) build_target darwin ;;
  all)
    build_target linux
    build_target darwin
    ;;
  *)
    echo "Unsupported target: $1"
    exit 1
    ;;
esac