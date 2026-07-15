#!/bin/bash
cd "$(dirname "$0")"
APP="元数AI.app"
DMG="元数AI.dmg"
VOLUME="元数AI"

if [ ! -d "$APP" ]; then
    echo "请先运行 build-app.sh"
    exit 1
fi

TMP_DIR=$(mktemp -d)
cp -R "$APP" "$TMP_DIR/"
ln -s /Applications "$TMP_DIR/Applications"

hdiutil create -volname "$VOLUME" -srcfolder "$TMP_DIR" \
    -ov -format UDZO -size 100m "$DMG" 2>&1 | tail -3
rm -rf "$TMP_DIR"
echo "✅ $DMG 已生成"
ls -lh "$DMG"
