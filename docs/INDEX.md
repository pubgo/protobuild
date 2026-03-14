# 文档总览

本文档用于串联项目全部核心文档，确保阅读顺序清晰、职责边界明确。

## 阅读路径

```mermaid
flowchart TD
  A[README: 快速入口] --> B[DESIGN: 架构与设计]
  B --> C[MULTI_SOURCE_DEPS: 依赖机制]
  C --> D[EXAMPLES: 可直接复用配置]
  D --> E[AUDIT_REVIEW: 版本评估与风险]
  E --> F[AGENT_SKILL: 智能体接入]
```

## 文档职责图

```mermaid
flowchart LR
  R[README]
  D[DESIGN]
  M[MULTI_SOURCE_DEPS]
  E[EXAMPLES]
  A[AUDIT_REVIEW]
  S[AGENT_SKILL]

  R --> D
  R --> E
  D --> M
  M --> E
  D --> A
  E --> A
  R --> S
```

## 文档清单

| 文档                                             | 作用                         | 建议阅读顺序 |
| ------------------------------------------------ | ---------------------------- | ------------ |
| [`README.md`](../README.md)                      | 项目入口、命令速览、上手流程 | 1            |
| [`DESIGN.md`](./DESIGN.md)                       | 系统架构、核心流程、状态模型 | 2            |
| [`MULTI_SOURCE_DEPS.md`](./MULTI_SOURCE_DEPS.md) | 多源依赖设计与实现约束       | 3            |
| [`EXAMPLES.md`](./EXAMPLES.md)                   | 配置模板与场景示例           | 4            |
| [`AUDIT_REVIEW.md`](./AUDIT_REVIEW.md)           | 版本审计、风险与改进建议     | 5            |
| [`AGENT_SKILL.md`](./AGENT_SKILL.md)             | 智能体接入与技能规范         | 6            |

## 维护状态图

```mermaid
stateDiagram-v2
  [*] --> 草稿
  草稿 --> 已发布: 内容审阅通过
  已发布 --> 待更新: 代码行为变化
  待更新 --> 已发布: 文档修订完成
  已发布 --> [*]
```
