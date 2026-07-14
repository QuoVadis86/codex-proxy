#!/bin/bash
set -e

# ============================================================
#  元数智慧科技 · Codex AI Proxy — 一键退出工具
#  版本 1.0 | 专有资产 · 未经授权不得传播
# ============================================================

CODEX_HOME="$HOME/.codex"
YUANSHU_DIR="$CODEX_HOME/yuanshu"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║       元数智慧 AI Proxy · 退出            ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""

# 恢复原始设置
if [ -f "$YUANSHU_DIR/backup.config.toml" ]; then
    echo "  → 正在恢复原始设置..."
    cp "$YUANSHU_DIR/backup.config.toml" "$CODEX_HOME/config.toml"
    echo "  ✅ 原始设置已恢复"
else
    echo "  → 未找到备份，恢复默认设置..."
    cat > "$CODEX_HOME/config.toml" << TOMLCFG
model = "gpt-5.5"
model_reasoning_effort = "medium"
TOMLCFG
    echo "  ✅ 已重置为默认设置"
fi

# ---------- 删除 SSL 证书 ----------
if security find-certificate -c "YuanshuStatsigCA" &>/dev/null; then
    echo "  → 删除 SSL 证书（需要输入电脑密码）..."
    SCRIPT="do shell script \"security delete-certificate -c 'YuanshuStatsigCA'\" with administrator privileges"
    osascript -e "$SCRIPT" 2>/dev/null
fi

# 还原 hosts
if grep -q "ab.chatgpt.com" /etc/hosts 2>/dev/null; then
    echo "  → 恢复 hosts（需要输入电脑密码）..."
    SCRIPT="do shell script \"sed -i '' '/ab.chatgpt.com/d' /etc/hosts\" with administrator privileges"
    osascript -e "$SCRIPT" 2>/dev/null && echo "  ✅ 已恢复" || echo "  ⚠️  恢复失败，但不影响使用"
fi

# 清理临时文件
rm -f "$YUANSHU_DIR/metaproxy-models.json"
rm -f "$YUANSHU_DIR/custom-proxy.config.toml"
rm -f "$YUANSHU_DIR/custom-proxy-fast.config.toml"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║        🎉 已退出，可以放心使用            ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""
echo "  所有设置已恢复原样"
echo "  如果想再次使用，双击「login.command」即可"
echo ""
echo "  (按 Enter 键关闭)"
read
