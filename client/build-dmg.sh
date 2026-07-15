#!/bin/bash
# 构建 macOS DMG 安装包
cd "$(dirname "$0")"

APP="codex-proxy.app"
DMG="Codex-AI-Proxy.dmg"
VOLUME="元数智慧·Codex"

# 确保 .app 存在
if [ ! -d "$APP" ]; then
    echo "请先运行 build-app.sh"
    exit 1
fi

# 创建临时目录
TMP_DIR=$(mktemp -d)
cp -R "$APP" "$TMP_DIR/"

# 创建 Applications 快捷方式
ln -s /Applications "$TMP_DIR/Applications"

# 创建 DMG
hdiutil create -volname "$VOLUME" -srcfolder "$TMP_DIR" \
    -ov -format UDZO -size 100m "$DMG" 2>&1 | tail -3

# 清理
rm -rf "$TMP_DIR"

echo "✅ $DMG 已生成"
ls -lh "$DMG"
