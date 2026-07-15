# 元数AI

一键配置 Codex 使用自定义 AI 模型。

## 使用

**macOS**: 下载 `元数AI.dmg` → 安装 → 双击运行

**Windows**: 下载 `yuanshu-ai.exe` → 双击运行

## 从源码构建

```bash
cd client

# macOS
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o yuanshu-ai

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o yuanshu-ai.exe

# 打包 .app
bash build-app.sh

# 打包 .dmg
bash build-dmg.sh
```

## 目录

```
client/
├── main.go           # 入口
├── config.go         # 登录/退出逻辑
├── server.go         # Statsig 加速服务
├── cert.go           # 证书管理
├── hosts.go          # hosts 管理
├── gui.go            # Web GUI
├── web/index.html    # GUI 页面
├── yuanshu-ai.icns   # macOS 图标
├── yuanshu-ai.ico    # Windows 图标
├── build-app.sh      # 构建 .app
├── build-dmg.sh      # 构建 .dmg
└── 元数AI.command     # 双击启动脚本
```
