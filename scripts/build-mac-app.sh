#!/usr/bin/env bash
set -euo pipefail

APP_NAME="Plop"
BUNDLE_DIR="out/${APP_NAME}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
BIN_NAME="plop"

VERSION="${VERSION:-}"
if [[ -z "${VERSION}" ]]; then
  VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
fi

mkdir -p "${MACOS_DIR}" "${RESOURCES_DIR}"

# Build the app binary inside the .app bundle.
go build \
  -tags noassets \
  -ldflags "-X github.com/alex-vit/plop/cmd.Version=${VERSION}" \
  -o "${MACOS_DIR}/${BIN_NAME}" \
  .

chmod +x "${MACOS_DIR}/${BIN_NAME}"

if [[ -f "icon/icon.png" ]]; then
  cp "icon/icon.png" "${RESOURCES_DIR}/AppIcon.png"
fi

cat > "${CONTENTS_DIR}/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>
  <string>${APP_NAME}</string>
  <key>CFBundleDisplayName</key>
  <string>${APP_NAME}</string>
  <key>CFBundleIdentifier</key>
  <string>com.alexvit.plop</string>
  <key>CFBundleVersion</key>
  <string>${VERSION}</string>
  <key>CFBundleShortVersionString</key>
  <string>${VERSION}</string>
  <key>CFBundleExecutable</key>
  <string>${BIN_NAME}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleIconFile</key>
  <string>AppIcon</string>
  <key>LSMinimumSystemVersion</key>
  <string>13.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
PLIST

echo "Built macOS app bundle: ${BUNDLE_DIR}"
echo "Open it with: open ${BUNDLE_DIR}"
