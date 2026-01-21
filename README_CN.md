# protobuild

[![Go Report Card](https://goreportcard.com/badge/github.com/pubgo/protobuild)](https://goreportcard.com/report/github.com/pubgo/protobuild)
[![License](https://img.shields.io/github/license/pubgo/protobuild)](LICENSE)

> 一个强大的 Protocol Buffers 构建和管理工具

[English](./README.md)

## 特性

- 🚀 **统一构建** - 一条命令编译所有 proto 文件
- 📦 **多源依赖** - 支持 Go 模块、Git、HTTP、S3、GCS 和本地路径
- 🔌 **插件支持** - 灵活的 protoc 插件配置
- 🔍 **代码检查** - 内置基于 AIP 规则的 proto 文件检查
- 📝 **格式化** - 自动格式化 proto 文件
- ⚙️ **配置驱动** - 基于 YAML 的项目配置
- 📊 **进度显示** - 可视化进度条和详细错误信息
- 🗑️ **缓存管理** - 清理和管理依赖缓存

## 安装

```bash
go install github.com/pubgo/protobuild@latest
```

## 快速开始

1. 在项目根目录创建 `protobuf.yaml` 配置文件：

```yaml
vendor: .proto
root:
  - proto
includes:
  - proto
deps:
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf
plugins:
  - name: go
    out: pkg
    opt:
      - paths=source_relative
```

2. 同步依赖：

```bash
protobuild vendor
```

3. 生成代码：

```bash
protobuild gen
```

## 命令说明

| 命令 | 说明 |
|------|------|
| `gen` | 编译 protobuf 文件 |
| `vendor` | 同步 proto 依赖到 vendor 目录 |
| `vendor -u` | 强制重新下载所有依赖（忽略缓存）|
| `deps` | 显示依赖列表及状态 |
| `install` | 安装 protoc 插件 |
| `lint` | 使用 AIP 规则检查 proto 文件 |
| `format` | 格式化 proto 文件 |
| `clean` | 清理依赖缓存 |
| `clean --dry-run` | 预览将被清理的内容 |
| `version` | 显示版本信息 |

## 配置说明

### 配置文件结构

```yaml
# 校验和，用于追踪变更（自动生成）
checksum: ""

# proto 依赖的 vendor 目录
vendor: .proto

# 基础插件配置（应用于所有插件）
base:
  out: pkg
  paths: source_relative
  module: github.com/your/module

# proto 源文件目录
root:
  - proto
  - api

# protoc 的 include 路径
includes:
  - proto
  - .proto

# 排除的路径
excludes:
  - proto/internal

# proto 依赖配置
deps:
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf
    version: v21.0
    optional: false

# protoc 插件配置
plugins:
  - name: go
    out: pkg
    opt:
      - paths=source_relative
  - name: go-grpc
    out: pkg
    opt:
      - paths=source_relative

# 插件安装器（go install）
installers:
  - google.golang.org/protobuf/cmd/protoc-gen-go@latest
  - google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 检查器配置
linter:
  rules:
    enabled_rules:
      - core::0131::http-method
    disabled_rules:
      - all
  format_type: yaml
```

### 插件配置

每个插件支持以下选项：

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 插件名称（用作 protoc-gen-{name}）|
| `path` | string | 自定义插件二进制路径 |
| `out` | string | 输出目录 |
| `opt` | string/list | 插件选项 |
| `shell` | string | 通过 shell 命令运行 |
| `docker` | string | 通过 Docker 容器运行 |
| `skip_base` | bool | 跳过基础配置 |
| `skip_run` | bool | 跳过此插件 |
| `exclude_opts` | list | 排除的选项 |

### 依赖配置

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | vendor 目录中的本地名称/路径 |
| `url` | string | 源 URL（Go 模块、Git URL、HTTP 归档、S3、GCS 或本地路径）|
| `path` | string | 源内的子目录 |
| `version` | string | 指定版本（用于 Go 模块）|
| `ref` | string | Git 引用（分支、标签、提交）用于 Git 源 |
| `source` | string | 源类型：`gomod`、`git`、`http`、`s3`、`gcs`、`local`（未指定时自动检测）|
| `optional` | bool | 找不到时跳过 |

#### 支持的依赖源

```yaml
deps:
  # Go 模块（默认）
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  # Git 仓库
  - name: googleapis
    url: https://github.com/googleapis/googleapis.git
    ref: master

  # HTTP 归档
  - name: envoy
    url: https://github.com/envoyproxy/envoy/archive/v1.28.0.tar.gz
    path: api

  # 本地路径
  - name: local-protos
    url: ./third_party/protos

  # S3 存储桶
  - name: internal-protos
    url: s3://my-bucket/protos.tar.gz

  # GCS 存储桶
  - name: shared-protos
    url: gs://my-bucket/protos.tar.gz
```

## 使用示例

### 使用自定义配置文件

```bash
protobuild -c protobuf.custom.yaml gen
```

### 检查 Proto 文件

```bash
protobuild lint
protobuild lint --list-rules  # 显示可用规则
protobuild lint --debug       # 调试模式
```

### 格式化 Proto 文件

```bash
protobuild format
```

### 强制更新 Vendor

```bash
protobuild vendor -f      # 强制更新，即使没有检测到变更
protobuild vendor -u      # 重新下载所有依赖（忽略缓存）
```

### 显示依赖状态

```bash
protobuild deps
```

输出示例：
```
📦 Dependencies:

  NAME                                SOURCE     VERSION      STATUS
  ----                                ------     -------      ------
  google/protobuf                     Go Module  v21.0        🟢 cached
  googleapis                          Git        master       ⚪ not cached

  Total: 2 dependencies
```

### 清理依赖缓存

```bash
protobuild clean           # 清理所有缓存的依赖
protobuild clean --dry-run # 预览将被清理的内容
```

### 安装插件

```bash
protobuild install
protobuild install -f  # 强制重新安装
```

## 目录级配置

你可以在任何 proto 目录中放置 `protobuf.plugin.yaml` 文件，以覆盖该目录及其子目录的根配置。

```yaml
# proto/api/protobuf.plugin.yaml
plugins:
  - name: go
    out: pkg/api
    opt:
      - paths=source_relative
```

## 支持的 Protoc 插件

- `google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest`
- `github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest`
- `github.com/bufbuild/protoc-gen-validate/cmd/protoc-gen-validate@latest`
- 以及更多...

## 错误处理

当依赖解析失败时，protobuild 会提供详细的错误信息和建议：

```
❌ Failed to download dependency: google/protobuf
   Source:  Git
   URL:     git::https://github.com/protocolbuffers/protobuf.git?ref=v99.0
   Ref:     v99.0
   Error:   reference not found

💡 Suggestions:
   • 检查仓库 URL 是否正确且可访问
   • 验证 ref（标签/分支/提交）是否存在
   • 确保您有正确的身份验证（SSH 密钥或令牌）
```

## 缓存位置

依赖缓存在：
- **macOS/Linux**: `~/.cache/protobuild/deps/`
- **Go 模块**: 标准 Go 模块缓存 (`$GOPATH/pkg/mod`)

## 文档

- [配置示例](./docs/EXAMPLES.md) - 各种使用场景的详细配置示例
- [多源依赖设计](./docs/MULTI_SOURCE_DEPS.md) - 多源依赖解析设计文档
- [设计文档](./docs/DESIGN_CN.md) - 架构和设计文档

## 项目架构

```
protobuild
├── cmd/
│   ├── protobuild/     # 主要 CLI 命令
│   ├── format/         # Proto 文件格式化
│   ├── formatcmd/      # 格式化命令包装器
│   └── linters/        # AIP 检查规则
└── internal/
    ├── depresolver/    # 多源依赖解析器
    ├── modutil/        # Go 模块工具
    ├── plugin/         # 插件管理
    ├── protoutil/      # Protobuf 工具
    ├── shutil/         # Shell 工具
    └── template/       # 模板工具
```

