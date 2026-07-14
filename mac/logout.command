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
    cp "$YUANSHU_DIR/backup.config.toml" "$CODEX_HOME/config.toml"
else
    cat > "$CODEX_HOME/config.toml" << TOMLCFG
model = "gpt-5.5"
model_reasoning_effort = "medium"
TOMLCFG
fi

# ---------- 清理登录环境（输一次密码） ----------
CMDS=""
if security find-certificate -c "YuanshuStatsigCA" &>/dev/null; then
    CMDS="$CMDS security delete-certificate -c 'YuanshuStatsigCA';"
fi
if grep -q "ab.chatgpt.com" /etc/hosts 2>/dev/null; then
    CMDS="$CMDS sed -i '' '/ab.chatgpt.com/d' /etc/hosts;"
fi
if [ -n "$CMDS" ]; then
echo "  → 需要输入电脑密码..."
SCRIPT="do shell script \"$CMDS\" with administrator privileges"
osascript -e "$SCRIPT" 2>/dev/null || echo "  ⚠️  清理失败"
fi

# 清理临时文件
rm -f "$YUANSHU_DIR/metaproxy-models.json"
rm -f "$YUANSHU_DIR/custom-proxy.config.toml"
rm -f "$YUANSHU_DIR/custom-proxy-fast.config.toml"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║           🎉 已退出                       ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""
echo "  (按 Enter 键关闭)"
read
