#!/bin/sh
# Cross-compile the alttpr-brick harness for the Trimui Brick (linux/arm64)
# and assemble a ready-to-copy app directory.
#
# Usage:
#   scripts/build-brick.sh           # builds dist/trimui-brick/
#   scripts/build-brick.sh out=/tmp  # builds /tmp/alttpr-brick/

set -eu

repo_root=$(cd "$(dirname "$0")/.." && pwd)
out=${1:-"$repo_root/dist/trimui-brick"}

# 1. Refresh the embedded base patch from storage/patches/.
src_patch="$repo_root/storage/patches/edc01f3db798ae4dfe21101311598d44.json"
embed_dir="$repo_root/internal/patch/all_patches_embed"
embed_patch="$embed_dir/edc01f3db798ae4dfe21101311598d44.json"
mkdir -p "$embed_dir"
if [ ! -f "$src_patch" ]; then
    if [ ! -f "$embed_patch" ]; then
        echo "ERROR: $src_patch missing — run 'php artisan alttp:updatebuildrecord' first." >&2
        exit 1
    fi
    echo "note: $src_patch missing; reusing already-mirrored $embed_patch"
else
    cp -f "$src_patch" "$embed_patch"
fi

# 2. Build.
mkdir -p "$out"
cd "$repo_root"
echo "building: linux/arm64 → $out/alttpr-brick"
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOFLAGS=-mod=mod \
    go build -trimpath -ldflags="-s -w" \
    -o "$out/alttpr-brick" \
    ./cmd/alttpr-brick

# 3. Copy app skeleton (skip if src == dst, e.g. when building in-place
# under dist/trimui-brick).
copy_if_diff() {
    src="$1"; dst="$2"
    if [ "$(readlink -f "$src" 2>/dev/null || echo "$src")" = \
         "$(readlink -f "$dst" 2>/dev/null || echo "$dst")" ]; then
        return 0
    fi
    cp -f "$src" "$dst"
}
copy_if_diff "$repo_root/dist/trimui-brick/launch.sh"           "$out/launch.sh"
copy_if_diff "$repo_root/dist/trimui-brick/settings.json.example" "$out/settings.json.example"
copy_if_diff "$repo_root/dist/trimui-brick/README.txt"          "$out/README.txt"
chmod 0755 "$out/launch.sh" "$out/alttpr-brick"

# 4. Report.
echo
file "$out/alttpr-brick"
echo
ls -la "$out"
echo
echo "Done. Copy '$out' to /mnt/SDCARD/Apps/alttpr/ on your Brick's SD card."
