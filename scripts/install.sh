#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== opencode-reflector install ==="

# Determine install paths
REFLECTOR_BIN="${REFLECTOR_BIN:-$HOME/.local/bin/reflector}"
OPENCODE_DIR="${OPENCODE_DIR:-$HOME/.config/opencode}"
ADAPTER_TARGET="$OPENCODE_DIR/plugins/opencode-reflector"

# Step 1: Build
echo "[1/3] Building..."
bash "$SCRIPT_DIR/build.sh"

# Step 2: Install binary
echo "[2/3] Installing binary to $REFLECTOR_BIN..."
mkdir -p "$(dirname "$REFLECTOR_BIN")"
cp "$PROJECT_ROOT/reflector" "$REFLECTOR_BIN"
chmod +x "$REFLECTOR_BIN"
echo "  ✓ Binary installed"

# Step 3: Install adapter as opencode plugin
echo "[3/3] Installing opencode adapter plugin..."
mkdir -p "$ADAPTER_TARGET"
cp -r "$PROJECT_ROOT/adapters/opencode/dist/" "$ADAPTER_TARGET/dist/"
cp "$PROJECT_ROOT/adapters/opencode/package.json" "$ADAPTER_TARGET/"
echo "  ✓ Adapter installed to $ADAPTER_TARGET"

# Step 4: Initialize .reflector in current directory
REFLECTOR_DIR="${REFLECTOR_DIR:-$PROJECT_ROOT/.reflector}"
if [ ! -d "$REFLECTOR_DIR" ]; then
  echo ""
  echo "[extra] Initializing .reflector directory..."
  "$REFLECTOR_BIN" -data-dir "$REFLECTOR_DIR" -config "$PROJECT_ROOT/reflector.yaml" &
  PID=$!
  sleep 1
  kill $PID 2>/dev/null || true
  echo "  ✓ .reflector directory created"
fi

echo ""
echo "Install complete!"
echo ""
echo "Usage:"
echo "  reflector                          # Start with defaults"
echo "  reflector -config ./reflector.yaml # Custom config"
echo "  reflector -data-dir .reflector     # Custom data dir"
