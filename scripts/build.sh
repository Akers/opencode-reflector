#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== opencode-reflector build ==="

# Build Go binary
echo "[1/2] Building Go binary..."
cd "$PROJECT_ROOT"
go build -ldflags "-s -w -X main.version=${VERSION:-dev}" -o "$PROJECT_ROOT/reflector" ./cmd/reflector/
echo "  ✓ reflector binary: $PROJECT_ROOT/reflector"

# Build TypeScript adapter
echo "[2/2] Building opencode adapter..."
cd "$PROJECT_ROOT/adapters/opencode"
if [ -f "package.json" ]; then
  npm install --silent 2>/dev/null || true
  npx tsc --outDir dist --declaration --sourceMap
  echo "  ✓ adapter compiled: $PROJECT_ROOT/adapters/opencode/dist/"
else
  echo "  - No adapter to build (skipping)"
fi

echo ""
echo "Build complete!"
