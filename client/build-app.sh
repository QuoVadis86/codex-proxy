#!/bin/bash
# 构建 macOS .app 包
cd "$(dirname "$0")"
rm -rf codex-proxy.app
mkdir -p codex-proxy.app/Contents/MacOS
mkdir -p codex-proxy.app/Contents/Resources

# 复制图标
cp codex-proxy.icns codex-proxy.app/Contents/Resources/

cat > codex-proxy.app/Contents/Info.plist << PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>codex-proxy-mac</string>
    <key>CFBundleIdentifier</key>
    <string>com.yuanshu.codex-proxy</string>
    <key>CFBundleName</key>
    <string>Codex AI Proxy</string>
    <key>CFBundleDisplayName</key>
    <string>Codex AI Proxy</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleIconFile</key>
    <string>codex-proxy.icns</string>
</dict>
</plist>
PLIST

cp codex-proxy-mac codex-proxy.app/Contents/MacOS/
echo "✅ codex-proxy.app 已生成"
