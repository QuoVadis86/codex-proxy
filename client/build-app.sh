#!/bin/bash
# 构建 macOS .app 包
cd "$(dirname "$0")"
APP="元数AI.app"
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS"
mkdir -p "$APP/Contents/Resources"

# 复制图标
cp codex-proxy.icns "$APP/Contents/Resources/"

cat > "$APP/Contents/Info.plist" << PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>yuanshu-ai</string>
    <key>CFBundleIdentifier</key>
    <string>com.yuanshu.yuanshu-ai</string>
    <key>CFBundleName</key>
    <string>元数AI</string>
    <key>CFBundleDisplayName</key>
    <string>元数AI</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleIconFile</key>
    <string>codex-proxy.icns</string>
</dict>
</plist>
PLIST

cp yuanshu-ai "$APP/Contents/MacOS/"
echo "✅ $APP 已生成"
