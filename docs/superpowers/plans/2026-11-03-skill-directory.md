# protobuild Agent Skill 目录化分发 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 protobuild 的对外 Agent Skill 从 CLI 内联模板改为仓库真实目录 `skills/protobuild/`（`npx skills add` 分发），删除 `protobuild skills` 子命令，并加入防漂移守卫测试。

**Architecture:** 单一真源 = `skills/protobuild/`（SKILL.md 渐进披露主入口 + references/ 三个自包含参考文件 + agents/openai.yaml）。CLI 侧纯删除。一个 Go 测试遍历 `Main().Children` 断言每个子命令都被 `references/commands.md` 记载，机制性防漂移。

**Tech Stack:** Go（redant CLI 框架、标准测试）、Markdown（Agent Skills 规范 frontmatter）。

**Spec:** `docs/superpowers/specs/2026-11-03-skill-directory-design.md`

**分支：** 当前检出 `chore/agents-skills`（已建，基线 `go test ./...` 全绿），直接在其上实施。

## Global Constraints

- 全仓 `task all`（vet=gofumpt → lint=golangci-lint v2 → test → build）必须全绿；CI 同标准。
- 不动 `.version/VERSION`，不发版。
- 不改其他子命令行为；`skills` 子命令纯删除，不做 tombstone。
- 命令/flag 事实必须取自 `go run . <cmd> --help` 实际输出或源码，不得凭记忆编造。
- SKILL.md 正文中文、frontmatter description 英文为主；正文精短（≤ ~100 行），细节走 references。
- 提交信息英文 conventional commits 风格。
- 删除 `skills/` 以外不得新增 Go 依赖；`skills/` 目录不得被 `.gitignore` 命中（`git check-ignore` 验证）。

## File Structure

- Create `skills/protobuild/SKILL.md` — 技能主入口（frontmatter + 触发条件 + 前置 + 工作流 + Safety + 索引导航）
- Create `skills/protobuild/references/commands.md` — 16 命令速查表（守卫测试锚点）
- Create `skills/protobuild/references/config.md` — protobuf.yaml / protobuf.plugin.yaml 操作要点
- Create `skills/protobuild/references/troubleshooting.md` — doctor/verify 用法 + 常见故障表
- Create `skills/protobuild/agents/openai.yaml` — OpenAI 兼容元数据（无 icon 字段）
- Create `cmd/protobuild/skill_docs_test.go` — 守卫测试（package protobuild）
- Modify `cmd/protobuild/cmd.go` — 删除模板串、`newSkillsCommand` 及注册、两个 helper
- Modify `README.md:109` — 删 `skills` 表行
- Rewrite（部分）`docs/AGENT_SKILL.md` — 接入方式与迁移提示
- Modify + 首次入库 `AGENTS.md` — 三件套改为两件套 + 同步 commands.md
- Commit 本计划与 spec（`docs/superpowers/`）

---

### Task 1: 删除 `protobuild skills` 子命令

**Files:**
- Modify: `cmd/protobuild/cmd.go`（`defaultSkillContent` 约 41-79、`defaultOpenAIContent` 约 81-91、`Children` 中 `newSkillsCommand(),` 约 144、`newSkillsCommand` 函数约 154-215、`normalizeTemplateIndent` 约 217-219、`writeFileIfNeeded` 约 221-231）
- Modify: `README.md:109`（`| \`skills\` | 生成智能体技能模板 |` 行）

**Interfaces:**
- Consumes: 无
- Produces: `Main().Children` 共 16 个命令：init, doctor, gen, vendor, update, verify, buf, breaking, install, lint, format, deps, clean, web, version, upgrade（Task 2 的测试遍历即此集合；`Use` 首词与命令名一致，如 `update [name...]` → `update`）。README 命令表不含 `skills`。

- [ ] **Step 1: 删除 cmd.go 六处**：两个包级字符串变量（保留 `var` 块其余成员）、`newSkillsCommand(),` 注册行、`newSkillsCommand` 函数、`normalizeTemplateIndent`、`writeFileIfNeeded`（后两者已确认仅此处使用、无测试引用）。

- [ ] **Step 2: 清残留引用**

```bash
grep -rnE 'defaultSkillContent|defaultOpenAIContent|newSkillsCommand|normalizeTemplateIndent|writeFileIfNeeded' cmd/ internal/ *.go
```

预期：0 命中。

- [ ] **Step 3: 修 unused imports 并构建**：`go build ./...`；若有 `imported and not used`（候选：`errors`、`io/fs`、`path/filepath`，按编译器提示删除）。预期：构建成功。

- [ ] **Step 4: 验证 help 不再列出 skills**：`go run . --help` 输出中 grep `skills`，预期 0 命中。

- [ ] **Step 5: 删 README:109 的 `skills` 行。**

- [ ] **Step 6: 提交**

```bash
git add cmd/protobuild/cmd.go README.md
git commit -m "refactor(cli)!: remove redundant skills subcommand in favor of repo skills dir"
```

---

### Task 2: 守卫测试 + `references/commands.md`

**Files:**
- Create: `skills/protobuild/references/commands.md`
- Test: `cmd/protobuild/skill_docs_test.go`

**Interfaces:**
- Consumes: Task 1 的 `Main().Children`（16 命令）。
- Produces: `skills/protobuild/references/commands.md` 含全部 16 个命令名（反引号包裹，守卫测试据此断言；Task 5 的 AGENTS.md 指引引用此文件路径）。

- [ ] **Step 1: 写失败测试** `cmd/protobuild/skill_docs_test.go`（package protobuild）：

```go
package protobuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommandsDocumentedInSkillRefs guards skills/protobuild/references/commands.md
// against command-list drift: every subcommand registered in Main() must appear
// in the reference doc, wrapped in backticks.
func TestCommandsDocumentedInSkillRefs(t *testing.T) {
	ref := filepath.Join("..", "..", "skills", "protobuild", "references", "commands.md")
	data, err := os.ReadFile(ref)
	if err != nil {
		t.Fatalf("read %s: %v", ref, err)
	}
	md := string(data)
	for _, c := range Main().Children {
		name := strings.Fields(c.Use)[0]
		if !strings.Contains(md, "`"+name+"`") {
			t.Errorf("command %q missing from %s (match by backtick-wrapped name)", name, ref)
		}
	}
}
```

- [ ] **Step 2: 跑测试确认失败**：`go test ./cmd/protobuild/ -run TestCommandsDocumentedInSkillRefs`，预期 FAIL（文件不存在 → `read ...: no such file or directory`）。

- [ ] **Step 3: 写 `skills/protobuild/references/commands.md`**：命令速查表，每行一个命令，**命令名必须反引号包裹**。事实来源：`go run . <cmd> --help` 逐条取（不得凭记忆）。结构：

```markdown
# protobuild 命令速查

> 本表由守卫测试 TestCommandsDocumentedInSkillRefs 保证命令名覆盖完整；新增子命令时同步本表。

| 命令 | 用途 | 关键 flag | 写盘 |
|------|------|-----------|------|
| `init` | ... | `--template` | 是 |
| `doctor` | ... | `--fix` | ... |
（16 行全部列出；写盘列对 gen、format -w、install、buf、clean 标"是"）
```

- [ ] **Step 4: 跑测试确认通过**：同上命令，预期 PASS。

- [ ] **Step 5: 负向验证守卫非摆设**：临时从 commands.md 删一行（如 `breaking`），重跑测试预期 FAIL 且报 `command "breaking" missing`；恢复该行，重跑预期 PASS。

- [ ] **Step 6: 提交**

```bash
git add cmd/protobuild/skill_docs_test.go skills/protobuild/references/commands.md
git commit -m "test: add skill docs drift guard with commands reference"
```

---

### Task 3: `SKILL.md` 主入口 + `agents/openai.yaml`

**Files:**
- Create: `skills/protobuild/SKILL.md`
- Create: `skills/protobuild/agents/openai.yaml`

**Interfaces:**
- Consumes: Task 2 的 `references/commands.md`（正文引用其路径）。
- Produces: 可被 `skills-ref validate` 校验的技能目录 `skills/protobuild/`；Task 4 往同目录追加两个 references 文件、Task 5 的 AGENT_SKILL.md 指向 `skills/protobuild/SKILL.md`。

- [ ] **Step 1: 写 `skills/protobuild/SKILL.md`**，frontmatter 照抄（description 触发范围含 init 场景，是 spec 修订点）：

```markdown
---
name: protobuild
description: "Manage protobuf projects with protobuild: vendor dependencies, generate code, lint/format, breaking checks and doctor diagnostics. Use in projects containing protobuf.yaml, or when the user asks to set up, initialize or diagnose protobuild."
license: MIT
metadata:
  author: pubgo
  version: "1.0.0"
compatibility: "Requires protobuild CLI and protoc. Run commands in the project root containing protobuf.yaml."
---
```

正文（中文，≤ ~100 行）分节，每节必须覆盖：
- **When to use**：有 `protobuf.yaml` 的 proto 工程；用户要求 init/搭建/诊断 protobuild。
- **Prerequisites**：`protobuild` CLI、protoc 与插件缺失时 `protobuild install`。
- **Workflows**：① 新项目初始化：`init` → 配置 protobuf.yaml（deps 多源）→ `vendor` → `gen`；② 日常改动：改 proto → `gen` → `lint`/`format`，动公开 API 先 `breaking`；③ CI 门禁：`format --exit-code`、`lint`、`verify`；④ 故障：`doctor`、`doctor --fix`。
- **Safety**：写盘命令（gen、format -w、install、buf、clean）须先告知用户；`.proto/` 是 vendor 产物禁止手改，异常用 `vendor -f` 重建；buf 只认 buf.yaml，root/vendor 变更由 `protobuild buf`/init 同步。
- **Command index / Config / Troubleshooting**：链接 `references/commands.md`、`references/config.md`、`references/troubleshooting.md`（后者 Task 4 创建，路径先行写死）。

- [ ] **Step 2: 写 `skills/protobuild/agents/openai.yaml`**（去掉 icon 字段后的原元数据）：

```yaml
interface:
  display_name: "Protobuild"
  short_description: "Proto build/lint/format tool"
  brand_color: "#0F9D58"
  default_prompt: "Use protobuild to vendor deps and generate proto code."

dependencies:
  tools: []
```

- [ ] **Step 3: 校验 frontmatter 与触发描述**：`npx -y skills-ref validate skills/protobuild`（离线或工具不可用时人工核对：`name` 小写连字符、`description` 单行非空、frontmatter 为文件首个块）。预期：pass。
- [ ] **Step 4: 检查 `.gitignore` 不吞文件**：`git check-ignore -v skills/protobuild/SKILL.md skills/protobuild/agents/openai.yaml skills/protobuild/references/commands.md`，预期均无命中（非 0 退出码即未命中）。
- [ ] **Step 5: 提交**

```bash
git add skills/protobuild/SKILL.md skills/protobuild/agents/openai.yaml
git commit -m "feat(skills): author protobuild agent skill entry"
```

---

### Task 4: `references/config.md` + `references/troubleshooting.md`

**Files:**
- Create: `skills/protobuild/references/config.md`
- Create: `skills/protobuild/references/troubleshooting.md`

**Interfaces:**
- Consumes: Task 3 的 SKILL.md 链接路径（文件名必须一字不差）。
- Produces: 无（叶子文档）。

- [ ] **Step 1: 写 `config.md`**：智能体操作要点（受众切分 C：不整篇抄 docs/EXAMPLES.md）。字段事实以 `internal/config/` 模型为准（读源码），覆盖：protobuf.yaml 关键字段（root/roots、vendor、deps 多源类型 git/gomod/http/s3/gcs/local 及各自最小配置）、`protobuf.plugin.yaml` 目录级覆盖规则、与 buf.yaml 的关系（buf 只认 buf.yaml，改 root/vendor 后跑 `protobuild buf`）。结尾一行注明完整人类向示例见仓库 docs/EXAMPLES.md。
- [ ] **Step 2: 写 `troubleshooting.md`**：`doctor`（检查项与 `--fix` 能力）、`verify`（按 `.proto/manifest.yaml` per-file sha256 校验 vendor 产物，适合 CI 门禁）、故障→修法表：gen 失败先 `vendor`；`.proto/` 内容异常（如 package 被文本替换弄坏）→ `protobuild vendor -f` 重建；buf/LSP 报红 → 检查 buf.yaml 是否经 `protobuild buf` 同步；版本显示异常 → 版本号来自 `.version/VERSION` go:embed，非 git tag。
- [ ] **Step 3: 事实抽查**：对照源码/--help 核对两份文件里每条命令与字段名（至少 `deps` 源类型 6 种、`--fix`、`vendor -f`、manifest 路径），发现编造立即改。
- [ ] **Step 4: 提交**

```bash
git add skills/protobuild/references/config.md skills/protobuild/references/troubleshooting.md
git commit -m "docs(skills): add config and troubleshooting references"
```

---

### Task 5: 文档同步（AGENT_SKILL.md / AGENTS.md 入库 / INDEX）

**Files:**
- Modify: `docs/AGENT_SKILL.md`
- Modify: `AGENTS.md`（当前 untracked，本次首次入库）
- Modify: `docs/INDEX.md`（仅核对，必要时）
- Modify: `docs/superpowers/specs/2026-11-03-skill-directory-design.md`（状态行改"已批准"）

**Interfaces:**
- Consumes: `skills/protobuild/SKILL.md`、`references/commands.md` 路径。
- Produces: AGENTS.md 中的新增子命令指引（后续维护约定）。

- [ ] **Step 1: 改 AGENTS.md**（第 8 行三件套句 + 结构节新增一条）：
  - 三件套改为："新增子命令两件套：`cmd/protobuild/cmd_xxx.go` + 注册进 `cmd.go` 的 `Children`；并同步 `skills/protobuild/references/commands.md` 命令表（守卫测试 `TestCommandsDocumentedInSkillRefs` 会拦截遗漏）"。
  - "结构"节增补一行：`skills/protobuild/` = 对外分发的 Agent Skill（唯一真源，`npx skills add pubgo/protobuild` 安装；原 `protobuild skills` 子命令已移除）。

- [ ] **Step 2: 重写 `docs/AGENT_SKILL.md` 的接入部分**：删除旧的 CLI 通道描述与手抄命令 enum（60 行 `enum` 列表）；接入方式改为两条：`npx skills add pubgo/protobuild`（支持 `-g` 全局）或直接拷贝 `skills/protobuild/` 目录进项目智能体技能位置；新增"迁移提示"一节：旧 `protobuild skills` 已移除，此前生成的 `.agents/skills/protobuild/` 请删除或用新装覆盖，避免旧漂移内容遮蔽新技能；"推荐调用策略/最小操作清单"保留并指向 `skills/protobuild/SKILL.md`；流程图"注册技能描述"步骤同步改为 npx 安装。文档定位行注明真源在 `skills/protobuild/`。

- [ ] **Step 3: 核对 `docs/INDEX.md`**：AGENT_SKILL 行描述（"智能体接入与技能规范"）仍准确则不动；若 INDEX 有文档清单结构，为 `skills/protobuild/` 增一行入口说明。

- [ ] **Step 4: 更新 spec 状态行**：`docs/superpowers/specs/2026-11-03-skill-directory-design.md` 中 `- 状态：待批准` 改为 `- 状态：已批准（2026-11-03），实施见 docs/superpowers/plans/2026-11-03-skill-directory.md`。

- [ ] **Step 5: 残留 grep**：

```bash
grep -rnE 'protobuild skills|defaultSkillContent' README.md AGENTS.md docs/*.md cmd/ internal/ main.go
```

预期：仅 `docs/AGENT_SKILL.md` 迁移提示一节允许提及历史命令；其余 0 命中。

- [ ] **Step 6: 提交（含 AGENTS.md 首次入库与本计划/spec）**

```bash
git add AGENTS.md docs/AGENT_SKILL.md docs/INDEX.md docs/superpowers/
git commit -m "docs: switch skill distribution to repo skills dir; onboard AGENTS.md"
```

---

### Task 6: 全量验收

**Files:**
- 无新增（发现问题则回改相应文件并补充提交）。

**Interfaces:**
- Consumes: 全部前序任务产物。

- [ ] **Step 1: 全绿门禁**：`task all`（gofumpt → golangci-lint → go test → build），预期全部通过。若 `golangci-lint` 的 `issues.fix` 改写了文件，把改动并进最近相关提交。
- [ ] **Step 2: CLI 面验证**：`go run . --help` 无 `skills`；`go run . version` 正常。
- [ ] **Step 3: 技能面验证**：`skills/` 入库文件清单 = `git ls-files skills/` 应有 5 个（SKILL.md、openai.yaml、3×references）；守卫测试绿。
- [ ] **Step 4: spec 验收对照**：逐条勾 `docs/superpowers/specs/2026-11-03-skill-directory-design.md` 验收 1-6；`npx skills add pubgo/protobuild` 的真实安装依赖推送与默认分支，本 PR 阶段以 `skills-ref validate` 通过代替，验证推迟到 push 后（如条件允许跑一次并注明）。
- [ ] **Step 5: 最终提交（如有修补）**：`git add -A && git commit -m "chore: final acceptance fixes"`（无改动则跳过）。

---

## Review Focus

本计划对应 spec 隐含、但任务测试不直接覆盖的失败模式（人使用此软件时最可能踩的）：

1. **新子命令加入但 commands.md 只写了名没写 flag 行** → 守卫测试绿但表已陈旧。预期：flag 列与 `--help` 一致属人工评审项，Task 2 Step 3 已规定逐条取自 `--help`；评审者抽查 3 条。
2. **合并前文档宣称 `npx skills add pubgo/protobuild` 可用** → 用户此刻安装会失败（默认分支尚无该目录）。预期：Task 6 Step 4 显式推迟真实安装验证；AGENT_SKILL.md 文案避免"立即可用"的现在时承诺历史分支。
3. **旧命令生成的 `.agents/skills/protobuild/` 与新技能同名共存** → harness 按解析顺序可能加载旧漂移版。预期：迁移提示（Task 5 Step 2）给出删除/覆盖指令。
4. **放宽后的 description 触发过界**（在非 protobuild 的 proto 项目抢答，误导用户装 protobuild）。预期：Task 3 Step 1 的措辞以"set up, initialize or diagnose **protobuild**"限定，评审对照 spec 触发条款。
5. **`skills/` 被 gitignore 规则误吞**（`vendor`、`*.svg` 等既有规则）。预期：Task 3 Step 4 的 `git check-ignore -v` 三文件验证，未来加图标需 `git add -f` 或补 ignore 例外。

## 执行提醒

- 逐任务提交，保持每个任务独立可评审。
- Task 1 有编译副作用（unused imports），先构建再提交。
- 内容型任务（2/3/4）的事实必须回源仓库 `--help`/源码，spec 反目标"与仓库实际不符"是第一红线。
