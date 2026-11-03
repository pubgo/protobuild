# protobuf.yaml / protobuf.plugin.yaml 操作要点

面向智能体的最小可用配置知识；完整人类向示例与讲解见仓库 `docs/EXAMPLES.md`、`docs/MULTI_SOURCE_DEPS.md`。

## protobuf.yaml 顶层字段

| 字段 | 含义 | 要点 |
|------|------|------|
| `root` | proto 源目录列表 | 相对项目根；`gen`/`format` 缺省作用范围 |
| `includes` / `excludes` | 编译期 import 搜索路径 / 排除 | `includes` 常与 `root` 相同，再叠加 vendor |
| `vendor` | 依赖存放目录 | 默认 `.proto`；vendor 产物目录，**禁止手改** |
| `deps` | 依赖列表 | 见下方多源类型 |
| `plugins` | protoc 插件配置 | 见下方插件字段 |
| `base` | 生成代码公共参数 | `out`（输出根）、`paths`（如 `source_relative`）、`module`（Go module 前缀） |
| `installers` | 自动安装的插件名 | `protobuild install` 读取 |
| `linter` | lint 规则配置 | `rules.enabled_rules` / `rules.disabled_rules`（AIP 规则名，如 `core::0131::http-method`；`disabled_rules: [all]` 表示只跑显式启用项）、`format_type` |
| `checksum` | 配置指纹（vendor 记录用） | 由 `vendor` 维护，**不要手填** |

## deps 依赖源（六种）

`source` 取值 `git` / `gomod` / `http` / `s3` / `gcs` / `local`；缺省自动探测，无法识别时按 `gomod` 处理。

```yaml
deps:
  - name: retag                          # git 源：取最新 semver tag
    url: github.com/pubgo/protoc-gen-retag
    path: proto/retag                    # 源仓库内的子目录
    version: v0.0.5                      # 固定版本；缺省跟随最新
  - name: common-proto
    source: gomod                        # go module：go list -m -versions
    url: example.com/org/common-proto
  - name: models
    source: s3                           # http/s3/gcs 同理：url 带 scheme，如 s3://bucket/path.tar.gz
    url: s3://company-artifacts/protos/models-v1.2.0.tar.gz
  - name: google/protobuf
    source: local                        # local：url 为本地路径
    url: /usr/local/include/google/protobuf
```

要点：

- `protobuild update` 只能自动发现 `git`/`gomod` 的新版本；`http/s3/gcs/local` 会被跳过（升级需手改 `version`/`url`）。
- 依赖同步落地到 `vendor` 目录并记录于 manifest；`.proto/` 内容异常用 `vendor -f` 重建。

## plugins 插件配置

```yaml
plugins:
  - name: go                       # 需 PATH 中存在 protoc-gen-go（doctor 检查 / install 安装）
    out: pkg/api                   # 该插件输出目录（覆盖 base.out）
    opt:
      - paths=source_relative      # 传给插件的参数
```

常用字段：`out`、`opt`/`opts`/`exclude_opts`（参数）、`shell`/`docker`/`remote`（以 shell 脚本、docker 或远程方式调用插件）、`skip_base`/`skip_run`。

## protobuf.plugin.yaml：目录级覆盖

在**子目录**放置 `protobuf.plugin.yaml`，其中 `plugins` 配置仅覆盖作用于该子树下的 proto，不影响根配置。适合"某目录用 gateway、其余只用 go"这类差异化生成。

## 与 buf.yaml 的关系

buf（LSP / format / breaking 引擎）只认 `buf.yaml`，其 import 相对 module root 解析。改 `protobuf.yaml` 的 `root`/`vendor` 后，`vendor` / `init` 会自动同步生成 buf.yaml；需要手动重生成时跑 `protobuild buf`，手写的 buf.yaml 默认不覆盖，接管需 `protobuild buf -f`。
