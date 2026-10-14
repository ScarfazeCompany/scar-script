#!/usr/bin/env bash
# Build ScarScript compiler for Windows, Linux, and macOS
# Usage: ./scripts/build.sh [version]

set -euo pipefail

VERSION="${1:-1.0.0}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$REPO_ROOT/dist"
CMD_PATH="$REPO_ROOT/cmd/scar"

# Clean
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

LDFLAGS="-s -w -X main.version=$VERSION"

targets=(
    "windows amd64 .exe scar-windows-amd64"
    "linux   amd64 ''   scar-linux-amd64"
    "darwin  amd64 ''   scar-darwin-amd64"
    "darwin  arm64 ''   scar-darwin-arm64"
)

echo "Building ScarScript v$VERSION..."

for target in "${targets[@]}"; do
    read -r GOOS GOARCH EXT NAME <<< "$target"
    OUT_NAME="${NAME}${EXT}"
    OUT_PATH="$DIST_DIR/$OUT_NAME"
    
    echo -n "  Building $OUT_NAME... "
    
    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 \
        go build -ldflags "$LDFLAGS" -o "$OUT_PATH" "$CMD_PATH"
    
    if [[ $? -ne 0 ]]; then
        echo "FAILED"
        exit 1
    fi
    echo "OK"
done

# Generate checksums
echo ""
echo "Generating checksums..."
cd "$DIST_DIR"
sha256sum scar-* > checksums.txt
echo "  checksums.txt created"

echo ""
echo "Done! Binaries in $DIST_DIR"
ls -lh "$DIST_DIR"