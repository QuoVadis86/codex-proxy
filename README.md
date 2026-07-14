# 元数智慧 · Codex AI Proxy

一键配置 Codex 使用自定义 AI 模型（DeepSeek / Qwen），并修复界面语言为中文。

## 使用方法

### macOS

1. 把 `mac/login.command` 发给用户
2. 双击，输入 API Key
3. 弹密码时点确认（装证书 + 改 hosts）

退出时双击 `mac/logout.command` 即可还原。

### 服务器部署

```bash
docker load -i server/statsig-proxy.tar
docker run -d --restart always -p 443:443 --name statsig-proxy statsig-proxy:1.0
```

同时需要在 80 端口提供 `/ca.crt`（供 login.command 下载）。

### 构建

```bash
cd server
go build -ldflags="-s -w" -o statsig-proxy main.go
upx --best statsig-proxy -o statsig-proxy-compressed
docker build -t statsig-proxy:1.0 .
```

## 目录结构

```
codex-proxy/
├── mac/
│   ├── login.command        # 给用户的登录脚本（唯一需要的文件）
│   └── logout.command       # 给用户的退出脚本
├── server/
│   ├── main.go              # Go 源码
│   ├── Dockerfile           # 构建文件
│   ├── statsig-proxy.tar    # 预编译镜像 (2.3MB)
│   ├── ca.crt               # CA 证书（服务器提供下载）
│   ├── server.crt           # 服务器证书
│   └── server.key           # 服务器私钥
└── README.md
```
