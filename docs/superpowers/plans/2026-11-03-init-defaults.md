# init 默认体验重做 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按已批准设计（T1-T4）重做 `protobuild init` 默认产出：三模板+交互补 `base`、删 WKT 依赖、gateway 带 linter 块、接线示例 proto。

**Architecture:** 纯 `cmd/protobuild/cmd_init.go` 数据改造 + runInit 接线；产物正确性用真实断言测试锁住；技能文档（skills/protobuild/references）同步对外承诺，由既有守卫测试兜命令名覆盖。

**Tech Stack:** Go（标准测试，t.Chdir 隔离 cwd）、YAML 产物断言。

**Spec:** `docs/superpowers/specs/2026-11-03-init-defaults-design.md`

**分支：** 当前检出 `chore/agents-skills`，续 PR #35。

## Global Constraints

- TDD：每个行为改动先见 RED 再 GREEN；测试放 `cmd/protobuild/cmd_init_test.go`（package protobuild）。
- 不动 init CLI 面（flag/模板名/交互问题数）；不动 vendor/gen/lint 执行逻辑。
- 守卫测试保持绿；`task all` 全绿才可提交推送。
- 技能文档与新行为不一致处必须同步（skill 是对外承诺）。

## File Structure

- Modify `cmd/protobuild/cmd_init.go` — 模板数据（base/WKT/linter）、interactiveInit deps、runInit 接线 example 与提示
- Create `cmd/protobuild/cmd_init_test.go` — 模板与 runInit 产物断言
- Modify `skills/protobuild/references/commands.md` — init 行补"生成示例 proto"
- Modify `skills/protobuild/SKILL.md` — Workflows ① 措辞对齐（无 deps 时 vendor 为空操作）
- Commit `docs/superpowers/specs/2026-11-03-init-defaults-design.md` + 本计划

---

### Task 1: 模板数据重做（T1 base + T2 删 WKT + T3 gateway linter）

**Files:**
- Modify: `cmd/protobuild/cmd_init.go`（initTemplates 18-66、interactiveInit 196-216）
- Test: `cmd/protobuild/cmd_init_test.go`

**Interfaces:**
- Consumes: `initTemplates map[string]*config.Config`（包级）、`config.BasePluginCfg`、`config.Linter`
- Produces: basic/minimal 的 `BasePlugin != nil`；三模板与交互模式 deps 无 `google/protobuf`；grpc-gateway `Linter != nil`（Task 2 的 runInit 产物测试以此为输入）

- [ ] **Step 1: RED** — 新建 `cmd_init_test.go`：

```go
package protobuild

import (
	"strings"
	"testing"
)

func TestTemplatesHaveBase(t *testing.T) {
	for _, name := range []string{"basic", "minimal", "grpc-gateway"} {
		tmpl := initTemplates[name]
		if tmpl.BasePlugin == nil {
			t.Fatalf("template %q: BasePlugin nil, generated config unusable for Go projects", name)
		}
		if tmpl.BasePlugin.Out == "" || tmpl.BasePlugin.Paths == "" {
			t.Errorf("template %q: BasePlugin out/paths empty", name)
		}
	}
}

func TestTemplatesNoWKTDep(t *testing.T) {
	for _, name := range []string{"basic", "minimal", "grpc-gateway"} {
		for _, dep := range initTemplates[name].Depends {
			if dep.Name == "google/protobuf" {
				t.Errorf("template %q: WKT dep %q must be dropped (protoc resolves WKT natively)", name, dep.Url)
			}
		}
	}
}

func TestGatewayTemplateHasLinterBlock(t *testing.T) {
	l := initTemplates["grpc-gateway"].Linter
	if l == nil || l.Rules == nil || len(l.Rules.EnabledRules) == 0 {
		t.Fatal("grpc-gateway template must ship the curated AIP rule block")
	}
	for _, want := range []string{"core::0131::http-method", "core::0131::http-body", "core::0235::plural-method-name"} {
		if !strings.Contains(strings.Join(l.Rules.EnabledRules, ","), want) {
			t.Errorf("enabled_rules missing %q", want)
		}
	}
}

func TestBasicMinimalNoLinter(t *testing.T) {
	for _, name := range []string{"basic", "minimal"} {
		if initTemplates[name].Linter != nil {
			t.Errorf("template %q must not ship disabled_rules:[all] (silences lint for generic projects)", name)
		}
	}
}
```

- [ ] **Step 2: 跑 RED** — `go test ./cmd/protobuild/ -run 'TestTemplates|TestGateway|TestBasicMinimal' -v`，预期 base/WKT/linter 测试 FAIL（basic/minimal BasePlugin nil；WKT dep 存在；gateway Linter nil），`TestBasicMinimalNoLinter` PASS（现状即合规，允许通过）。

- [ ] **Step 3: GREEN** — 改 `cmd_init.go`：
  - basic/minimal：加 `BasePlugin: &config.BasePluginCfg{Out: "./pkg", Paths: "source_relative"}`。
  - 三模板删 `google/protobuf` dep（grpc-gateway 的 google/api、protoc-gen-openapiv2/options 保留）。
  - grpc-gateway：加 `Linter: &config.Linter{Rules: &config.LinterRules{EnabledRules: [...3条], DisabledRules: []string{"all"}}, FormatType: "yaml"}`（字段名以 internal/config 定义为准）。
  - interactiveInit：删 cfg.Depends 里的 google/protobuf 条目（Depends 置 nil 或空）。
- [ ] **Step 4: 跑 GREEN** — 同 Step 2 命令，预期全 PASS。
- [ ] **Step 5: 提交** `git add cmd/protobuild/cmd_init.go cmd/protobuild/cmd_init_test.go && git commit -m "feat(init): templates carry base config, drop redundant WKT dep, gateway ships linter rules"`

---

### Task 2: runInit 接线示例 proto + 修正提示（T4）

**Files:**
- Modify: `cmd/protobuild/cmd_init.go`（runInit 132-154）
- Test: `cmd/protobuild/cmd_init_test.go`

**Interfaces:**
- Consumes: `createExampleProto(protoDir string) error`（已存在）、Task 1 的模板形态
- Produces: `runInit -t basic` 在空目录产出 protobuf.yaml + buf.yaml + proto/example/v1/greeter.proto

- [ ] **Step 1: RED** — 追加产物级测试（用 t.Chdir 到 t.TempDir，避免污染包内其他测试的 cwd）：

```go
func TestRunInitProducesUsableProject(t *testing.T) {
	if os.Getenv("PB_INIT_RUN_TEST") == "1" {
		// run via subprocess in temp dir; marker env keeps go test happy
	}
	t.Chdir(t.TempDir())
	if err := os.WriteFile("go.mod", []byte("module example.com/demo\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit("basic", false); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	for _, want := range []string{"protobuf.yaml", "buf.yaml", "proto/example/v1/greeter.proto"} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("expected artifact %s: %v", want, err)
		}
	}
	content := string(mustRead(t, "protobuf.yaml"))
	if !strings.Contains(content, "module: example.com/demo/pkg") {
		t.Errorf("config must contain detected module, got:\n%s", content)
	}
	if strings.Contains(content, "protocolbuffers") {
		t.Error("config must not contain WKT github dep")
	}
}
```

（`mustRead` 为测试内小 helper；os 需 import）

- [ ] **Step 2: 跑 RED** — `go test ./cmd/protobuild/ -run TestRunInitProducesUsableProject -v`，预期 FAIL（greeter.proto 不存在——createExampleProto 未接线）。

- [ ] **Step 3: GREEN** — runInit 在 `MkdirAll` 循环后调用 `createExampleProto`（root 首个目录，函数自带"已有 proto 则跳过"检查，错误 warn 不中断 init）；"下一步"提示改为条件输出：deps 为空时跳过 vendor 行，示例已生成时把"创建 .proto"改为"直接 gen 示例或替换为你的 proto"。
- [ ] **Step 4: 跑 GREEN** — 同 Step 2，预期 PASS。
- [ ] **Step 5: 提交** `git commit -am "feat(init): wire up example proto generation and conditional next-steps"`

---

### Task 3: 端到端冒烟 + 技能文档同步

**Files:**
- Modify: `skills/protobuild/references/commands.md`（init 行）
- Modify: `skills/protobuild/SKILL.md`（Workflows ①）
- Create/Commit: spec + plan 入库

**Interfaces:**
- Consumes: Task 1/2 的最终行为
- Produces: 对外文档与新默认一致；守卫测试仍绿

- [ ] **Step 1: 冒烟** — scratch 目录：`go build -o $T/pb .`，`$T/pb init -t basic`（带 go.mod）→ `$T/pb vendor`（预期 "No dependencies configured" 秒过）→ `$T/pb gen`（预期生成 pkg/example/v1/*.pb.go，需 protoc-gen-go 在 PATH——本机已装，缺则记录为环境前置）。输出留档。
- [ ] **Step 2: 文档同步** — commands.md init 行用途加"生成示例 proto"；SKILL.md Workflows ① 第 2-4 步措辞对齐（deps 可选、init 自带示例、无 deps 时 vendor 空操作）；grep `protocolbuffers|WKT` skills/ 确认无残留误导。
- [ ] **Step 3: 全门禁** — `task all` 全绿（含守卫测试）。
- [ ] **Step 4: 提交** — spec+plan+文档改动：`git commit -m "docs: sync init behavior into skill refs; commit design and plan"`

---

## Review Focus

1. **paths=source_relative vs import 的选择**影响生成布局与 retag 需求——basic/minimal 用 source_relative（不依赖 retag），评审确认与"Go 项目开箱可用"一致。
2. **createExampleProto 的 go_package 与 base.module 一致性**：module 检测成功时示例 go_package=module/pkg/example/v1；检测失败回落 placeholder——产物测试覆盖前者，placeholder 分支为人工可判断。
3. **交互模式无模板名可断言**：其 deps 删 WKT 由代码评审 + `TestRunInit...`（模板路）间接覆盖；交互路径冒烟走一遍默认回车。
4. **vendor "No dependencies configured" 新默认下成为常见输出**——确认 verify/lint 对零 deps 项目不误报。
5. **技能文档漂移**：init 行为变了，SKILL.md/commands.md 若不同步即对外撒谎——Task 3 Step 2 承接。

## 执行提醒

- Task 1/2 严格 RED→GREEN；模板测试纯数据断言，快。
- t.Chdir 防 cwd 泄漏；runInit 会写真实文件（在 t.TempDir 内）。
