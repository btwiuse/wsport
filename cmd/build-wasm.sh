#!/bin/bash
set -euo pipefail

for d in */main.go; do
  name=$(dirname "$d")
  echo "==> Building $name"
  GOOS=js GOARCH=wasm go build -o "build/$name.wasm" "./$name/"
  size=$(du -sh "$name.wasm" | cut -f1)
  echo "    $size  $name.wasm"
done
