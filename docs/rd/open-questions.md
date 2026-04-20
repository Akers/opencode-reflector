# 八、开放问题（Open Questions）

> 本文档是 [需求与设计方案主文档](../requirements-and-design.md) 的组成部分。

---

以下问题需要在开发前进一步调研确认：

| # | 问题 | 影响 | 建议 |
|---|------|------|------|
| 1 | opencode 插件 API 是否暴露每条消息的 Token 计数？ | 决定 M-005~M-008 指标是否可获取 | 需 PoC 验证 |
| 2 | openclaw/claudecode 的插件机制和 Session 存储格式？ | Phase 3 进度 | 需调研文档 |
| 3 | Go 与 Node.js 之间的 HTTP 通信性能是否可接受？ | 整体架构 | 需基准测试 |
| 4 | 情感分析的中英文混合场景准确度如何？ | 情感分析模块可信度 | 需样本测试 |
| 5 | 是否需要支持多项目（多 `.reflector` 目录）？ | Core Engine 实例管理 | 建议暂不支持，V2 考虑 |
| 6 | Core Engine 的进程生命周期管理（随智能体工具启动/退出）？ | 运维体验 | 建议由 Adapter 管理，按需启动 |
