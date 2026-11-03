# protobuild 命令速查

> 本表由守卫测试 `TestCommandsDocumentedInSkillRefs` 保证命令名覆盖完整；新增子命令时同步本表。
> 所有命令均可用全局 `-c/--conf <path>` 指定配置文件（默认 `protobuf.yaml`）。
> "写盘"列标注会修改项目文件或外部状态的命令；执行前须告知用户。

| 命令 | 用途 | 关键 flag | 写盘 |
|------|------|-----------|------|
| `init` | 初始化 protobuf 项目（生成 protobuf.yaml + buf.yaml + 示例 greeter.proto） | `-f/--force` 覆盖已有配置；`-t/--template` 模板（`basic`、`grpc-gateway`、`minimal`，区分大小写） | 是（创建配置与示例） |
| `doctor` | 检查开发环境（protoc、插件、配置） | `--fix` 尝试自动修复 | `--fix` 时可能修复环境 |
| `vendor` | 同步 protobuf 依赖到 `.proto/` 目录 | `-f/--force` 强制重建；`-u/--update` 忽略缓存强制重新下载 | 是（写 `.proto/`） |
| `update`（可带 `[name...]`） | 升级依赖到最新发布版本并重新 vendor；git/gomod 源可自动发现版本，http/s3/gcs/local 源会被跳过 | `--dry-run` 只显示可升级项 | 是（改 protobuf.yaml 并重新 vendor） |
| `gen` | 编译 protobuf 文件，按配置生成代码 | —（读 protobuf.yaml 与 protobuf.plugin.yaml） | 是（写生成代码） |
| `verify` | 校验 `.proto/` 内容与 manifest（per-file sha256）一致，CI 门禁 | — | 否 |
| `buf` | 生成 `buf.yaml`，供 buf CLI / buf LSP 解析 import | `-f/--force` 覆盖已存在的 buf.yaml | 是（写 buf.yaml） |
| `breaking` | proto 兼容性检查（包装 `buf breaking`，自动排除 vendor）；默认对比自动物化的 base 分支（origin/main → master → main） | `--against` buf 输入（如 `.git#branch=main`、`.git#tag=v1.0.0`）；`--error-format`（text/json/msvs/junit/github-actions/gitlab-code-quality） | 否 |
| `lint` | AIP lint 规则检查（api-linter + protocompile） | `--debug`；`--list-rules` 列出规则后退出 | 否 |
| `format` | 格式化 proto 文件（默认 buf 引擎），缺省作用于配置 root 目录，可显式给路径 | `-w/--write` 写回；`-d/--diff` 显示差异；`--exit-code` 需要格式化时非零退出（CI）；`--builtin` / `--clang-format` 换引擎；`--clang-style`（google 等） | `-w` 时是 |
| `deps` | 显示依赖列表及状态 | — | 否 |
| `clean` | 清理依赖缓存（默认全部） | `--dry-run` 预览；`--max-age`（如 `30d`）；`--max-size`（如 `500MB`） | 是（删缓存，不碰项目文件） |
| `install` | 安装 protoc 插件 | `-f/--force` 强制更新 | 是（写插件二进制） |
| `web` | 启动可视化配置管理界面 | `-p/--port`（默认 8080） | 否 |
| `version` | 显示版本信息 | — | 否 |
| `upgrade` | 自升级管理（子命令 `list`） | — | 是（替换自身二进制） |
