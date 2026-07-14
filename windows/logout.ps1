<#
    元数智慧科技 · Codex AI Proxy — 一键退出工具 (Windows)
    版本 1.0 | 专有资产 · 未经授权不得传播
#>

$CODEX_HOME = "$env:USERPROFILE\.codex"
$YUANSHU_DIR = "$CODEX_HOME\yuanshu"

Write-Host "`n  ╔═══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║       元数智慧 AI Proxy · 退出            ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════════╝" -ForegroundColor Cyan

# 恢复原始设置
if (Test-Path "$YUANSHU_DIR\backup.config.toml") {
    Write-Host "`n  → 正在恢复原始设置..."
    Copy-Item "$YUANSHU_DIR\backup.config.toml" "$CODEX_HOME\config.toml" -Force
    Write-Host "  ✅ 原始设置已恢复" -ForegroundColor Green
} else {
    Write-Host "`n  → 未找到备份，恢复默认设置..."
    @"
model = "gpt-5.5"
model_reasoning_effort = "medium"
"@ | Set-Content "$CODEX_HOME\config.toml" -Encoding UTF8
    Write-Host "  ✅ 已重置为默认设置" -ForegroundColor Green
}

# 移除 hosts 屏蔽
$hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
$hostsContent = Get-Content $hostsPath -Raw -ErrorAction Ignore
if ($hostsContent -match [regex]::Escape("ab.chatgpt.com")) {
    Write-Host "  → 还原模型列表显示..." -ForegroundColor Gray
    try {
        $newContent = $hostsContent -replace "(?m)^.*ab\.chatgpt\.com.*`r*`n*", ""
        Set-Content -Path $hostsPath -Value $newContent -Force -ErrorAction Stop
        Write-Host "  ✅ 已修复" -ForegroundColor Green
    } catch { Write-Host "  ⚠️  修复失败，但不影响使用" -ForegroundColor Yellow }
}

# 清理临时文件
Remove-Item "$YUANSHU_DIR\metaproxy-models.json" -ErrorAction Ignore
Remove-Item "$YUANSHU_DIR\custom-proxy.config.toml" -ErrorAction Ignore
Remove-Item "$YUANSHU_DIR\custom-proxy-fast.config.toml" -ErrorAction Ignore

Write-Host "`n  ╔═══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║        🎉 已退出，可以放心使用            ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host "`n  所有设置已恢复原样" -ForegroundColor White
Write-Host "  如果想再次使用，双击「login.bat」即可`n" -ForegroundColor Gray
Read-Host "  (按 Enter 键关闭)"
