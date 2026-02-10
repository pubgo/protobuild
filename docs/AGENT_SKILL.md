# Protobuild Agent Skill 指南（符合技能规范版）

面向把 **protobuild** 暴露为 Agent/LLM 的“技能（Skill）”，以下提供一份**规范化描述**，可直接放入你的技能清单/Manifest/知识库。

## 技能卡（Manifest 精简版，通用 Skills 规范）

```yaml
schema_version: v1
name_for_human: Protobuild
name_for_model: protobuild
description_for_human: 用于管理/生成/格式化 proto 的 CLI 工具
description_for_model: Run protobuild to vendor deps, generate code, lint/format proto projects.
version: 1.0.0
capabilities:
  categories: ["build", "lint", "format"]
  mode: cli
  safety:
    writes_files: ["gen", "format", "install"]
    needs_project_root: true
    requires_file: protobuf.yaml
tools:
  - name: protobuild_run
    description: Execute a protobuild subcommand
    input_schema:
      type: object
      properties:
        command:
          type: string
          enum: ["gen", "vendor", "lint", "format", "clean", "deps", "install", "doctor", "web"]
          description: Subcommand to run
        args:
          type: array
          items: { type: string }
          description: Optional extra flags, e.g. ["-w", "--diff"]
        working_dir:
          type: string
          description: Project root containing protobuf.yaml
      required: ["command"]
    output_schema:
      type: object
      properties:
        stdout: { type: string }
        stderr: { type: string }
        exit_code: { type: integer }
        changed_files:
          type: array
          items: { type: string }
      required: ["stdout", "stderr", "exit_code"]
usage_notes:
  - "Before gen, run vendor if deps are not ready."
  - "format -w / gen may modify files; surface diffs to the user."
  - "CI/readonly: prefer format --exit-code, lint."
```

## Agent Skills 规范的 SKILL.md 模板（agentskills.io/specification）

将下列文件命名为 `SKILL.md`，放置于技能目录（例如 `.agents/skills/protobuild/`），目录名需与 `name` 一致且为小写字母/数字/连字符。前置 YAML frontmatter 仅允许规范字段，其余内容为指令正文。

```markdown
---
name: protobuild
description: "Manage protobuf projects: vendor dependencies, generate code, lint/format. Use only when protobuf.yaml is present in the project root."
license: MIT  # 可选
metadata:
  author: pubgo
  version: "1.0.0"
compatibility: "Requires protobuild CLI, protoc, plugins; run in repo root with protobuf.yaml."  # 可选
allowed-tools: "Bash(protobuild:* )"  # 可选、实验字段，按需填写
---

## When to use
- 项目存在 `protobuf.yaml`，需要下载依赖、生成代码、lint/format proto。
- 优先执行 `vendor` 后再 `gen`；若只需检查格式/规范，使用 `lint` 或 `format --exit-code`。

## Safety / Constraints
- 会写盘的命令：`gen`、`format -w`、`install`（需告知用户）。
- 必须在包含 `protobuf.yaml` 的项目根目录执行。
- 如未 vendor 过依赖，先运行 `protobuild vendor`。

## Inputs
- command: one of [gen, vendor, lint, format, clean, deps, install, doctor, web]
- args: optional string list, e.g. ["-w", "--diff"]
- working_dir: project root (defaults to current CWD if已在项目根)

## Steps
1) 确认 `working_dir` 内存在 `protobuf.yaml`，否则提示用户补全。
2) 如 command=gen 且依赖未就绪，先运行 `protobuild vendor`。
3) 执行：`protobuild <command> <args...>`（在 working_dir）。
4) 收集 stdout/stderr/exit_code，若命令会写盘（gen/format -w/install），提示用户查看变更。
5) 失败时给出 stderr 摘要与下一步建议（如检查路径/依赖/插件）。

## Examples
- 下载依赖：`protobuild vendor`
- 生成代码：`protobuild gen`
- 仅检查格式：`protobuild format --exit-code`
- Lint：`protobuild lint`

```

### 目录与验证
- 目录：`.agents/skills/protobuild/SKILL.md`（或任意上层 `.agents/skills`，遵循查找规则）。
- 校验：`skills-ref validate ./.agents/skills/protobuild`

### 可选：agents/openai.yaml（用于 UI 外观和依赖声明）
将下列文件置于同级 `agents/openai.yaml`，以声明图标/描述或 MCP 依赖。

```yaml
interface:
  display_name: "Protobuild"
  short_description: "Proto build/lint/format tool"
  icon_small: "./assets/protobuild-16.png"
  icon_large: "./assets/protobuild-64.png"
  brand_color: "#0F9D58"
  default_prompt: "Use protobuild to vendor deps and generate proto code."

dependencies:
  tools: []  # 若需 MCP 依赖，可在此声明
```

## 能力概览
- 主要子命令：`vendor`、`gen`、`lint`、`format`、`clean`、`deps`、`install`、`doctor`
- 输入：命令名 + 追加参数（可选工作目录）
- 输出：结构化 `stdout` / `stderr` / `exit_code`，可选返回生成的文件列表
- 前置：项目根需已有 `protobuf.yaml`；若要生成代码，先确保依赖已 `vendor`

## OpenAI / 函数调用工具描述示例
```jsonc
{
  "type": "function",
  "function": {
    "name": "protobuild_run",
    "description": "Run protobuild commands (gen, vendor, lint, format, clean, deps, install, doctor)",
    "parameters": {
      "type": "object",
      "properties": {
        "command": {
          "type": "string",
          "enum": ["gen", "vendor", "lint", "format", "clean", "deps", "install", "doctor", "web"],
          "description": "Subcommand to run"
        },
        "args": {
          "type": "array",
          "items": {"type": "string"},
          "description": "Optional extra flags, e.g. ['-w', '--diff']"
        },
        "working_dir": {
          "type": "string",
          "description": "Project root containing protobuf.yaml"
        }
      },
      "required": ["command"]
    }
  }
}
```

> 上述与 Manifest 字段一一对应：`command/args/working_dir`，返回 `stdout/stderr/exit_code`。

### Agent 使用提示
- 若执行 `gen` 前未 `vendor`，先调用 `protobuild_run(command="vendor")`。
- `format -w` / `gen` 会修改文件，调用后可提示用户查看 diff。
- CI/只读场景：用 `format --exit-code`、`lint`，避免写盘。

## MCP（Model Context Protocol）工具定义示例
- 注册一个 action：`run_protobuild`
- 输入：同上 `command` / `args` / `working_dir`
- 白名单：只允许 `["gen", "vendor", "lint", "format", "clean", "deps", "install", "doctor"]`
- 服务端执行后返回：`stdout`、`stderr`、`exit_code`、可选 `changed_files`
- 可选：强制 `working_dir` 必须包含 `protobuf.yaml`

MCP manifest 片段示例：
```jsonc
{
  "name": "protobuild",
  "description": "Run protobuild for proto build/lint/format",
  "tools": [
    {
      "name": "run_protobuild",
      "description": "Execute a protobuild subcommand",
      "input_schema": {
        "type": "object",
        "properties": {
          "command": {"type": "string", "enum": ["gen", "vendor", "lint", "format", "clean", "deps", "install", "doctor"]},
          "args": {"type": "array", "items": {"type": "string"}},
          "working_dir": {"type": "string"}
        },
        "required": ["command"]
      }
    }
  ]
}
```

## 通用“命令工具”封装（无函数调用也可用）
- 命令：`protobuild`
- 允许子命令：`gen|vendor|lint|format|clean|deps|install|doctor`
- 执行目录：项目根
- 返回：标准输出/错误 & 退出码

## 推荐的 Agent 行为守则
1) 检查 `protobuf.yaml` 是否存在，不存在先提示用户。
2) 生成前先 `vendor`，再 `gen`。
3) 写盘操作需提示：`gen`、`format -w`、`install` 可能修改文件或安装依赖。
4) Lint/Format CI 模式：`format --exit-code`、`lint`。
5) 若命令失败，向用户返回 `stderr` + 建议（如检查路径/依赖）。

## 依赖准备
- 已安装 `protobuild` 可执行文件（建议固定版本）
- 已安装 `protoc` 与必要插件，或让 Agent 先调用 `protobuild install`

## 快速测试命令
```bash
protobuild vendor
protobuild gen
protobuild lint
protobuild format --exit-code
```

## 第三方使用指南（消费该 Skill）

1. **准备环境**
  - 安装 `protobuild`、`protoc` 及必要插件（可运行 `protobuild install`）。
  - 确保目标项目根目录存在 `protobuf.yaml`。

2. **放置 Skill**
  - 在目标项目内新建目录：`.agents/skills/protobuild/`
  - 将本仓库提供的 `SKILL.md`（位于 `docs/AGENT_SKILL.md` 中的模板部分）复制为 `.agents/skills/protobuild/SKILL.md`。
  - 如需 UI/依赖声明，可选择性复制 `agents/openai.yaml` 至同级。

3. **（可选）校验**
  - 安装 skills-ref 后运行：`skills-ref validate ./.agents/skills/protobuild`

4. **让 Agent 发现技能**
  - 使用支持 Agent Skills 的产品/CLI（如 Codex/Claude Code 等）在项目根启动；它会扫描 `.agents/skills` 目录。
  - 若未被自动加载，可重启 Agent/CLI。

5. **调用示例**
  - 明示调用：在对话中引用技能名 `protobuild`，例如 “使用 protobuild 先 vendor 再 gen”。
  - 函数/工具层：调用注册的 `protobuild_run`（参数 `command/args/working_dir`）。

6. **常见场景**
  - 拉取依赖：`protobuild vendor`
  - 生成代码：`protobuild gen`
  - 检查格式（CI）：`protobuild format --exit-code`
  - 规范检查：`protobuild lint`
