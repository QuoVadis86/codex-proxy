#!/bin/bash
set -e

# ============================================================
#  元数智慧科技 · Codex AI Proxy — 一键登录工具
#  版本 1.0 | 专有资产 · 未经授权不得传播
# ============================================================

PROXY_URL="http://113.90.157.107:8317/v1"
STATSIG_SERVER="94.191.115.90"
CODEX_HOME="$HOME/.codex"
YUANSHU_DIR="$CODEX_HOME/yuanshu"
mkdir -p "$YUANSHU_DIR"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║       欢迎使用元数智慧 AI Proxy           ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""

# 检测是否已经登录过了
if grep -q "model_provider.*custom-proxy" "$CODEX_HOME/config.toml" 2>/dev/null; then
    echo "  ✅ 你已经登录过了，不用重复操作"
    echo ""
    echo "  (按 Enter 键关闭)"
    read
    exit 0
fi

# ---------- 输入 Key ----------
if [ -z "$CODEX_API_KEY" ]; then
    read -s -p "  请输入你的 API Key: " CODEX_API_KEY
    echo ""
fi

if [ -z "$CODEX_API_KEY" ]; then
    echo "  ❌ 没有 API Key 无法使用，请联系管理员获取"
    exit 1
fi

# ---------- 连接服务器 ----------
echo ""
echo "  → 正在连接服务器..."
MODELS=$(curl -s --max-time 10 "${PROXY_URL}/models" \
    -H "Authorization: Bearer ${CODEX_API_KEY}" \
    | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    models = [m['id'] for m in data.get('data', []) if m['id'] != 'wan2.7-image']
    print('\n'.join(models))
except: pass
" 2>/dev/null) || MODELS=""

if [ -z "$MODELS" ]; then
    echo "  ⚠️  连接失败，请检查 API Key 是否正确"
    MODELS="deepseek-v4-flash
deepseek-v4-pro"
else
    echo "  ✅ 连接成功！共 $(echo "$MODELS" | wc -l | tr -d ' ') 个模型可用"
fi

# ---------- 备份原始设置 ----------
if [ -f "$CODEX_HOME/config.toml" ] && ! grep -q "model_provider.*custom-proxy" "$CODEX_HOME/config.toml" 2>/dev/null; then
    cp "$CODEX_HOME/config.toml" "$YUANSHU_DIR/backup.config.toml"
    echo -n ""
elif [ ! -f "$YUANSHU_DIR/backup.config.toml" ] && [ ! -f "$CODEX_HOME/config.toml" ]; then
    cat > "$YUANSHU_DIR/backup.config.toml" << TOMLCFG
model = "gpt-5.5"
model_reasoning_effort = "medium"
TOMLCFG
fi

# ---------- 写配置 ----------
python3 << PYEOF
import json, os, re
YUANSHU_DIR = os.path.join(os.path.expanduser("$CODEX_HOME"), "yuanshu")

def slug_to_display(slug):
    known = {"deepseek":"DeepSeek","gpt":"GPT","qwen":"Qwen","codex":"Codex","glm":"GLM","kimi":"Kimi","claude":"Claude","gemini":"Gemini","mistral":"Mistral","llama":"Llama","yi":"Yi","moonshot":"Moonshot"}
    parts = slug.replace("-"," ").split()
    r = []
    for p in parts:
        l = p.lower()
        if l in known: r.append(known[l])
        elif re.match(r'^[\d.]+[a-z]?$', l): r.append(p.upper() if p.isupper() else p)
        else: r.append(p.capitalize())
    return " ".join(r)

GPT = [{"effort":"low","description":"Fast responses with lighter reasoning"},{"effort":"medium","description":"Balances speed and reasoning depth for everyday tasks"},{"effort":"high","description":"Greater reasoning depth for complex problems"},{"effort":"xhigh","description":"Extra high reasoning depth for complex problems"},{"effort":"max","description":"Maximum reasoning depth for the hardest problems"},{"effort":"ultra","description":"Maximum reasoning with automatic task delegation"}]
DS = [{"effort":"low","description":"Fast responses (mapped to High on server)"},{"effort":"medium","description":"Balanced (mapped to High on server)"},{"effort":"high","description":"Step-by-step reasoning (default)"},{"effort":"xhigh","description":"Extra deep reasoning (mapped to Max on server)"},{"effort":"max","description":"Maximum reasoning effort"}]
DEF = [{"effort":"low","description":"Fast responses with lighter reasoning"},{"effort":"medium","description":"Balances speed and reasoning depth"},{"effort":"high","description":"Greater reasoning depth for complex problems"}]
def lvl(s):
    s = s.lower()
    if s.startswith("gpt"): return GPT
    if s.startswith("deepseek"): return DS
    return DEF

T = {"shell_type":"shell_command","visibility":"list","supported_in_api":True,"additional_speed_tiers":[],"service_tiers":[],"availability_nux":None,"upgrade":None,"base_instructions":"You are Codex, a coding agent.","model_messages":{},"include_skills_usage_instructions":True,"supports_reasoning_summaries":True,"default_reasoning_summary":"none","support_verbosity":True,"default_verbosity":"low","apply_patch_tool_type":"freeform","web_search_tool_type":"text_and_image","truncation_policy":{"mode":"tokens","limit":10000},"supports_parallel_tool_calls":True,"supports_image_detail_original":True,"context_window":1000000,"max_context_window":1000000,"effective_context_window_percent":95,"experimental_supported_tools":[],"input_modalities":["text","image"],"supports_search_tool":True,"use_responses_lite":False,"default_reasoning_level":"medium"}
entries = []
for i, slug in enumerate([m for m in """$MODELS""".strip().split('\n') if m.strip()]):
    e = json.loads(json.dumps(T))
    e["slug"]=slug; e["display_name"]=slug_to_display(slug); e["description"]=f"{slug_to_display(slug)} via proxy"; e["supported_reasoning_levels"]=lvl(slug); e["priority"]=i+1
    if "qwen" in slug.lower(): e["context_window"]=131072; e["max_context_window"]=131072
    entries.append(e)
with open(os.path.join(YUANSHU_DIR, "metaproxy-models.json"), 'w') as f:
    json.dump({"models":entries}, f, indent=2, ensure_ascii=False)
PYEOF

FIRST_MODEL=$(echo "$MODELS" | head -1)

cat > "$CODEX_HOME/config.toml" << TOMLCFG
model = "$FIRST_MODEL"
model_provider = "custom-proxy"
model_reasoning_effort = "medium"
model_catalog_json = "$YUANSHU_DIR/metaproxy-models.json"

[model_providers.custom-proxy]
name = "元数智慧 · Codex AI Proxy"
base_url = "${PROXY_URL}"
experimental_bearer_token = "${CODEX_API_KEY}"
requires_openai_auth = false
TOMLCFG

cat > "$YUANSHU_DIR/custom-proxy.config.toml" << TOMLCFG
model = "$FIRST_MODEL"
model_provider = "custom-proxy"
model_reasoning_effort = "high"
TOMLCFG

cat > "$YUANSHU_DIR/custom-proxy-fast.config.toml" << TOMLCFG
model = "$FIRST_MODEL"
model_provider = "custom-proxy"
model_reasoning_effort = "low"
TOMLCFG

# ---------- 配置登录环境（输一次密码） ----------
CA_CERT_DST="$CODEX_HOME/yuanshu/statsig-server/ca.crt"
mkdir -p "$(dirname "$CA_CERT_DST")"
curl -sk --max-time 5 "https://${STATSIG_SERVER}:8318/ca.crt" -o "$CA_CERT_DST" 2>/dev/null

CMDS=""
if [ -f "$CA_CERT_DST" ] && [ -s "$CA_CERT_DST" ]; then
    CMDS="$CMDS security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '$CA_CERT_DST';"
fi
if ! grep -q "ab.chatgpt.com" /etc/hosts 2>/dev/null; then
    CMDS="$CMDS echo '${STATSIG_SERVER} ab.chatgpt.com' >> /etc/hosts && echo '::1 ab.chatgpt.com' >> /etc/hosts;"
fi
if [ -n "$CMDS" ]; then
    echo "  → 需要输入电脑密码..."
SCRIPT="do shell script \"$CMDS\" with administrator privileges"
osascript -e "$SCRIPT" 2>/dev/null || echo "  ⚠️  配置失败"
fi

# ---------- 完成 ----------
echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║           🎉 登录成功                     ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""
echo "  可用模型:"
echo "$MODELS" | while read m; do
    [ -n "$m" ] && echo "    • $m"
done
echo ""
echo "  (按 Enter 键关闭)"
read
