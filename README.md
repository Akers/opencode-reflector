<a id="readme-top"></a>

<br />
<div align="center">
  <h3 align="center">opencode-reflector（反思怪）</h3>
  <p align="center">
    AI 编程智能体的反思监控插件 —— 让你看见 AI 的工作，量化人机协作效率。
    <br />
    <a href="docs/"><strong>查看完整文档</strong></a>
  </p>
</div>

---

## Table of Contents

- [About The Project](#about-the-project)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Configuration Reference](#configuration-reference)
- [Command Reference](#command-reference)
- [Architecture](#architecture)
- [Roadmap](#roadmap)
- [Documentation](#documentation)
- [License](#license)

---

## About The Project

opencode-reflector 是一个为 AI 编程智能体（opencode / openclaw / claudecode）设计的反思监控插件。它自动追踪每次会话中的关键指标，生成结构化日报，帮助开发者：

- **透明化 AI 工作过程** — 了解每次会话消耗了多少 Token、调用了哪些工具、花了多长时间
- **量化人机协作效率** — 追踪人类参与度、干预次数、情绪反馈
- **为框架选型提供数据支撑** — 对比不同 AI 编程工具的实际表现

### 核心特性

- 📊 **25+ 监控指标** — 时间、Token、工具调用、Agent 参与、人类交互、情感分析
- 📝 **自动日报生成** — Markdown 格式的每日工作报告
- 🧠 **智能情感分析** — 基于宿主工具的 SDK 调用 LLM，自动降级为规则引擎
- 🎯 **三级任务判定** — L1(指令检测) → L2(LLM分类) → L3(关键词匹配)
- 🔌 **Adapter 架构** — Core Engine (Go) + Adapter (Node.js)，轻松适配新工具
- ⚡ **多种触发方式** — 定时 / 事件（N轮对话）/ 手动（`/or:reflect_now`）
- 🔒 **本地优先** — 所有数据存储在本地 SQLite，敏感信息自动脱敏
- 🪝 **Hook 扩展** — 10 个生命周期节点支持自定义脚本

### Built With

| 组件 | 技术 |
|------|------|
| Core Engine | Go 1.25, chi/v5, modernc.org/sqlite, yaml.v3 |
| opencode Adapter | TypeScript, @opencode-ai/plugin, @opencode-ai/sdk |
| 数据存储 | SQLite (纯 Go 驱动，无 CGO) |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 22+ / Bun
- opencode v1.14+

### Installation

#### 1. 构建 Core Engine

```sh
# 克隆仓库
git clone https://github.com/user/opencode-reflector.git
cd opencode-reflector

# 编译 Go 二进制
./scripts/build.sh
```

#### 2. 安装 opencode Adapter

```sh
./scripts/install.sh
```

#### 3. 配置

编辑项目根目录下的 `.reflector/reflector.yaml`：

```yaml
# 情感分析模式: agent=SDK调用LLM | builtin=规则引擎 | off=关闭
sentiment:
  enabled: true
  mode: "agent"

# 定时触发
trigger:
  time:
    enabled: true
    schedule: "00:00"
```

<p align="right">(<a href="#readme-top">back to">back to top</a>)</p>

---

## Usage

### 命令触发

在 opencode TUI 中使用命令：

| 命令 | 说明 |
|------|------|
| `/or:reflect_now` | 立即执行一次反思分析 |
| `/or:cleanup` | 清理旧数据（默认 90 天） |

### 自动触发

- **定时触发** — 每日 00:00 自动执行（可在 `reflector.yaml` 中配置）
- **事件触发** — 每 10 轮对话自动触发一次分析

### 生成的日报

日报存储在 `.reflector/reports/` 下，包含：

```
📁 .reflector/
├── reflector.db          # SQLite 指标数据库
├── reflector.yaml        # 配置文件
├── reports/
│   └── dayreport-2026-04-20.md
├── prompts/              # 可自定义的提示词
│   ├── sentiment.md
│   └── classify.md
├── hooks/                # 自定义 Hook 脚本
└── bin/
    └── reflector         # Go 二进制
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Configuration Reference

配置文件位于 `.reflector/reflector.yaml`，所有字段均有默认值，缺失字段自动回退。

### 完整配置示例

```yaml
# ── 触发方式 ──────────────────────────────────────
trigger:
  time:
    enabled: true           # 是否启用定时触发
    schedule: "00:00"       # 24小时制，如 "08:30"、"23:00"
  events:
    enabled: true           # 是否启用事件触发
    types:                  # 监听的事件类型
      - TASK_FINISHED       #   检测到任务完成时触发
      - N_MESSAGES          #   每N轮对话触发
    messageInterval: 10     # N_MESSAGES 的 N 值

# ── Core Engine ───────────────────────────────────
port: 19870                  # Core Engine HTTP 监听端口

# ── 模型配置 ──────────────────────────────────────
model:
  override: ""               # 强制覆盖模型，格式 "provider/model"
                             # 留空则自动获取宿主工具默认模型
                             # 示例: "anthropic/claude-sonnet-4"

# ── 情感分析 ──────────────────────────────────────
sentiment:
  enabled: true              # 是否启用情感分析
  skipOnMessageTrigger: true # 事件触发时跳过（节省 Token）
  mode: "agent"              # agent  - 通过宿主 SDK 调用 LLM（推荐）
                             # builtin - 基于关键词的规则引擎（离线）
                             # off     - 完全关闭

# ── 任务分类 (L2) ─────────────────────────────────
classification:
  enabled: true              # 是否启用 LLM 意图分类
  confidenceThreshold: 0.7   # 置信度阈值 (0.0~1.0)

# ── 数据管理 ──────────────────────────────────────
retention:
  days: 90                   # 数据保留天数，超过自动清理
report:
  template: "default"        # 日报模板（目前仅 default）
logLevel: "info"             # debug | info | warn | error
```

### 情感分析模式说明

| 模式 | 说明 | Token 消耗 | 准确度 |
|------|------|-----------|--------|
| `agent` | 创建临时影子 Session，通过 SDK 调用宿主 LLM | 有（使用宿主默认模型） | 高 |
| `builtin` | 中英文关键词统计（正面/负面计数→比率/分数） | 无 | 中 |
| `off` | 不执行情感分析，所有指标标记为 N/A (-1) | 无 | — |

`agent` 模式降级链：SDK Prompt → 规则引擎 → N/A

### Hook 扩展

在 `.reflector/hooks/` 目录下放置可执行脚本，文件名格式为 `{hookPoint}.{ext}`：

```
.reflector/hooks/
├── before-reflect.sh
├── after-save-metrics.py
└── on-error.go           # 需预编译为二进制
```

**可用 Hook 节点**：

| Hook Point | 触发时机 | 输入 |
|------------|---------|------|
| `before-reflect` | 反思分析开始前 | trigger 信息 |
| `after-reflect` | 反思分析完成后 | 分析结果 |
| `before-save-metrics` | 指标保存前 | SessionMetrics JSON |
| `after-save-metrics` | 指标保存后 | SessionMetrics JSON |
| `before-sentiment` | 情感分析前 | 待分析消息 |
| `after-sentiment` | 情感分析后 | SentimentResult JSON |
| `before-report` | 日报生成前 | 指标汇总 |
| `after-report` | 日报生成后 | 报告路径 |
| `after-classify` | 任务状态判定后 | 分类结果 |
| `on-error` | 发生错误时 | 错误信息 |

Hook 脚本通过 **stdin** 接收 JSON 数据，通过 **stdout** 返回结果。脚本退出码非 0 时记录警告但不阻断主流程。

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Command Reference

### TUI 命令（在 opencode 对话中输入）

| 命令 | 别名 | 说明 |
|------|------|------|
| `/or:reflect_now` | `/or:reflect`, `/or:rf` | 立即执行反思分析，分析所有未处理的会话 |
| `/or:cleanup` | `/or:clean` | 清理过期数据，默认保留 90 天 |

### HTTP API（Core Engine）

Core Engine 启动后监听 `http://127.0.0.1:{port}`，提供以下端点：

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/reflect` | 执行反思分析 |
| `GET` | `/api/v1/health` | 健康检查 |
| `GET` | `/api/v1/stats` | 运行统计 |
| `POST` | `/api/v1/cleanup` | 清理旧数据 |

#### POST /api/v1/reflect

```json
{
  "trigger_type": "MANUAL",
  "trigger_detail": "TOOL_TRIGGER",
  "sessions": [...],
  "config": {
    "sentiment_enabled": true,
    "sentiment_source": "agent"
  }
}
```

#### POST /api/v1/cleanup

```json
{ "days": 90 }
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Architecture

```
┌──────────────────────────────────────────┐
│            opencode Runtime              │
│  ┌────────────────────────────────────┐  │
│  │        Reflector Plugin            │  │
│  │  ┌─────────────┐ ┌──────────────┐ │  │
│  │  │ Sentiment   │ │ Task         │ │  │
│  │  │ Analyzer    │ │ Classifier   │ │  │
│  │  │ (SDK+降级)  │ │ (SDK)        │ │  │
│  │  └─────────────┘ └──────────────┘ │  │
│  └────────────┬───────────────────────┘  │
└───────────────┼──────────────────────────┘
                │ HTTP POST /api/v1/reflect
                │ (session.metadata 含分析结果)
                ▼
┌──────────────────────────────────────────┐
│         Core Engine (Go)                 │
│  ┌──────────┐ ┌────────┐ ┌───────────┐  │
│  │ 指标提取  │ │ 日报   │ │ SQLite    │  │
│  │ 25项指标  │ │ 生成   │ │ 持久化    │  │
│  └──────────┘ └────────┘ └───────────┘  │
└──────────────────────────────────────────┘
```

**设计原则**：让宿主工具做它最擅长的事（调 LLM），让 Core Engine 做它最擅长的事（存储、指标计算、报告生成）。

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Roadmap

- [x] Core Engine (Go) — 指标提取、SQLite 持久化、日报生成
- [x] opencode Adapter — 插件注册、SDK 集成、情感分析
- [x] 情感分析 — SDK Prompt 模式 + 规则引擎降级
- [x] L2 任务分类 — Adapter 侧 LLM 调用
- [ ] openclaw Adapter
- [ ] claudecode Adapter
- [ ] 周报 / 月报 / 趋势分析
- [ ] Web Dashboard

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Documentation

| 文档 | 说明 |
|------|------|
| [docs/original-requirements.md](docs/original-requirements.md) | 原始需求文档 |
| [docs/requirements-and-design.md](docs/requirements-and-design.md) | 需求与设计总览 |
| [docs/rd/](docs/rd/) | 详细设计文档（功能需求、架构、API 等） |
| [docs/rd/sentiment-redesign.md](docs/rd/sentiment-redesign.md) | 情感分析重设计文档 |
| [openspec/](openspec/) | OpenSpec 变更管理 |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## License

MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>
