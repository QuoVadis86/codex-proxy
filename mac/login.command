#!/bin/bash
# ============================================================
#  元数智慧科技 · Codex AI Proxy — 一键登录
#  只需要 API Key，自动配置自定义模型
# ============================================================

set -e

PROXY_URL="http://113.90.157.107:8317/v1"
CODEX_HOME="$HOME/.codex"
YUANSHU_DIR="$CODEX_HOME/yuanshu"
mkdir -p "$YUANSHU_DIR"

echo ""
echo "  ╔═══════════════════════════════════════════╗"
echo "  ║       欢迎使用元数智慧 AI Proxy           ║"
echo "  ╚═══════════════════════════════════════════╝"
echo ""

# 检测是否已登录
if grep -q "custom-proxy" "$CODEX_HOME/config.toml" 2>/dev/null; then
    echo "  ✅ 已经登录过了"
    read -p "  (按 Enter 键关闭)"
    exit 0
fi

# 输入 API Key
read -s -p "  请输入你的 API Key: " CODEX_API_KEY
echo ""

# 连接服务器拿模型列表
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
    echo "  ⚠️  连接失败，请检查 API Key"
    MODELS="deepseek-v4-flash deepseek-v4-pro"
else
    echo "  ✅ 连接成功！共 $(echo "$MODELS" | wc -l | tr -d ' ') 个模型"
fi

# 备份原始配置
if [ -f "$CODEX_HOME/config.toml" ] && ! grep -q "custom-proxy" "$CODEX_HOME/config.toml" 2>/dev/null; then
    cp "$CODEX_HOME/config.toml" "$YUANSHU_DIR/backup.config.toml"
fi

# 生成模型配置
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

GPT = [{"effort":"low","description":"Fast"},{"effort":"medium","description":"Balanced"},{"effort":"high","description":"Deep"},{"effort":"xhigh","description":"Extra deep"},{"effort":"max","description":"Max"},{"effort":"ultra","description":"Ultra"}]
DS = [{"effort":"low","description":"Fast"},{"effort":"medium","description":"Balanced"},{"effort":"high","description":"Default"},{"effort":"xhigh","description":"Extra deep"},{"effort":"max","description":"Max"}]
DEF = [{"effort":"low","description":"Fast"},{"effort":"medium","description":"Balanced"},{"effort":"high","description":"Deep"}]

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

# 写入配置
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
read -p "  (按 Enter 键关闭)"
