# 七、开发阶段规划

> 本文档是 [需求与设计方案主文档](../requirements-and-design.md) 的组成部分。

---

### Phase 1：基础框架（MVP）

**目标**：跑通核心流程，验证架构可行性

- [ ] Core Engine 基础框架（Go HTTP 服务）
- [ ] SQLite 数据模型与 CRUD
- [ ] opencode Adapter（MVP，仅支持定时触发）
- [ ] 基础指标提取（时间、消息数、Token）
- [ ] 简易日报生成（无情感分析）

### Phase 2：完整功能

**目标**：补全所有核心功能

- [ ] 全部触发机制（事件触发、手动触发）
- [ ] 任务状态多级判定（L1/L2/L3）
- [ ] 情感分析模块
- [ ] 工具/MCP/Skill 调用统计
- [ ] 完整日报模板
- [ ] Hook 机制
- [ ] 配置文件支持

### Phase 3：多工具适配

**目标**：适配 openclaw 和 claudecode

- [ ] openclaw Adapter
- [ ] claudecode Adapter
- [ ] 能力矩阵完善
- [ ] 跨工具对比报告

### Phase 4：分析与洞察（V2 预留）

**目标**：提供更深层次的数据洞察

- [ ] 周报/月报
- [ ] 趋势分析（Token 消耗趋势、人类参与度变化）
- [ ] 智能体工具对比分析
- [ ] 自定义报告模板
