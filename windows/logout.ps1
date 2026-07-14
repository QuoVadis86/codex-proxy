<#
    元数智慧科技 · Codex AI Proxy — 一键退出工具 (Windows)
    版本 1.0 | 专有资产 · 未经授权不得传播
#>

$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

$CODEX_HOME = "$env:USERPROFILE\.codex"
$YUANSHU_DIR = "$CODEX_HOME\yuanshu"

Write-Host "`n  ╔═══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║       元数智慧 AI Proxy · 退出            ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════════╝" -ForegroundColor Cyan

# 恢复原始设置
if (Test-Path "$YUANSHU_DIR\backup.config.toml") {
    Copy-Item "$YUANSHU_DIR\backup.config.toml" "$CODEX_HOME\config.toml" -Force
} else {
    @"
model = "gpt-5.5"
model_reasoning_effort = "medium"
"@ | Set-Content "$CODEX_HOME\config.toml" -Encoding UTF8
}

# 清理系统配置
$hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
$needAdmin = $false
if (Test-Path "$YUANSHU_DIR\ca.crt") { $needAdmin = $true }
if (Select-String -Path $hostsPath -Pattern "ab.chatgpt.com" -Quiet -ErrorAction Ignore) { $needAdmin = $true }

if ($needAdmin) {
    if (-not $isAdmin) {
        Write-Host "  → 正在恢复原始账号..."
        Start-Process powershell -Verb RunAs -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`""
        Read-Host "`n  (按 Enter 键关闭)"
        exit 0
    }
    Write-Host "  → 正在恢复原始账号..."
    certutil -delstore Root "YuanshuStatsigCA" | Out-Null
    $hostsContent = Get-Content $hostsPath -Raw -ErrorAction Ignore
    if ($hostsContent -match [regex]::Escape("ab.chatgpt.com")) {
        $newContent = $hostsContent -replace "(?m)^.*ab\.chatgpt\.com.*`r*`n*", ""
        Set-Content -Path $hostsPath -Value $newContent -Force
    }
}

# 清理临时文件
Remove-Item "$YUANSHU_DIR\metaproxy-models.json" -ErrorAction Ignore
Remove-Item "$YUANSHU_DIR\custom-proxy.config.toml" -ErrorAction Ignore
Remove-Item "$YUANSHU_DIR\custom-proxy-fast.config.toml" -ErrorAction Ignore
Remove-Item "$YUANSHU_DIR\ca.crt" -ErrorAction Ignore

Write-Host "`n  ╔═══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║           🎉 已退出                       ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════════╝" -ForegroundColor Cyan
Read-Host "`n  (按 Enter 键关闭)"
