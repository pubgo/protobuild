# 设计：重做 protobuild init 默认体验（对齐自举配置）

- 日期：2026-11-03
- 分支：`chore/agents-skills`（续入 PR #35）
- 状态：已批准（2026-11-03，实施中按冒烟发现修订 T1 module 规则），实施见 docs/superpowers/plans/2026-11-03-init-defaults.md
- 前置讨论：用户选方向 A（照自举 protobuf.yaml 的样子重做 init 默认产出），并否决了"仅修不可用缺陷"（B）与"暂不动"（C）。

## 问题（现状为什么差）

1. **basic/minimal 模板没有 `base:` 段**——`detectGoModule()` 只在 `cfg.BasePlugin != nil` 时填充，两个模板 BasePlugin 为 nil，module/out/paths 全部缺失，生成的配置对 Go 项目直接不可用。
2. **所有模板的 WKT 依赖克隆整个 protocolbuffers 仓库**（`url: https://github.com/protocolbuffers/protobuf, path: src/google/protobuf`）。实测：protoc 裸 `-I` 即可解析 `google/protobuf/*.proto`（exit=0），lint 引擎（protocompile）自带标准库 resolver，buf 有内置 WKT——**这条依赖完全多余**，且是 vendor 最慢最大的一块。
3. **linter 零配置**：本仓库自举配置带精选规则块（`disabled_rules: [all]` + 3 条 gateway 规则），init 产物没有，lint 默认跑全量 AIP 噪音大。
4. **`createExampleProto` 是死代码**：定义了（示例 greeter.proto + go_package 检测）但 `runInit` 从未调用，init 后没有可编译对象，"下一步"提示让用户手动建 proto。

## 目标（钉死的决策）

### T1 三模板 + 交互模式统一补 `base`
- basic / minimal：`BasePlugin{Out: "./pkg", Paths: "source_relative"}`（与交互模式缺省一致，Go 项目最稳的映射方式）。
- grpc-gateway：保持现状（`Out: "./pkg", Paths: "import"`——gateway 代码要按 import path 布局，不动）。
- `detectGoModule()` 填充逻辑：**仅当 `BasePlugin.Paths == "import"`** 时填充 `<module>/pkg`。〔实施修订，冒烟证实〕原逻辑"BasePlugin 非 nil 即填"会产出 protoc-gen-go 直接拒绝的组合（`cannot use module= with paths=source_relative`，交互模式因此也在生成坏配置——即用户报"默认方式不太好"的最实锤病因）。source_relative 项目由 go_package + 目录镜像决定布局，无需 module。

### T2 WKT 依赖全面移除
- 三模板与 `interactiveInit()` 的 deps 里删除 `google/protobuf` 条目。
- grpc-gateway 保留 `google/api` 与 `protoc-gen-openapiv2/options`（**不是** WKT，protoc 不内置，必须 vendor）。
- basic/minimal 变成零 deps——vendor 输出 "No dependencies configured" 属正常。

### T3 linter 规则块进模板（按模板形态差异化）
- grpc-gateway：带自举同款规则块（`enabled_rules: [core::0131::http-method, core::0131::http-body, core::0235::plural-method-name]` + `disabled_rules: [all]` + `format_type: yaml`）——该模板的产出天然是 AIP 风格 HTTP 注解服务，这组规则就是为它调的。
- basic / minimal：不带 linter 块（没有 gateway 语义可校验，硬塞 `[all]` 禁用等于静音 lint；用户按需自配）。

### T4 接线 `createExampleProto`
- `runInit` 在创建 root 目录后：目录无 .proto 文件时生成示例 `proto/example/v1/greeter.proto`（函数已存在，go_package 用检测到的 module）。
- 交互模式同样接线。
- "下一步"提示同步修正（不再让用户手动建 proto；无 deps 时把"1. vendor"改为条件输出）。

## 非目标

- 不改 init 的 CLI 面（flag、模板名、交互问题数）。
- 不动 vendor/gen/lint 的执行逻辑（protoc WKT 解析是其既有行为，不新增 -I）。
- 不动 `protobuf.lock`（已砍）与 checksum 机制。

## 约束

- 守卫测试 `TestCommandsDocumentedInSkillRefs` 必须保持绿：init 行若行为描述变化，同步 `references/commands.md`（写"生成示例 proto"）。
- `skills/protobuild/references/config.md`、SKILL.md Workflows ① 若与新默认冲突（如 WKT 写法示例），一并修正——skill 是对外承诺。
- TDD：模板内容与 runInit 行为写真实断言测试（临时目录跑 runInit，读产物 protobuf.yaml 断言 base/deps/linter/example.proto）。
- `task all` 全绿；PR #35 在 init 工作完成前不合并（用户决定）。

## Approaches（已选）

- **A（选定）**：照自举配置重做默认产出（T1–T4 打包）。
- B：只修不可用缺陷（base + WKT）。被否：linter/example 的"不好用"仍是首次体验的一部分，修一半等于没修。
- C：另开 PR。被否：用户要求 #35 处理完 init 再合并。

## 留给实现者的

- runInit 调用顺序与幂等（root 目录已有 proto 时 example 不生成——`createExampleProto` 已自带该检查）。
- minimal 模板是否连 go-grpc 都不带（现状：basic 带 go+go-grpc，minimal 只带 go——保持）。

## 验收

1. `task all` 全绿（vet→lint→test→build）。
2. 新测试：模板四路断言（base 存在、无 WKT dep、gateway 带 linter 块）+ runInit 产物断言（basic：example.proto 生成、source_relative **无** module；grpc-gateway：import + module 检测填充）。
3. 现场冒烟：scratch 目录 `init -t basic` → `vendor`（应秒过：无依赖）→ `gen`（示例 greeter 生成 .pb.go 成功）→ `lint`（gateway 模板目录跑通）。
4. 技能文档一致性：commands.md init 行、SKILL.md workflow ①、config.md 无 WKT 残留误导；守卫测试绿。
5. `go run . init -t basic` 在本仓库外新目录跑通全流程后，README 若含 init 示例则核对（构建时 grep）。
