#!/bin/bash
set -e

CODEX_HOME="$HOME/.codex"
YUANSHU_DIR="$CODEX_HOME/yuanshu"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║       元数智慧 AI Proxy · 退出            ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""

# 恢复配置
if [ -f "$YUANSHU_DIR/backup.config.toml" ]; then
    cp "$YUANSHU_DIR/backup.config.toml" "$CODEX_HOME/config.toml"
else
    rm -f "$CODEX_HOME/config.toml"
fi

# 删 hosts
if grep -q "ab.chatgpt.com" /etc/hosts 2>/dev/null; then
    echo "  → 需要输入电脑密码..."
    SCRIPT="do shell script \"sed -i '' '/ab.chatgpt.com/d' /etc/hosts\" with administrator privileges"
    osascript -e "$SCRIPT" 2>/dev/null || true
fi

rm -f "$YUANSHU_DIR/metaproxy-models.json"
rm -f "$YUANSHU_DIR/custom-proxy.config.toml"
rm -f "$YUANSHU_DIR/custom-proxy-fast.config.toml"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║           🎉 已退出                       ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""
read -p "  (按 Enter 键关闭)"
