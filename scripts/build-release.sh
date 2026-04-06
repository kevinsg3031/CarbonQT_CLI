#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT_DIR/dist"
BINARY="greentrace"

mkdir -p "$OUT_DIR"

TARGETS=(
  "windows/amd64"
  "windows/arm64"
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

for target in "${TARGETS[@]}"; do
  os="${target%/*}"
  arch="${target#*/}"
  ext=""
  if [[ "$os" == "windows" ]]; then
    ext=".exe"
  fi

  out_name="${BINARY}_${os}_${arch}${ext}"
  echo "Building $out_name"
  GOOS="$os" GOARCH="$arch" go build -o "$OUT_DIR/$out_name" .
done

echo "Artifacts in $OUT_DIR"