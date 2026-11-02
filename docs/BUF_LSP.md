# buf lsp 与 buf.yaml 集成记录

## 文档定位

本文件记录一次「buf lsp 无法解析 `import "retag/retag.proto"`」问题的完整排查过程、
结论，以及为解决该问题新增的 `protobuild buf` 子命令的设计说明。

- 总览入口：[`INDEX.md`](./INDEX.md)
- 相关文档：[`MULTI_SOURCE_DEPS.md`](./MULTI_SOURCE_DEPS.md)（vendor 目录的来源）

## 问题现象

`proto/example/v1/user.proto` 中：

```protobuf
import "google/protobuf/timestamp.proto"; // 正常
import "retag/retag.proto";               // buf lsp 报红: imported file does not exist
```

vendor 目录 `.proto/retag/retag.proto` 明明存在，但 LSP 无法解析；
`google/protobuf/*.proto` 同样不在工作区内，却不报红。

## buf 的 import 解析机制

buf CLI / buf lsp 加载 protobuf 的规则与 protoc 不同：

1. buf **只识别自己的配置**：`buf.yaml`（v1/v2）、`buf.work.yaml` / `buf.workspace.yaml`。
   `protobuf.yaml` 是 protobuild 的配置，buf 完全不感知。
2. buf 从被编辑文件向上查找最近的 `buf.yaml` 定位所属 module；
   **import 路径永远相对 module root**（`buf.yaml` 所在目录或 workspace 声明的目录）解析。
3. 找不到任何配置时，buf 把打开的项目根当作唯一 module root。
4. import 在工作区内找不到时，buf 用**内置的 well-known types**（`google/protobuf/*.proto`）兜底。

对应到本仓库（当时没有任何 buf 配置）：

| import | 解析行为 | 结果 |
| ------ | -------- | ---- |
| `google/protobuf/timestamp.proto` | 工作区找不到 → 内置 WKT 兜底 | 不报红（但是兜底，不是 vendor 的文件） |
| `retag/retag.proto` | 相对项目根找 `<root>/retag/retag.proto` | 实际在 `<root>/.proto/retag/`，报红 |

## 排查过程

1. 确认仓库内没有 `buf.yaml` / `buf.work.yaml`，只有 `protobuf.yaml`
   （`root: [proto]`，`vendor: .proto`，`includes: [proto, .proto]`）。
2. 用最小复现实验验证：无 buf.yaml 时 `buf build` 仅对 `retag/retag.proto` 报
   `imported file does not exist`，timestamp 正常 —— 与编辑器表现一致，确认是
   「module root 缺失 + WKT 兜底」两件事叠加。
3. 在仓库根新增 v2 buf.yaml，把 `proto` 和 `.proto` 都声明为 module，`buf build` 通过。
4. 加上 buf.yaml 后 `buf build` 仍报
   `unexpected absolute path in package declaration` —— 发现 vendor 产物
   `.proto/retag/retag.proto` 本身损坏：

   ```diff
   - package retag;
   + package .proto/retag;    // 非法
   - option go_package = "github.com/pubgo/protoc-gen-retag/pkg/retag;retagpb";
   + option go_package = "github.com/pubgo/protoc-gen-.proto/retag/pkg/.proto/retag;retagpb";
   ```

   即文件内容被盲目做了 `retag` → `.proto/retag` 的文本替换。
5. 克隆上游 `github.com/pubgo/protoc-gen-retag@v0.0.5` 对比，上游文件是正确的。
6. 验证 vendor 流程本身：分别用已安装的旧版 protobuild 二进制和当前源码在沙箱中
   重新执行 `protobuild vendor`，产物都是干净的纯拷贝 —— 排除工具链因素。
   结论：损坏来自一次本地的手工 / IDE 辅助操作（时间上与 go.mod、go.sum 的改动同批）。
7. 用上游原始文件覆盖损坏文件后，`buf build` 完整通过。

### 经验教训

- vendor 目录属于生成产物，出现内容异常时优先 `protobuild vendor` 重建，不要手改。
- 文本替换类操作（IDE 重构、全局替换）不要作用于 vendor 目录。
- buf 的「WKT 兜底」会掩盖 `google/protobuf` 依赖缺失，报红与否不能代表文件真的被解析到了。

## 解决方案

### 1. 新增 `protobuild buf` 子命令

`cmd/protobuild/cmd_buf.go`：根据 `protobuf.yaml` 生成 buf.yaml（v2）：

- `modules` = `root` 目录 + `vendor` 目录（去重、跳过不存在的目录），与 protobuild
  传给 protoc 的 import 路径（`includes`）保持一致。
- 文件已存在且内容一致时提示 up to date；内容不一致时报错，需 `--force` 覆盖。
- 生成后提示重启 buf lsp。

```bash
protobuild buf            # 生成 buf.yaml
protobuild buf --force    # 覆盖已存在且内容不同的 buf.yaml
```

生成结果示例（本仓库）：

```yaml
version: v2
modules:
  - path: proto
  - path: .proto
```

### 2. 与 protoc 流程的关系

`protobuild gen` 依旧走 protoc（`includes: [root..., vendor]`），buf.yaml 只服务于
buf CLI / buf lsp（IDE 解析、`buf build`、`buf lint` 等场景），两者互不影响。

### 局限与后续

- `buf lint` 会按 buf 规则检查 vendored 文件（如 `PACKAGE_DIRECTORY_MATCH`），
  如需使用，要在 buf.yaml 中为 vendored module 增加 ignore，本命令暂不生成 lint 配置。
- buf.yaml 中未映射 protobuf.yaml 的 `excludes`；如后续有需求，可按 module 前缀做映射。
- `protobuild vendor` 未来可考虑自动调用 `buf` 生成逻辑，保证 vendor 后 buf 配置同步。
