@echo off
chcp 65001 >nul
powershell -ExecutionPolicy Bypass -File "%~dp0logout.ps1"
pause
