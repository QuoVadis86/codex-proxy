# 元数智慧 · Codex AI Proxy

一键配置 Codex 使用自定义 AI 模型（DeepSeek / Qwen），并修复界面语言为中文。

## 目录结构

```
codex-proxy/
├── mac/
│   ├── login.command        # macOS 一键登录（可双击）
│   └── logout.command       # macOS 一键退出
├── windows/
│   ├── login.bat            # Windows 一键登录
│   ├── login.ps1
│   ├── logout.bat           # Windows 一键退出
│   └── logout.ps1
├── server/
│   ├── main.go              # Statsig 代理服务器源码
│   ├── Dockerfile           # Docker 构建文件
│   ├── statsig-proxy.tar    # 预编译镜像 (4.98MB)
│   ├── ca.crt               # CA 证书（用户安装用）
│   ├── server.crt           # 服务器证书
│   └── server.key           # 服务器私钥（已 gitignore）
└── README.md
```

## 快速开始

### macOS

双击 `mac/login.command`，输入 API Key 即可。

### 服务器部署

```bash
docker load -i server/statsig-proxy.tar
docker run -d --restart always -p 443:443 --name statsig-proxy statsig-proxy:1.0
```

## 原理

- 修改 hosts 将 `ab.chatgpt.com` 指向代理服务器
- 代理服务器转发 `/v1/initialize` 到真实 Statsig，剥离模型限制
- 自动安装 CA 证书使 HTTPS 握手通过

## 构建

```bash
cd server
go build -ldflags="-s -w" -o statsig-proxy main.go
upx --best statsig-proxy -o statsig-proxy-compressed
docker build -t statsig-proxy:1.0 .
```

## 许可证

专有资产 · 未经授权不得传播
