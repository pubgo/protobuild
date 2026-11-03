---
name: protobuild
description: "Manage protobuf projects with protobuild: vendor dependencies, generate code, lint/format, breaking checks and doctor diagnostics. Use in projects containing protobuf.yaml, or when the user asks to set up, initialize or diagnose protobuild."
license: MIT
metadata:
  author: pubgo
  version: "1.0.0"
compatibility: "Requires protobuild CLI and protoc. Run commands in the project root containing protobuf.yaml."
---

protobuild 是 protobuf 工程的构建管理 CLI：粘合层定位，直调 protoc 做代码生成；buf 只承担引擎能力（lsp / format / breaking）。本技能教你（智能体）在用户项目里正确、安全地驱动 protobuild 完成全生命周期。

## When to use

- 项目根存在 `protobuf.yaml`：需要下载依赖、生成代码、lint/format、查兼容性、诊断环境。
- 用户要求为 proto 项目**搭建/初始化** protobuild（`init`），或要求**诊断**构建环境（`doctor`）。
- 纯 buf/纯 protoc 工程且用户没有选择 protobuild 时，不要代做决定。

## Prerequisites

1. `protobuild` 与 `protoc` 可用：`protobuild --help`；缺失时先引导安装 protobuild。
2. protoc 插件（如 go-grpc、gateway）缺失时用 `protobuild install`（可 `-f` 强制更新）。
3. 环境异常（插件、配置、网络源）跑 `protobuild doctor`；`--fix` 只会 `go install` 缺失的 protoc-gen-go / protoc-gen-go-grpc，其余按提示手动处理。

## Workflows

### ① 新项目：初始化到首次生成
1. `protobuild init`（可选 `-t grpc-gateway|minimal|basic` 模板，`-f` 覆盖）。
2. 与用户确认 `protobuf.yaml`：root 目录、deps 依赖源（git/gomod/http/s3/gcs/local 六种，写法见 references/config.md）。
3. `protobuild vendor` 同步依赖到 `.proto/`，并生成 manifest。
4. `protobuild gen` 生成代码。**先 vendor 再 gen**。

### ② 日常改动
1. 编辑 proto 后 `protobuild gen`（依赖未就绪先 `vendor`）。
2. `protobuild lint`（AIP 规则）与 `protobuild format`（预览 `-d/--diff`，写回 `-w`）。
3. 改动涉及公开 API（消息字段、rpc、枚举）→ `protobuild breaking`（默认对比 base 分支，自动排除 vendor，不会因依赖升级误报）。

### ③ CI 门禁
`protobuild format --exit-code` + `protobuild lint` + `protobuild verify`（按 manifest 校验 `.proto/` 未被篡改）。

### ④ 故障排查
`protobuild doctor`（`--fix`）、`protobuild deps`（依赖状态）、`protobuild vendor -f`（`.proto/` 内容异常时重建）。症状-修法对照见 references/troubleshooting.md。

## Safety / Constraints

- **会写盘的命令**：`gen`、`format -w`、`install`、`buf`、`clean`、`init`、`vendor`、`update`——执行前告知用户文件/缓存会变化，执行后提示查看 diff。
- 命令须在含 `protobuf.yaml` 的项目根执行（全局 `-c` 可换配置文件路径）。
- **`.proto/` 是 vendor 产物，禁止手改**；内容异常（如 package 被文本替换弄坏）用 `protobuild vendor -f` 重建，`verify` 失败先怀疑手改。
- **buf 只认 `buf.yaml`**：改 root/vendor 后 `vendor`/`init` 会自动同步 buf.yaml；手动重生成用 `protobuild buf`，手写的 buf.yaml 默认不覆盖（接管需 `-f`）。
- `upgrade`/`version` 是 protobuild 自身管理，非用户项目操作，除非用户明确要求勿跑。

## 进一步阅读（references）

- 全命令与关键 flag 速查：[references/commands.md](references/commands.md)
- `protobuf.yaml` / `protobuf.plugin.yaml` 配置要点：[references/config.md](references/config.md)
- doctor / verify 与常见故障修法：[references/troubleshooting.md](references/troubleshooting.md)
