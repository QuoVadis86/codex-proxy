# 元数AI

一键配置 Codex / ChatGPT Desktop 使用自定义 AI 模型（DeepSeek、Qwen 等）。

## 工作原理

```
用户双击 → 浏览器界面打开
    ↓
输入 API Key → 连接代理服务器获取模型列表
    ↓
1. 安装 CA 证书（MITM 代理需要，仅首次弹密码）
2. 设置 PAC 代理（仅拦截 ab.chatgpt.com）
3. 写入 Codex 配置（合并到原配置，不覆盖）
    ↓
打开 ChatGPT → 自定义模型可用 ✅
ChatGPT 对话历史 → 保留 ✅
```

### 架构

```
ChatGPT/Codex 请求
    │
    ├── ab.chatgpt.com ──→ 系统 PAC ──→ 本地 goproxy (:9090)
    │                                           │
    │                                     MITM 拦截 /v1/initialize
    │                                     删除 available_models 等限制字段
    │                                           │
    │                                    转发到真实 Statsig 服务器
    │
    └── api.openai.com ──→ openai_base_url ──→ 代理服务器
                                                │
                                         转发到模型 API
```

**关键点：**

| 原理 | 说明 |
|---|---|
| 自定义模型列表 | `model_catalog_json` 指定本地 JSON 文件，包含从代理获取的模型 |
| 拦截 ab.chatgpt.com | goproxy MITM 代理，拦截 Statsig 的 `/v1/initialize` 响应，删除模型限制字段 |
| PAC 代理 | 仅 `ab.chatgpt.com` 走本地代理，其他域名直连，不影响 VPN |
| 对话历史保留 | 不改 `model_provider`，用 `openai_base_url` 重定向 API 请求 |
| 配置合并 | 读取原 `config.toml`，插入我们的配置行，保留原设置 |

## 使用

### 快速开始

```bash
make run
```

浏览器自动打开 `http://127.0.0.1:18900`，输入 API Key 登录。

- **首次登录**：弹密码框安装 CA 证书（一次），之后不再弹
- **登录后**：配置写入 + PAC 设置 + 代理启动，全自动
- **退出**：点击"退出"恢复原始配置，清 PAC
- **关闭**：点击"关闭应用"退出进程

### 命令行

```bash
# 构建 + 启动
make run

# 构建全部产物（DMG + EXE）
make build

# 仅 macOS 二进制
make mac

# 仅 Windows 可执行文件
make win

# 仅 DMG 安装包
make dmg

# 清理构建产物
make clean
```

### 命令

```
yuanshu-ai           → 打开浏览器 GUI（默认）
yuanshu-ai login     → 命令行登录（会启动代理并等待）
yuanshu-ai logout    → 退出
yuanshu-ai server    → 仅启动代理服务
```

## 构建产物

```
build/
├── yuanshu-ai           macOS 二进制（7MB）
├── yuanshu-ai.exe       Windows 可执行文件（8MB，无控制台窗口）
├── 元数AI.app            macOS 应用包
└── 元数AI.dmg            macOS 安装包（3.5MB）
```

双击 `元数AI.dmg` 安装后，从 Applications 打开。

## 项目结构

```
codex-proxy/
├── main.go                   入口
├── app/
│   ├── app.go                App 结构体
│   ├── cert.go               CA 证书生成（通用）
│   ├── cert_darwin.go        CA 安装/卸载（macOS）
│   ├── cert_windows.go       CA 安装/卸载（Windows）
│   ├── config.go             Codex 配置管理 + 模型目录
│   ├── gui.go                Web GUI 服务（通用）
│   ├── login.go              登录/退出流程
│   ├── machine_darwin.go     机器 UUID 获取（macOS）
│   ├── machine_windows.go    机器 UUID 获取（Windows）
│   ├── proxyconfig_darwin.go PAC 设置（macOS networksetup）
│   ├── proxyconfig_windows.go PAC 设置（Windows 注册表）
│   ├── browser_darwin.go     打开浏览器（macOS）
│   ├── browser_windows.go    打开浏览器（Windows）
│   ├── process_unix.go       进程检测（Unix）
│   ├── process_windows.go    进程检测（Windows）
│   ├── server.go             goproxy MITM 代理
│   └── web/index.html        前端页面
├── build/                    构建产物
├── Makefile
├── go.mod / go.sum
├── logo-final.png / yuanshu-ai.icns / yuanshu-ai.ico
└── yuanshu-ai.command        macOS 双击启动脚本
```

## 常见问题

**Q: 登录后模型没出现？**
完全退出 ChatGPT（Dock 右键 → 退出）再重新打开，让 PAC 配置生效。

**Q: 对话历史不见了？**
改用了 `openai_base_url` 代替 `model_provider` 切换，对话历史应该保留。如果仍不显示，退出登录恢复原配置即可找回。

**Q: 每次登录都弹密码框？**
仅首次需要安装 CA 证书。之后不再弹。如果频繁弹，可能是证书文件损坏（`~/.codex/yuanshu/ca.key`），清掉重新登录。

**Q: 如何换 API Key？**
浏览器点"退出"回到登录页，重新输入 Key。

**Q: 端口被占用？**
GUI 默认 `:18900`，代理默认 `:9090`。修改 `app/gui.go` 和 `app/server.go` 中的端口号。

**Q: macOS 提示"无法验证开发者"？**
右键 → 打开 → 仍要打开。

**Q: Windows 报安全警告？**
双击 `yuanshu-ai.exe` 时点"仍要运行"即可。

**Q: 双击提示"未能打开文稿"？**
用 `make dmg` 生成 `.app`，或终端运行 `./build/yuanshu-ai`。
