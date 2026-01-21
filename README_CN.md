# protobuild

[![Go Report Card](https://goreportcard.com/badge/github.com/pubgo/protobuild)](https://goreportcard.com/report/github.com/pubgo/protobuild)
[![License](https://img.shields.io/github/license/pubgo/protobuild)](LICENSE)

> 一个强大的 Protocol Buffers 构建和管理工具

[English](./README.md)

## 特性

- 🚀 **统一构建** - 一条命令编译所有 proto 文件
- 📦 **依赖管理** - 自动管理 proto 依赖的 vendor
- 🔌 **插件支持** - 灵活的 protoc 插件配置
- 🔍 **代码检查** - 内置基于 AIP 规则的 proto 文件检查
- 📝 **格式化** - 自动格式化 proto 文件
- ⚙️ **配置驱动** - 基于 YAML 的项目配置

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
| `install` | 安装 protoc 插件 |
| `lint` | 使用 AIP 规则检查 proto 文件 |
| `format` | 格式化 proto 文件 |
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
| `url` | string | Go 模块路径或本地路径 |
| `path` | string | 模块内的子目录 |
| `version` | string | 指定版本（可选）|
| `optional` | bool | 找不到时跳过 |

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
protobuild vendor -f
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

## 许可证

[MIT License](LICENSE)

## 贡献

欢迎贡献！请随时提交 Pull Request。
