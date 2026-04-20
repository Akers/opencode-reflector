# 五、架构设计

> 本文档是 [需求与设计方案主文档](../requirements-and-design.md) 的组成部分。

---

### 5.1 整体架构

采用 **Core Engine + Adapter** 架构，核心逻辑只写一次，通过薄 Adapter 层适配不同智能体工具。

```
┌─────────────────────────────────────────────────────────────┐
│                      智能体工具层                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   opencode   │  │   openclaw   │  │  claudecode  │      │
│  │  (Node.js)   │  │  (Node.js)   │  │  (Node.js)   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │              │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐      │
│  │   Opencode   │  │   Openclaw   │  │  ClaudeCode  │      │
│  │   Adapter    │  │   Adapter    │  │   Adapter    │      │
│  │  (Node.js)   │  │  (Node.js)   │  │  (Node.js)   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │              │
│         └──────────────────┼──────────────────┘              │
│                            │ CanonicalSession                │
├────────────────────────────┼────────────────────────────────┤
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Core Engine (Go 独立进程)                │    │
│  │                                                     │    │
│  │  ┌─────────────┐  ┌──────────────┐  ┌───────────┐  │    │
│  │  │  Trigger    │  │  Session     │  │  Metric   │  │    │
│  │  │  Manager    │  │  Parser      │  │  Extractor│  │    │
│  │  └─────────────┘  └──────────────┘  └───────────┘  │    │
│  │                                                     │    │
│  │  ┌─────────────┐  ┌──────────────┐  ┌───────────┐  │    │
│  │  │  Task       │  │  Sentiment   │  │  Report   │  │    │
│  │  │  Classifier │  │  Analyzer    │  │  Generator│  │    │
│  │  └─────────────┘  └──────────────┘  └───────────┘  │    │
│  │                                                     │    │
│  │  ┌─────────────┐  ┌──────────────┐                  │    │
│  │  │  Data Store │  │  Hook Runner │                  │    │
│  │  │  (SQLite)   │  │  (脚本执行器) │                  │    │
│  │  └─────────────┘  └──────────────┘                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  通信方式：Adapter ──HTTP localhost──▶ Core Engine            │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 通信架构

```
Adapter (Node.js)                    Core Engine (Go)
     │                                     │
     │  POST /api/v1/reflect               │
     │  { sessions: [...], config: {...} } │
     │ ──────────────────────────────────▶ │
     │                                     │
     │         HTTP 200                    │
     │  { reportPath: "...", status: "OK"} │
     │ ◀────────────────────────────────── │
```

- **Adapter 职责**（薄层）：
  1. 监听触发事件（定时/事件/手动）
  2. 获取原始 Session 数据
  3. 转换为 `CanonicalSession` 格式
  4. 调用 Core Engine 的 HTTP API

- **Core Engine 职责**（厚重）：
  1. 接收标准化 Session 数据
  2. 指标提取与计算
  3. 任务状态判定（L1/L2/L3）
  4. 情感分析
  5. 数据持久化（SQLite）
  6. 日报生成
  7. Hook 执行

### 5.3 目录结构

```
opencode-reflector/
├── cmd/                          # Go 入口
│   └── reflector/
│       └── main.go               # Core Engine 主程序
├── internal/                     # Go 内部包
│   ├── api/                      # HTTP API 服务
│   ├── engine/                   # 核心引擎
│   │   ├── trigger.go            # 触发管理
│   │   ├── parser.go             # Session 解析
│   │   ├── extractor.go          # 指标提取
│   │   ├── classifier.go         # 任务状态判定
│   │   ├── sentiment.go          # 情感分析
│   │   ├── report.go             # 日报生成
│   │   └── hook.go               # Hook 执行器
│   ├── model/                    # 数据模型
│   │   ├── session.go            # CanonicalSession 定义
│   │   └── metrics.go            # 指标数据结构
│   ├── store/                    # 数据持久化
│   │   └── sqlite.go             # SQLite 操作
│   └── config/                   # 配置管理
│       └── config.go             # 配置加载与解析
├── adapters/                     # 各工具的 Adapter
│   ├── opencode/                 # opencode Adapter
│   │   ├── package.json
│   │   ├── src/
│   │   │   ├── index.ts          # 插件入口
│   │   │   ├── adapter.ts        # SessionAdapter 实现
│   │   │   └── triggers.ts       # 触发事件监听
│   │   └── tsconfig.json
│   ├── openclaw/                 # openclaw Adapter
│   └── claudecode/               # claudecode Adapter
├── prompts/                      # 默认提示词
│   ├── task-status-classification.md
│   ├── sentiment-analysis.md
│   └── report-summary.md
├── scripts/                      # 构建与部署脚本
│   ├── build.sh                  # 编译 Go + 构建 Adapter
│   └── install.sh                # 安装到各智能体工具
├── README.md                     # 项目说明
└── docs/                         # 文档
    ├── requirements-and-design.md  # 需求与设计方案主文档
    └── rd/                          # 拆分文档目录
        ├── functional-requirements.md
        ├── non-functional-requirements.md
        ├── architecture-design.md
        ├── api-design.md
        ├── development-roadmap.md
        └── open-questions.md
```

### 5.4 技术选型

| 组件 | 技术选型 | 选型理由 |
|------|---------|---------|
| Core Engine | Go | 高性能文件读写、SQLite 天然亲和、跨平台编译为单一二进制 |
| Adapter 层 | Node.js/TypeScript | 各智能体工具插件机制均为 Node.js 生态 |
| 数据存储 | SQLite | 单文件、零部署、支持结构化查询、性能充足 |
| 通信协议 | HTTP localhost | 简单可靠、跨语言、支持流式响应 |
| 配置文件 | YAML | 可读性好、支持注释、开发者友好 |
| 日报格式 | Markdown | 可读性好、可直接在 IDE 中预览 |
| 提示词格式 | Markdown | 可读性好、支持热修改 |
