#!/bin/bash
# ============================================================
#  元数智慧科技 · Codex AI Proxy — 一键退出
# ============================================================

CODEX_HOME="$HOME/.codex"
YUANSHU_DIR="$CODEX_HOME/yuanshu"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║       元数智慧 AI Proxy · 退出            ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""

if [ -f "$YUANSHU_DIR/backup.config.toml" ]; then
    cp "$YUANSHU_DIR/backup.config.toml" "$CODEX_HOME/config.toml"
    echo "  ✅ 已恢复原始配置"
else
    rm -f "$CODEX_HOME/config.toml"
    echo "  ✅ 已清除配置"
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
