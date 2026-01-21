# protobuild 设计文档

## 概述

protobuild 是一个命令行工具，旨在简化 Protocol Buffers 的开发工作流程。它提供统一的配置管理、依赖处理、代码生成、代码检查和格式化功能。

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                      protobuild CLI                         │
├─────────────────────────────────────────────────────────────┤
│  命令: gen | vendor | install | lint | format | version     │
├─────────────────────────────────────────────────────────────┤
│                        配置层                                │
│              (protobuf.yaml / protobuf.plugin.yaml)          │
├─────────────────────────────────────────────────────────────┤
│    依赖        │     插件       │    检查器   │   格式化器  │
│    管理器      │     管理器     │    引擎     │    引擎     │
├─────────────────────────────────────────────────────────────┤
│                      protoc / Go Modules                     │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. 配置系统

配置系统支持具有继承性的层级配置：

- **根配置** (`protobuf.yaml`)：项目级配置
- **目录配置** (`protobuf.plugin.yaml`)：目录级覆盖配置

配置加载流程：

```
1. 加载根配置 protobuf.yaml
2. 遍历 proto 目录
3. 检查每个目录中的 protobuf.plugin.yaml
4. 合并配置，支持继承
5. 应用基础插件设置
```

### 2. 依赖管理器

负责管理 proto 文件依赖：

**功能特性：**
- 通过 `go mod graph` 自动解析版本
- 集成 Go 模块缓存（`$GOPATH/pkg/mod`）
- 支持本地路径
- 基于校验和的变更检测
- 可选依赖支持

**工作流程：**
```
1. 从配置解析依赖
2. 从 go.mod 或指定配置解析版本
3. 下载/定位 proto 文件
4. 复制到 vendor 目录
5. 更新校验和
```

### 3. 插件管理器

管理 protoc 插件的执行：

**插件类型：**
- 标准 protoc 插件 (protoc-gen-*)
- Shell 脚本插件
- Docker 容器插件

**执行流程：**
```
1. 加载插件配置
2. 应用基础设置
3. 构建带选项的 protoc 命令
4. 对每个 proto 目录执行
5. 特殊处理 retag 插件（后处理）
```

### 4. 检查器引擎

集成 [api-linter](https://github.com/googleapis/api-linter) 进行 proto 文件验证：

**功能特性：**
- AIP 规则执行
- 自定义规则启用/禁用
- 多种输出格式（YAML、JSON、GitHub Actions）
- 支持注释禁用

### 5. 格式化引擎

使用以下工具格式化 proto 文件：
- [protocompile](https://github.com/bufbuild/protocompile) 解析器
- 自定义格式化规则

## 数据流

### 生成流程

```
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│ protobuf.yaml│───▶│    配置     │───▶│   遍历目录   │
└──────────────┘    │    解析器   │    │   (*.proto)  │
                    └─────────────┘    └──────────────┘
                                              │
                                              ▼
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│    生成的    │◀───│   protoc    │◀───│   构建命令   │
│     代码     │    │    执行     │    │   带选项     │
└──────────────┘    └─────────────┘    └──────────────┘
```

### Vendor 流程

```
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│    deps      │───▶│    解析     │───▶│    下载      │
│    配置      │    │    版本     │    │   /定位      │
└──────────────┘    └─────────────┘    └──────────────┘
                                              │
                                              ▼
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│    更新      │◀───│    复制     │◀───│    过滤      │
│   校验和     │    │  到 Vendor  │    │   .proto     │
└──────────────┘    └─────────────┘    └──────────────┘
```

## 配置模式

### 主配置 (protobuf.yaml)

```yaml
# 自动生成的校验和，用于变更检测
checksum: string

# vendor 目录路径
vendor: string (默认: .proto)

# 基础插件配置
base:
  out: string      # 默认输出目录
  paths: string    # paths 选项 (source_relative|import)
  module: string   # 模块前缀

# proto 源文件目录
root: []string

# protoc -I 的 include 路径
includes: []string

# 排除处理的路径
excludes: []string

# proto 依赖
deps:
  - name: string     # vendor 中的本地路径
    url: string      # 模块路径或本地路径
    path: string     # 模块内子目录
    version: string  # 指定版本
    optional: bool   # 找不到时跳过

# 插件配置
plugins:
  - name: string         # 插件名称
    path: string         # 自定义二进制路径
    out: string          # 输出目录
    opt: string|[]string # 插件选项
    shell: string        # Shell 命令
    docker: string       # Docker 镜像
    skip_base: bool      # 跳过基础配置
    skip_run: bool       # 跳过执行
    exclude_opts: []string # 排除的选项

# 插件安装器
installers: []string

# 检查器配置
linter:
  rules:
    enabled_rules: []string
    disabled_rules: []string
  format_type: string
  ignore_comment_disables_flag: bool
```

## 关键设计决策

### 1. YAML 配置

**理由：** YAML 提供人类可读的格式，支持注释，便于文档化和维护配置。

### 2. Go 模块集成

**理由：** 利用现有的 Go 工具链进行依赖解析，避免需要单独的依赖管理系统。

### 3. 命令中间件模式

**理由：** 中间件模式（通过 redant）允许清晰地分离关注点：
- 配置解析中间件
- 错误处理
- 恢复机制

### 4. 基于校验和的变更检测

**理由：** 通过 SHA1 校验和跟踪配置变更，避免不必要的 vendor 更新。

### 5. 层级配置

**理由：** 允许目录特定的覆盖配置，同时保持项目范围的默认值，适用于 monorepo 结构。

## 错误处理

项目使用一致的错误处理方法：

1. **断言** (`assert.Must`, `assert.Exit`)：用于不可恢复的错误
2. **恢复** (`recovery.Exit`, `recovery.Err`)：用于 panic 恢复
3. **错误包装** (`errors.WrapTag`)：用于提供上下文丰富的错误消息

## 扩展点

### 自定义插件

支持三种类型的自定义插件：

1. **二进制插件**：标准 protoc 插件
2. **Shell 插件**：通过 shell 命令执行
3. **Docker 插件**：通过 Docker 容器执行

### 自定义检查规则

通过检查器配置：
- 启用特定 AIP 规则
- 全局或按文件禁用规则
- 自定义输出格式

## 依赖项

关键外部依赖：

| 包 | 用途 |
|---|------|
| `github.com/pubgo/redant` | CLI 框架 |
| `github.com/googleapis/api-linter` | Proto 检查 |
| `github.com/bufbuild/protocompile` | Proto 解析/格式化 |
| `github.com/samber/lo` | 工具函数 |
| `gopkg.in/yaml.v3` | YAML 解析 |

## 未来考虑

1. **远程插件支持**：通过远程服务执行插件
2. **并行执行**：并发 proto 编译
3. **监视模式**：文件监视以自动重新生成
4. **插件缓存**：缓存插件二进制文件以加快执行
5. **Proto 注册表**：与 Buf Schema Registry 集成
