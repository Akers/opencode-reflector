# opencode-reflector（反思怪）需求与设计方案

> **文档版本**：v1.0  
> **最后更新**：2026-04-19  
> **状态**：草案（Draft）

---

## 一、项目概述

### 1.1 项目定位

**opencode-reflector**（中文名：反思怪）是一个为 AI 编程智能体工具打造的反思型监控插件。它通过从会话（Session）中自动提取、分析任务执行数据，为智能体框架选型评估与自我进化提供数据支撑。

### 1.2 目标用户

- 使用 AI 编程智能体进行日常开发的工程师
- 需要评估智能体工具效率的团队管理者
- 进行智能体框架选型决策的技术负责人

### 1.3 核心价值

| 价值维度 | 描述 |
|---------|------|
| **数据透明** | 自动采集 Token 消耗、工具调用、任务完成率等关键指标，消除"黑盒" |
| **人机协作量化** | 量化人类参与度与情绪趋势，揭示真实的人机协作效率 |
| **框架选型支撑** | 跨工具横向对比，为 opencode/openclaw/claudecode 选型提供数据依据 |
| **持续改进** | 通过历史趋势分析，驱动智能体使用策略的持续优化 |

### 1.4 兼容的智能体工具

| 工具 | 插件开发文档 | 插件技术栈 |
|------|------------|-----------|
| opencode | https://opencode.ai/docs/plugins/ | Node.js (`@opencode-ai/plugin`) |
| openclaw | https://docs.openclaw.ai/plugins/building-plugins | 待调研 |
| claudecode | https://code.claude.com/docs/en/plugins | 待调研 |

---

## 二、术语与概念

| 术语 | 定义 |
|------|------|
| **Session（会话）** | 一次从用户发起任务到任务结束（或中断）的完整对话序列，包含多轮 human/agent 消息 |
| **Task（任务）** | 与一个 Session 对应的工作单元，有明确的开始和结束状态 |
| **Agent（智能体）** | AI 编程助手（包括主代理和子代理），在会话中响应用户请求 |
| **Tool（工具）** | 智能体框架提供的内置能力（如 read、write、bash 等） |
| **MCP（Model Context Protocol）** | 外部服务调用协议，用于扩展智能体的能力边界 |
| **Skill（技能）** | 封装特定工作流的高级能力模板（如 pptx、pdf 等） |
| **Reflect（反思）** | 本工具的核心动作——对会话数据进行指标提取、状态识别和情感分析 |
| **Daily Report（日报）** | 按自然日汇总的反思报告，以 Markdown 格式存储 |
| **CanonicalSession** | 统一的内部会话数据模型，屏蔽不同智能体工具的数据格式差异 |

---

## 📂 文档索引

本文档已拆分为以下独立章节，便于按需查阅：

| 章节 | 文件 | 内容概要 |
|------|------|---------|
| **三、功能需求** | [functional-requirements.md](rd/functional-requirements.md) | 触发机制（FR-001~004）、会话数据获取（FR-005~008）、25 项监控指标（FR-009~010）、数据持久化 SQLite（FR-011~012）、日报生成（FR-013~014）、配置管理（FR-015~016）、Hook 生命周期（FR-017） |
| **四、非功能需求** | [non-functional-requirements.md](rd/non-functional-requirements.md) | 性能指标、可靠性（降级策略）、隐私安全（脱敏+本地优先）、可观测性 |
| **五、架构设计** | [architecture-design.md](rd/architecture-design.md) | Core Engine(Go) + Adapter(Node.js) 整体架构、HTTP localhost 通信、目录结构、技术选型 |
| **六、接口设计** | [api-design.md](rd/api-design.md) | Core Engine HTTP API（POST /reflect）、Adapter 公共接口（SessionAdapter、CanonicalSession、CapabilityMatrix） |
| **七、开发阶段规划** | [development-roadmap.md](rd/development-roadmap.md) | 4 阶段路线图：MVP → 完整功能 → 多工具适配 → 分析洞察（V2） |
| **八、开放问题** | [open-questions.md](rd/open-questions.md) | 6 个待调研确认问题（Token 可获取性、插件机制、性能基准、情感分析准确度等） |

### 推荐阅读顺序

1. **初次阅读**：按章节顺序通读，建立整体认知
2. **开发参考**：根据当前开发阶段查阅对应章节（Phase 1 → 功能需求 + 架构设计 + 接口设计）
3. **技术调研**：优先查阅 [开放问题](rd/open-questions.md)，确认技术可行性
4. **指标实现**：直接查阅 [功能需求 §3.3 监控指标提取](rd/functional-requirements.md#33-监控指标提取)
