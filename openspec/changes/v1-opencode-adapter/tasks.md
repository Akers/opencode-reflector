## 1. 项目骨架与 Go 模块初始化

- [x] 1.1 初始化 Go 模块（`go mod init github.com/user/opencode-reflector`），创建 `cmd/reflector/main.go` 入口
- [x] 1.2 添加核心依赖：`go-chi/chi`（HTTP 路由）、`modernc.org/sqlite`（SQLite 驱动）、`gopkg.in/yaml.v3`（YAML 解析）、OpenAI Go SDK
- [x] 1.3 创建 Adapter 共享类型包 `adapters/shared/`，定义 `SessionAdapter`、`CanonicalSession`、`CanonicalMessage`、`CanonicalToolCall`、`CapabilityMatrix` TypeScript 接口
- [x] 1.4 创建 `.reflector/` 目录结构模板（data/、reports/、logs/、prompts/、hooks/），编写 `.gitignore` 规则

## 2. 数据模型与配置管理

- [x] 2.1 实现 Go 侧 CanonicalSession 数据模型（`internal/model/session.go`），JSON tag 与 TypeScript 定义对齐
- [x] 2.2 实现 Go 侧 Metrics 数据模型（`internal/model/metrics.go`），覆盖 M-001~M-025 全部 25 项指标
- [x] 2.3 实现配置管理模块（`internal/config/config.go`）：加载 `reflector.yaml`、支持热更新、内置默认值
- [x] 2.4 实现 Watermark 持久化（SQLite watermarks 表，在 store 包中实现）

## 3. SQLite 数据持久化

- [x] 3.1 实现 SQLite 存储层（`internal/store/sqlite.go`）：数据库初始化、建表 DDL 执行、连接池管理
- [x] 3.2 实现 sessions 表 CRUD：插入会话记录（含全部指标字段）、按日期范围查询、按状态查询
- [x] 3.3 实现 agent_participations 表 CRUD：批量插入智能体参与记录
- [x] 3.4 实现 tool_calls 表 CRUD：批量插入工具/MCP/Skill 调用记录
- [x] 3.5 实现 reflection_logs 表 CRUD：记录每次反思执行的日志
- [x] 3.6 实现 watermarks 表 CRUD：读取/更新上次反思完成时间
- [x] 3.7 实现数据清理功能：按天数删除过期记录及对应日报文件
- [x] 3.8 实现写入失败降级：重试 3 次后写入 fallback JSON 文件

## 4. 核心引擎模块

- [x] 4.1 实现 Session 解析器（`internal/engine/parser.go`）：接收 CanonicalSession JSON，解析为内部模型
- [x] 4.2 实现指标提取器（`internal/engine/extractor.go`）：从解析后的 Session 计算 M-001~M-022 指标（不含情感指标）
- [x] 4.3 实现任务状态判定器（`internal/engine/classifier.go`）：L1（OpenSpec 指令检测）、L2（LLM 意图分类）、L3（中英文关键词匹配）三级策略
- [x] 4.4 实现情感分析器（`internal/engine/sentiment.go`）：调用 OpenAI 兼容 API、结构化 JSON 输出、脱敏处理、成本控制
- [x] 4.5 实现触发管理器（`internal/engine/trigger.go`）：防抖（5分钟）、并发排队、手动触发优先级
- [x] 4.6 实现 Hook 执行器（`internal/engine/hook.go`）：扫描 hooks 目录、stdin/stdout JSON 通信、exit code 处理、失败不阻断

## 5. 提示词管理

- [x] 5.1 创建默认提示词文件（`internal/engine/prompts/sentiment.md`、`internal/engine/prompts/classify.md`）
- [x] 5.2 实现提示词加载器：优先从 `.reflector/prompts/` 读取，fallback 到编译时嵌入的默认值
- [x] 5.3 使用 Go `embed` 包将默认提示词编译时嵌入二进制

## 6. 日报生成

- [x] 6.1 实现日报生成器（`internal/engine/report.go`）：按模板生成 Markdown 日报
- [x] 6.2 实现日期归属逻辑：定时触发用前一天日期、事件/手动触发用当前日期
- [x] 6.3 实现日报幂等追加：已有日报文件追加新数据，不覆盖

## 7. HTTP API 服务

- [x] 7.1 实现 HTTP 服务框架（`internal/server/server.go`）：chi 路由
- [x] 7.2 实现 `POST /api/v1/reflect` 端点：接收请求、编排完整反思流程、返回标准化响应
- [x] 7.3 实现 `GET /api/v1/health` 端点：健康检查
- [x] 7.4 实现 `GET /api/v1/stats` 端点：反思工具运行统计
- [x] 7.5 实现 `POST /api/v1/cleanup` 端点：数据清理（配合 `/or:cleanup` 指令）
- [x] 7.6 编排 `main.go` 启动流程：加载配置 → 初始化数据库 → 注册路由 → 启动 HTTP 服务

## 8. opencode Adapter 实现

- [x] 8.1 初始化 Adapter 项目（`adapters/opencode/`）：`package.json`、`tsconfig.json`、安装 `@opencode-ai/plugin` 和 `@opencode-ai/sdk`
- [x] 8.2 实现插件入口（`adapters/opencode/src/index.ts`）：注册指令（`/or:reflect_now`、`/or:cleanup`）、初始化触发监听
- [x] 8.3 实现 Core Engine 进程管理（`adapters/opencode/src/engine-manager.ts`）：按需启动、健康检查、优雅关闭、重试
- [x] 8.4 实现 Session 获取与转换（`adapters/opencode/src/session-adapter.ts`）：调用 SDK 获取会话、转换为 CanonicalSession 格式
- [x] 8.5 实现定时触发（`adapters/opencode/src/index.ts`）：读取配置中的触发时间、setInterval 定时触发
- [x] 8.6 实现能力矩阵声明：根据 SDK 实际能力填写 CapabilityMatrix

## 9. 构建与安装脚本

- [x] 9.1 编写 `scripts/build.sh`：编译 Go 二进制 + 构建 Adapter TypeScript
- [x] 9.2 编写 `scripts/install.sh`：安装到 opencode 插件目录 + 初始化 `.reflector/` 目录结构
- [x] 9.3 编写默认配置文件模板（`reflector.yaml`）

## 10. 集成验证

- [x] 10.1 编写 Go 单元测试：指标提取、任务判定（L1/L2/L3）、情感分析 mock、日报生成
- [x] 10.2 编写 Go 集成测试：完整反思流程（HTTP API → 指标 → SQLite → 日报）
- [x] 10.3 端到端手动验证：构建脚本验证（build.sh 通过）
