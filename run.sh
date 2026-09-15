#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1
chmod +x "$SCRIPT_DIR/jee-anki-linux-amd64" 2>/dev/null || true
exec "$SCRIPT_DIR/jee-anki-linux-amd64" "$@"
