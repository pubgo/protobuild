# 故障排查

## doctor：环境与配置体检

`protobuild doctor` 逐项检查并汇总：

| 检查项 | 内容 |
|--------|------|
| protoc | 编译器是否存在 |
| protoc-gen-go / protoc-gen-go-grpc | 常用 Go 插件 |
| buf | buf CLI（format/lsp/breaking 引擎） |
| api-linter | AIP lint 引擎 |
| go | Go 工具链 |
| config | `protobuf.yaml` 可解析、必要字段完整 |
| vendor | 依赖是否已 vendor 到 `.proto/` |
| proto compile | 实际编译一遍 proto 验证 import 链 |

`protobuild doctor --fix` 在发现问题后自动执行：`go install` 缺失的 protoc-gen-go / protoc-gen-go-grpc。其余不代做：protoc、buf、api-linter 缺失只提示；`.proto/` 未 vendor 时也只打印"请手动运行 protobuild vendor"，**不会代跑 vendor**。

## verify：vendor 产物完整性校验

`protobuild verify` 按 `.proto/manifest.yaml` 中逐文件 `sha256` 与 `size` 校验 `.proto/` 内容与 vendor 时一致（manifest 还记录 `config_checksum` 与来源）。适合 CI 门禁：失败说明 `.proto/` 被手改或损坏。

## 症状 → 修法

| 症状 | 原因 | 修法 |
|------|------|------|
| `gen` 报错找不到 import | 依赖未同步 | 先 `protobuild vendor` 再 `gen` |
| `gen` 报插件不存在/不可执行 | protoc 插件缺失 | `protobuild install`（或 `doctor --fix`） |
| `.proto/` 内容异常（package 被文本替换弄坏等） | vendor 产物被误改 | `protobuild vendor -f` 整体重建，勿手工修补 |
| `verify` 失败 | vendor 产物被改动或缓存不一致 | `vendor -f` 重建后再 `verify` |
| buf / LSP 在 IDE 报红但编译正常 | `buf.yaml` 与 `protobuf.yaml` 的 root/vendor 不同步；或 `google/protobuf/*` 靠 buf 内置 WKT 兜底造成假象 | `protobuild buf` 重新生成（手写 buf.yaml 被接管需 `-f`） |
| `protobuild breaking` 大量误报 | 对比输入包含依赖升级 | 用缺省自动物化 base（tracked 文件 + 当前 vendor），不要手传含旧 vendor 的 `--against` |
| `version` 显示与 tag 不符 | 版本号来自嵌入的 `.version/VERSION`，非 git tag | 属正常：发版流程先改 VERSION 再打 tag，本地 `go run` 永远显示嵌入值 |
| 升级依赖后 vendor 未生效 | http/s3/gcs/local 源不支持自动发现版本 | `update` 会跳过这些源，需手改 `version`/`url` 后 `vendor -u` 强制重下 |
