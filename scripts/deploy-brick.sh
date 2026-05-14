#!/bin/sh
# Build the brick harness and push it to a connected Trimui Brick over adb.
#
# Usage:
#   scripts/deploy-brick.sh            # build + push
#   scripts/deploy-brick.sh --no-build # skip the build step, push current dist/
#   scripts/deploy-brick.sh --serial X # target a specific device when multiple
#
# What lands on the device (under /mnt/SDCARD/Apps/alttpr/):
#   alttpr-brick           — arm64 binary
#   config.json            — stock Trimui Apps-menu integration
#   launch.sh              — invocation script
#   icon.png               — menu icon
#   README.txt             — on-device install/usage notes
#   settings.json.example  — example user-prefs file
#   settings.json          — only if missing on device (preserves your prefs)
#
# The user's base.sfc is never touched.

set -eu

repo_root=$(cd "$(dirname "$0")/.." && pwd)
dist="$repo_root/dist/trimui-brick"
remote="/mnt/SDCARD/Apps/alttpr"

build=1
serial=""
while [ $# -gt 0 ]; do
    case "$1" in
        --no-build) build=0 ;;
        --serial)   shift; serial="$1" ;;
        -h|--help)
            sed -n '2,/^set/p' "$0" | sed 's/^# \{0,1\}//;/^set/d'
            exit 0
            ;;
        *)
            echo "unknown flag: $1" >&2
            exit 2
            ;;
    esac
    shift
done

if [ "$build" = "1" ]; then
    "$repo_root/scripts/build-brick.sh"
fi

# Sanity-check the dist is complete.
missing=
for f in alttpr-brick config.json launch.sh README.txt settings.json.example icon.png; do
    [ -e "$dist/$f" ] || missing="$missing $f"
done
if [ -n "$missing" ]; then
    echo "ERROR: dist is missing:$missing" >&2
    echo "Run scripts/build-brick.sh first." >&2
    exit 1
fi

# Resolve adb binary (allow override via $ADB).
ADB=${ADB:-adb}
if ! command -v "$ADB" >/dev/null 2>&1; then
    echo "ERROR: '$ADB' not found in PATH. Install android-tools / android-platform-tools." >&2
    exit 1
fi

adb() {
    if [ -n "$serial" ]; then
        "$ADB" -s "$serial" "$@"
    else
        "$ADB" "$@"
    fi
}

# Confirm exactly one device is available (unless --serial was given).
if [ -z "$serial" ]; then
    devs=$("$ADB" devices | awk 'NR>1 && $2=="device" {print $1}')
    count=$(printf '%s\n' "$devs" | grep -c . || true)
    if [ "$count" = "0" ]; then
        echo "ERROR: no adb devices found. Plug in the Brick and enable adb in stock OS settings." >&2
        exit 1
    fi
    if [ "$count" != "1" ]; then
        echo "ERROR: multiple adb devices visible — pass --serial <id>:" >&2
        printf '  %s\n' $devs >&2
        exit 1
    fi
fi

echo "→ target: $(adb get-serialno) → $remote"
adb shell "mkdir -p $remote" >/dev/null

# Files that always get overwritten (the app itself + menu integration).
for f in alttpr-brick config.json launch.sh icon.png README.txt settings.json.example; do
    printf '  push  %-26s' "$f"
    adb push -p "$dist/$f" "$remote/$f" >/dev/null 2>&1
    printf 'ok\n'
done

# settings.json: only push if the device doesn't already have one.
# The user's last_options/base_rom path lives there; we don't want to
# clobber it on a redeploy.
if adb shell "[ -f $remote/settings.json ]" >/dev/null 2>&1; then
    printf '  keep  settings.json              (already on device)\n'
else
    printf '  push  settings.json              '
    adb push -p "$dist/settings.json.example" "$remote/settings.json" >/dev/null 2>&1
    printf 'ok (seeded from example)\n'
fi

# Belt-and-suspenders: ensure executable bits stick (adb push preserves
# mode with -p, but some firmwares mount with restrictive umask).
adb shell "chmod 0755 $remote/launch.sh $remote/alttpr-brick" >/dev/null 2>&1 || true
adb shell "sync" >/dev/null 2>&1 || true

# Friendly reminder if the user hasn't dropped a base ROM yet.
if ! adb shell "[ -f $remote/base.sfc ]" >/dev/null 2>&1; then
    cat <<EOF

note: $remote/base.sfc is not present on the device.
Push your legally-owned ALttP base ROM with:

    $ADB push base.sfc $remote/base.sfc

EOF
fi

echo "done. Reboot the Brick (or refresh the Apps menu) and launch alttpr."
