# 元数AI

一键配置 Codex / ChatGPT Desktop 使用自定义 AI 模型（DeepSeek、Qwen 等）。

## 工作原理

```
用户双击 → 浏览器界面打开
    ↓
输入 API Key → 连接代理服务器获取模型列表
    ↓
1. 写入 Codex 配置（config.toml + 模型列表）
2. 修改 hosts（ab.chatgpt.com → 127.0.0.1）
3. 启动本地 Statsig 加速服务
    ↓
打开 ChatGPT → 自定义模型可用 ✅
```

### 关键原理

| 问题 | 答案 |
|---|---|
| 自定义模型怎么来的？ | Codex 配置中的 `model_catalog_json` 指定本地 JSON 文件 |
| 为什么改 hosts？ | 拦截 `ab.chatgpt.com` 请求，避免 Statsig 隐藏自定义模型 |
| 中文界面怎么来的？ | ChatGPT 系统语言设置（`--lang=zh-CN`），跟网络无关 |
| 为什么不需要证书？ | 模型列表来自本地配置文件，不需要 HTTPS |

## 控制模型列表

模型列表来自你的 MetaProxy 服务器。元数AI 调用 `/v1/models` 获取所有模型。

### 修改显示的模型

编辑 `config.go` 中的 `fetchModels` 函数：

```go
// 过滤掉不想显示的模型
for _, m := range result.Data {
    if m.ID != "wan2.7-image" {  // 排除这个模型
        models = append(models, m.ID)
    }
}
```

### 模型字段说明

每个模型包含 32 个字段，写入 `~/.codex/yuanshu/metaproxy-models.json`。

关键字段：

| 字段 | 说明 | 默认值 |
|---|---|---|
| `slug` | 模型 ID（唯一标识） | 来自 API |
| `display_name` | 界面显示的名称 | 自动从 slug 转换 |
| `description` | 模型描述 | "xxx via proxy" |
| `supported_reasoning_levels` | 可选的推理等级 | 按类型（GPT=6级，DS=5级） |
| `context_window` | 上下文窗口大小 | 1000000（Qwen=131072） |
| `truncation_policy` | 截断策略 | mode:tokens, limit:10000 |

其余字段（`shell_type`、`visibility`、`supported_in_api` 等）使用固定模板。

### 修改推理等级

编辑 `config.go` 中的 `levels` 函数：

```go
func levels(slug string) []Level {
    s := strings.ToLower(slug)
    if strings.HasPrefix(s, "gpt") {
        return []Level{
            {"low", "Fast"},
            {"medium", "Balanced"},
            {"high", "Deep"},
            {"xhigh", "Extra deep"},
            {"max", "Max"},
            {"ultra", "Ultra"},
        }
    }
    if strings.HasPrefix(s, "deepseek") {
        return []Level{
            {"low", "Fast"},
            {"medium", "Balanced"},
            {"high", "Default"},
            {"xhigh", "Extra deep"},
            {"max", "Max"},
        }
    }
    return []Level{{"low", "Fast"}, {"medium", "Balanced"}, {"high", "Deep"}}
}
```

## 使用

```
yuanshu-ai           → 打开浏览器 GUI（默认）
yuanshu-ai login     → 命令行登录
yuanshu-ai logout    → 命令行退出
yuanshu-ai server    → 启动加速服务
```

### 浏览器界面

默认打开 `http://127.0.0.1:18900`。

- **登录**：输入 API Key → 自动写配置 + 改 hosts + 启动加速服务
- **退出**：恢复配置 + 删 hosts + 停止加速服务
- **状态**：实时显示登录、加速服务、模型数量
- **模型列表**：登录后以网格显示所有可用模型

### 命令行

```bash
# 登录
./build/yuanshu-ai login
请输入你的 API Key: ****
→ 正在连接服务器...
✅ 连接成功！共 16 个模型
→ 启动 Statsig 加速服务...
✅ Statsig 加速服务已启动
🎉 登录成功

# 退出
./build/yuanshu-ai logout
🎉 已退出
```

## 构建

```bash
cd client

make        # 全部构建
make mac    # 仅 macOS 二进制
make win    # 仅 Windows 二进制
make dmg    # 仅 DMG 安装包
make clean  # 清理
```

产物在 `client/build/`：

```
build/
├── yuanshu-ai           # macOS 二进制（6.5MB）
├── yuanshu-ai.exe       # Windows 二进制（7.1MB）
├── 元数AI.app            # macOS 应用包
└── 元数AI.dmg            # macOS 安装包（3.3MB）
```

## 目录结构

```
client/
├── main.go              # 入口，命令分发
├── config.go            # 登录/退出 + Codex 配置生成
├── server.go            # Statsig 加速服务（本地劫持）
├── cert.go              # 证书安装/删除（跨平台）
├── hosts.go             # hosts 管理（跨平台）
├── gui.go               # Web GUI 服务
├── web/index.html       # 浏览器界面
├── Makefile             # 构建
├── logo-final.png       # 图标源图
├── scripts/             # 资源生成脚本
├── yuanshu-ai.icns      # macOS 图标
├── yuanshu-ai.ico       # Windows 图标
└── build/               # 构建产物
```

## 常见问题

**Q: 双击提示"未能打开文稿"？**
用 `make dmg` 生成 `.app`，或者终端运行 `./yuanshu-ai`。

**Q: 提示"无法验证开发者"？**
右键 → 打开 → 仍要打开。首次运行后不再提示。

**Q: 登录后模型没出现？**
完全退出 ChatGPT（Dock 右键 → 退出）再重新打开。

**Q: 如何换 API Key？**
先 `logout` 再 `login` 重新输入。

**Q: Windows 报安全警告？**
双击运行时会弹 Windows Defender 提示，点"仍要运行"即可。

**Q: 端口被占用？**
默认端口 3000（加速服务）和 18900（GUI），修改 `config.go` 和 `server.go` 中的端口号。
