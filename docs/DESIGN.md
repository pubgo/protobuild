# 架构设计文档

## 文档定位

本文件说明系统架构、核心流程与运行状态。

- 上游文档：[`README.md`](../README.md)
- 下游文档：[`MULTI_SOURCE_DEPS.md`](./MULTI_SOURCE_DEPS.md)、[`EXAMPLES.md`](./EXAMPLES.md)
- 总览入口：[`INDEX.md`](./INDEX.md)

## 总体架构图

```mermaid
flowchart TB
  CLI[命令入口]
  CFG[配置层\nprotobuf.yaml / protobuf.plugin.yaml]
  CORE[核心层]
  EXEC[执行层]

  subgraph CORE
    DEP[依赖管理]
    GEN[代码生成]
    LINT[规则检查]
    FMT[代码格式化]
    WEB[可视化界面]
  end

  CLI --> CFG
  CFG --> CORE
  CORE --> EXEC

  EXEC --> PROTOC[protoc]
  EXEC --> TOOL[外部插件与工具]
```

## 模块关系图

```mermaid
flowchart LR
  P[cmd/protobuild] --> R[internal/config]
  P --> D[internal/depresolver]
  P --> U[internal/modutil]
  P --> S[internal/shutil]
  P --> T[internal/typex]
  P --> W[cmd/webcmd]
  P --> L[cmd/linters]
  P --> F[cmd/formatcmd]
```

## 核心流程图

### 生成流程

```mermaid
flowchart TD
  A[读取配置] --> B[遍历根目录]
  B --> C[按目录合并插件配置]
  C --> D[构建 protoc 命令]
  D --> E[执行生成]
  E --> F[输出代码]
```

### 依赖同步流程

```mermaid
flowchart TD
  A[读取 deps] --> B[识别依赖源]
  B --> C[下载或命中缓存]
  C --> D[复制 proto 到 vendor]
  D --> E[更新校验信息]
```

## 生命周期状态图

```mermaid
stateDiagram-v2
  [*] --> 未初始化
  未初始化 --> 已初始化: 创建配置
  已初始化 --> 依赖已同步: 执行 vendor
  依赖已同步 --> 可生成: 插件可用
  可生成 --> 已生成: 执行 gen
  已生成 --> 已检查: 执行 lint
  已检查 --> 已格式化: 执行 format
  已格式化 --> [*]
```

## 设计要点

1. 配置分层：根配置负责全局默认，目录配置负责局部覆盖。
2. 依赖解耦：通过统一依赖管理层屏蔽不同依赖源差异。
3. 执行分离：命令解析、构建命令、执行命令分层处理。
4. 可观测性：关键操作提供进度与错误上下文。
5. 可扩展性：插件、依赖源、规则引擎均可演进。

## 关联阅读

- 依赖细节：[`MULTI_SOURCE_DEPS.md`](./MULTI_SOURCE_DEPS.md)
- 可用配置：[`EXAMPLES.md`](./EXAMPLES.md)
- 版本评估：[`AUDIT_REVIEW.md`](./AUDIT_REVIEW.md)
