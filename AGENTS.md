# AGENTS.md

protobuild：面向 protobuf 工程的构建管理 CLI（Go）。定位是粘合层，**不取代 buf**：代码生成由本工具直调 protoc；buf 只承担引擎能力（lsp / format / breaking）。边界决策见 `docs/BUF_LSP.md`。

## 结构

- 入口：根 `main.go` → `cmd/protobuild.Main()`；CLI 框架是 **redant，不是 cobra**
- 新增子命令两件套：`cmd/protobuild/cmd_xxx.go` + 注册进 `cmd.go` 的 `Children`；并同步 `skills/protobuild/references/commands.md` 命令表（守卫测试 `TestCommandsDocumentedInSkillRefs` 会拦截遗漏）
- 配置经中间件 `withParseConfig()` 注入包级 `globalCfg`；位置参数用 `inv.Args`
- `internal/`：`depresolver`（多源依赖 git/gomod/http/s3/gcs/local）、`config`（protobuf.yaml 模型）、`typex`（Options/Commands 包装）、`protoutil`（protogen 生成辅助）、`modutil`（go.mod 解析）、`shutil`（shell 封装）
- `cmd/format`（内置 proto 格式化器）和 `cmd/linters`（内置 AIP lint 规则）是独立实现包；`cmd/formatcmd`、`cmd/webcmd` 分别挂为 format / web 子命令
- `skills/protobuild/` 是对外分发的 Agent Skill **唯一真源**（`npx skills add pubgo/protobuild` 安装）；原 `protobuild skills` 子命令已移除
- 本仓库自举使用自己：`protobuf.yaml`、`buf.yaml` 已入库；`.proto/` 是 gitignore 的 vendor 产物

## 命令

```bash
task all            # vet(gofumpt) -> lint(golangci-lint) -> test -> build
go test ./cmd/protobuild/ -run TestXxx   # 单包/单测
golangci-lint run   # v2 配置，issues.fix: true 会自动改写文件
go run . vendor     # 自举：proto:* / release:dry 等见 Taskfile.yml
```

CI（`.github/workflows/ci.yml`）真实执行 lint + build + test，必须本地全绿再提 PR。

## 关键坑（都是踩出来的）

1. **版本号来自 `//go:embed .version/VERSION`**（main.go `SetVersion`），不是 git tag。funk v2 的版本变量全部未导出，`-X` ldflags 注入无效。发版流程：先改 `.version/VERSION` → 合并 → 再打 `v*.*.*` tag（触发 goreleaser）。
2. **`.proto/` 禁止手改**。vendor 是纯文件拷贝；内容异常（如 package 被文本替换弄坏）直接 `protobuild vendor -f` 重建。`protobuild verify` 按 `.proto/manifest.yaml` 的 per-file sha256 校验，适合 CI 门禁。
3. **buf 只认 buf.yaml**，import 相对 module root 解析。`protobuf.yaml` 的 root/vendor 变更后 `vendor`/`init` 会自动同步 buf.yaml；手写的 buf.yaml 不会被覆盖（带 `protobuild buf --force` 接管提示）。
4. **main.go 兼任 protoc wrapper plugin**：root 命令若 stdin 收到 CodeGeneratorRequest（非 TTY）会走 `handleStdinPlugin`，不是普通 CLI 执行。
5. 仓库**没有 `vendor/` 目录**，构建勿加 `-mod vendor`。
6. **stacked PR 不要对 base 分支 `--delete-branch`**：会连下游 PR 一起关死且无法 reopen。分支保护要求 review，维护者用 admin 权限合并；合并顺序 = 先给下游 `gh pr edit --base`，再合上游。

## 其他

- `google/protobuf/*.proto` 在 buf 里靠内置 WKT 兜底——lsp 不报红不代表真的解析到了仓库内文件
- lint 引擎是 api-linter + protocompile（非 buf lint）；format 默认委托 `buf format`
- 文档入口 `docs/INDEX.md`；`.github/copilot-instructions.md` 是 Copilot 评审分类标记，仅供 review 参考
