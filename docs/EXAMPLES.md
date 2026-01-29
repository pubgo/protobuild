# protobuild 配置示例 / Configuration Examples

本文档提供了各种使用场景的详细配置示例。

---

## 目录 / Table of Contents

- [基础配置](#基础配置--basic-configuration)
- [依赖配置示例](#依赖配置示例--dependency-examples)
- [插件配置示例](#插件配置示例--plugin-examples)
- [完整项目示例](#完整项目示例--complete-project-examples)
  - [示例 1: 标准 Go 微服务](#示例-1-标准-go-微服务)
  - [示例 2: gRPC Gateway 项目](#示例-2-grpc-gateway-项目)
  - [示例 3: Python 项目](#示例-3-python-项目)
  - [示例 4: 多语言 Monorepo](#示例-4-多语言-monorepo)
  - [示例 5: 企业内部项目](#示例-5-企业内部项目-私有依赖)
  - [示例 6: 使用 Validate 验证](#示例-6-使用-validate-验证)
  - [示例 7: 完整生产项目 (推荐)](#示例-7-完整生产项目-推荐参考)
    - [示例 8: 协议仓 + SDK 仓自动发布](#示例-8-协议仓--sdk-仓自动发布)
- [高级用法](#高级用法--advanced-usage)

---

## 基础配置 / Basic Configuration

### 最小配置 (Minimal)

```yaml
# protobuf.yaml
vendor: .proto
root:
  - proto
plugins:
  - name: go
    out: gen
```

### 标准 Go 项目配置

```yaml
# protobuf.yaml
vendor: .proto

root:
  - proto
  - api

includes:
  - proto
  - api
  - .proto

excludes:
  - proto/internal
  - proto/testdata

deps:
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

plugins:
  - name: go
    out: gen
    opt:
      - paths=source_relative
```

---

## 依赖配置示例 / Dependency Examples

### 1. Go Module 源 (默认)

适用于已发布到 Go 模块的 proto 文件。

```yaml
deps:
  # 标准 Google Protobuf 定义
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf
    version: v25.0  # 可选，指定版本

  # Google API 定义
  - name: google/api
    url: github.com/googleapis/googleapis
    path: google/api
    
  # gRPC Gateway 注解
  - name: protoc-gen-openapiv2/options
    url: github.com/grpc-ecosystem/grpc-gateway/v2
    path: protoc-gen-openapiv2/options
    
  # Validate 规则
  - name: validate
    url: github.com/bufbuild/protovalidate
    path: proto/protovalidate
```

### 2. Git 源

适用于直接从 Git 仓库拉取。

```yaml
deps:
  # 使用标签版本
  - name: googleapis
    source: git
    url: https://github.com/googleapis/googleapis.git
    ref: v0.168.0
    path: google

  # 使用分支
  - name: envoy-api
    source: git
    url: https://github.com/envoyproxy/envoy.git
    ref: main
    path: api/envoy

  # 使用 commit hash
  - name: grpc-proto
    source: git
    url: https://github.com/grpc/grpc-proto.git
    ref: abc123def456

  # 私有仓库 (SSH)
  - name: internal-protos
    source: git
    url: git@github.com:mycompany/internal-protos.git
    ref: v1.0.0

  # 使用 token 访问私有仓库
  - name: private-api
    source: git
    url: https://oauth2:${GITHUB_TOKEN}@github.com/mycompany/api-protos.git
    ref: main
```

### 3. HTTP 源

适用于从 HTTP/HTTPS URL 下载归档文件。

```yaml
deps:
  # GitHub Release 归档
  - name: protobuf
    source: http
    url: https://github.com/protocolbuffers/protobuf/releases/download/v25.0/protobuf-25.0.tar.gz
    path: src/google/protobuf

  # GitLab 归档
  - name: gitlab-protos
    source: http
    url: https://gitlab.com/myorg/protos/-/archive/main/protos-main.tar.gz
    
  # ZIP 格式
  - name: external-api
    source: http
    url: https://example.com/api/v1/proto.zip
    
  # 需要认证的 HTTP 源
  - name: auth-protos
    source: http
    url: https://artifacts.mycompany.com/protos/v1.0.0.tar.gz
    # 使用环境变量: HTTP_AUTH=user:password 或配置 .netrc
```

### 4. S3 源

适用于从 AWS S3 存储桶获取依赖。

```yaml
deps:
  # 公开的 S3 存储桶
  - name: public-protos
    source: s3
    url: s3://public-protos-bucket/api/v1.tar.gz

  # 带路径前缀
  - name: team-protos
    source: s3
    url: s3://company-protos/team-a/protos-v2.0.0.tar.gz

  # 指定区域 (通过环境变量)
  # AWS_REGION=us-west-2 protobuild vendor
  - name: regional-protos
    source: s3
    url: s3://regional-bucket/protos.tar.gz
```

**S3 认证配置:**
```bash
# 方式 1: AWS 配置文件
# ~/.aws/credentials

# 方式 2: 环境变量
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_REGION=us-west-2
```

### 5. GCS 源

适用于从 Google Cloud Storage 获取依赖。

```yaml
deps:
  # GCS 存储桶
  - name: gcp-protos
    source: gcs
    url: gs://my-company-protos/api/v1.0.0.tar.gz

  # 带路径
  - name: shared-protos
    source: gcs
    url: gcs://shared-bucket/protos/common.tar.gz
```

**GCS 认证配置:**
```bash
# 方式 1: 服务账户
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# 方式 2: gcloud CLI 登录
gcloud auth application-default login
```

### 6. 本地路径源

适用于使用本地文件系统中的 proto 文件。

```yaml
deps:
  # 相对路径
  - name: local-common
    source: local
    url: ./third_party/protos

  # 绝对路径
  - name: system-protos
    source: local
    url: /usr/local/include/google/protobuf

  # 上级目录
  - name: shared-protos
    source: local
    url: ../shared/protos
    
  # Monorepo 中的其他模块
  - name: other-service
    source: local
    url: ../other-service/proto
```

### 7. 可选依赖

当依赖不是必需的时候使用。

```yaml
deps:
  # 必需依赖
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  # 可选依赖 - 找不到时不会报错
  - name: optional-annotations
    url: github.com/some/optional-lib
    path: annotations
    optional: true
```

---

## 插件配置示例 / Plugin Examples

### 1. Go 插件

```yaml
plugins:
  # 基础 Go 生成
  - name: go
    out: gen/go
    opt:
      - paths=source_relative
      
  # gRPC 服务生成
  - name: go-grpc
    out: gen/go
    opt:
      - paths=source_relative
      - require_unimplemented_servers=false
```

### 2. gRPC Gateway 插件

```yaml
plugins:
  - name: go
    out: gen/go
    opt: paths=source_relative

  - name: go-grpc
    out: gen/go
    opt: paths=source_relative

  # HTTP 网关代码
  - name: grpc-gateway
    out: gen/go
    opt:
      - paths=source_relative
      - generate_unbound_methods=true

  # OpenAPI 文档
  - name: openapiv2
    out: docs/swagger
    opt:
      - allow_merge=true
      - merge_file_name=api
```

### 3. 多语言生成

```yaml
plugins:
  # Go
  - name: go
    out: gen/go
    opt: paths=source_relative

  # TypeScript (使用 ts-proto)
  - name: ts_proto
    path: ./node_modules/.bin/protoc-gen-ts_proto
    out: gen/ts
    opt:
      - esModuleInterop=true
      - outputServices=grpc-js

  # Python
  - name: python
    out: gen/python
    
  # Python gRPC
  - name: grpc_python
    out: gen/python
```

### 4. 使用 Docker 运行插件

```yaml
plugins:
  # 使用 Docker 容器中的插件
  - name: doc
    docker: pseudomuto/protoc-gen-doc:latest
    out: docs
    opt:
      - html,index.html
```

### 5. 使用 Shell 命令

```yaml
plugins:
  # 自定义 shell 处理
  - name: custom
    shell: |
      cat > ${OUTPUT}/proto_list.txt
    out: gen
```

### 6. 基础配置继承

```yaml
# 全局基础配置
base:
  out: gen
  paths: source_relative
  module: github.com/mycompany/myproject

plugins:
  # 继承 base 配置
  - name: go
    # out 和 paths 自动继承
    
  - name: go-grpc
    opt:
      - require_unimplemented_servers=false
    
  # 跳过基础配置
  - name: doc
    skip_base: true
    out: docs
```

### 7. 条件性跳过插件

```yaml
plugins:
  - name: go
    out: gen/go
    opt: paths=source_relative

  # 可以通过 skip_run 临时禁用
  - name: go-grpc
    out: gen/go
    opt: paths=source_relative
    skip_run: false  # 改为 true 可跳过此插件
```

---

## 完整项目示例 / Complete Project Examples

### 示例 1: 标准 Go 微服务

```yaml
# protobuf.yaml - Go 微服务项目

checksum: ""
vendor: .proto

root:
  - api

includes:
  - api
  - .proto

deps:
  # Google 公共定义
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  - name: google/api
    url: github.com/googleapis/googleapis
    path: google/api

  # 验证库
  - name: validate
    url: github.com/bufbuild/protovalidate
    path: proto/protovalidate

plugins:
  - name: go
    out: pkg/pb
    opt:
      - paths=source_relative

  - name: go-grpc
    out: pkg/pb
    opt:
      - paths=source_relative
      - require_unimplemented_servers=false

installers:
  - google.golang.org/protobuf/cmd/protoc-gen-go@latest
  - google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 示例 2: gRPC Gateway 项目

```yaml
# protobuf.yaml - gRPC Gateway 项目

vendor: .proto

root:
  - proto

includes:
  - proto
  - .proto

deps:
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  - name: google/api
    url: github.com/googleapis/googleapis
    path: google/api

  - name: protoc-gen-openapiv2/options
    url: github.com/grpc-ecosystem/grpc-gateway/v2
    path: protoc-gen-openapiv2/options

plugins:
  - name: go
    out: gen/go
    opt: paths=source_relative

  - name: go-grpc
    out: gen/go
    opt: paths=source_relative

  - name: grpc-gateway
    out: gen/go
    opt:
      - paths=source_relative
      - generate_unbound_methods=true

  - name: openapiv2
    out: docs/swagger
    opt:
      - allow_merge=true
      - merge_file_name=api

installers:
  - google.golang.org/protobuf/cmd/protoc-gen-go@latest
  - google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  - github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
  - github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### 示例 3: Python 项目

```yaml
# protobuf.yaml - Python gRPC 项目

vendor: third_party/proto

root:
  - proto

includes:
  - proto
  - third_party/proto

deps:
  # 直接从 Git 获取，不依赖 Go 模块
  - name: google/protobuf
    source: git
    url: https://github.com/protocolbuffers/protobuf.git
    ref: v25.0
    path: src/google/protobuf

  - name: googleapis
    source: git
    url: https://github.com/googleapis/googleapis.git
    ref: master
    path: google

plugins:
  - name: python
    out: gen/python

  - name: grpc_python
    out: gen/python
```

### 示例 4: 多语言 Monorepo

```yaml
# protobuf.yaml - Monorepo 多语言支持

vendor: .proto

root:
  - proto

includes:
  - proto
  - .proto

deps:
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  - name: google/api
    url: github.com/googleapis/googleapis
    path: google/api

# Go 生成
plugins:
  - name: go
    out: services/go/pb
    opt: paths=source_relative

  - name: go-grpc
    out: services/go/pb
    opt: paths=source_relative

  # TypeScript 生成
  - name: ts_proto
    path: ./node_modules/.bin/protoc-gen-ts_proto
    out: services/web/src/proto
    opt:
      - esModuleInterop=true
      - outputServices=grpc-js
      - useDate=true

  # Python 生成
  - name: python
    out: services/python/src/proto

  - name: grpc_python
    out: services/python/src/proto
```

### 示例 5: 企业内部项目 (私有依赖)

```yaml
# protobuf.yaml - 企业内部项目

vendor: .proto

root:
  - api

includes:
  - api
  - .proto

deps:
  # 公共依赖
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  # 内部 Git 仓库 (SSH)
  - name: internal/common
    source: git
    url: git@gitlab.company.com:platform/proto-common.git
    ref: v2.0.0
    path: common

  # 内部 S3 存储
  - name: internal/models
    source: s3
    url: s3://company-artifacts/protos/models-v1.2.0.tar.gz

  # 团队共享 (本地 Monorepo)
  - name: shared
    source: local
    url: ../shared-proto

plugins:
  - name: go
    out: pkg/pb
    opt: paths=source_relative

  - name: go-grpc
    out: pkg/pb
    opt: paths=source_relative

linter:
  rules:
    enabled_rules:
      - core::0131::http-method
      - core::0131::http-body
    disabled_rules:
      - all
  format_type: yaml
```

### 示例 6: 使用 Validate 验证

```yaml
# protobuf.yaml - 带验证规则的项目

vendor: .proto

root:
  - proto

includes:
  - proto
  - .proto

deps:
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  # protovalidate (新版)
  - name: buf/validate
    url: github.com/bufbuild/protovalidate
    path: proto/protovalidate/buf/validate

plugins:
  - name: go
    out: gen
    opt: paths=source_relative

  - name: go-grpc
    out: gen
    opt: paths=source_relative

installers:
  - google.golang.org/protobuf/cmd/protoc-gen-go@latest
  - google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 示例 7: 完整生产项目 (推荐参考)

这是一个真实的生产项目配置，展示了多种高级特性的组合使用：

```yaml
# protobuf.yaml - 完整生产项目配置示例

checksum: 37d087bae0164f496552680f76e933a473da9c31
vendor: proto-vendor

# 基础配置 - 所有插件继承这些设置
base:
  out: ./pkg/gen
  paths: import
  module: github.com/pubgo/catdogs/pkg/gen

root:
  - proto

includes:
  - proto

deps:
  # Google APIs (googleapis)
  - name: google
    url: github.com/googleapis/googleapis
    path: /google
    version: v0.0.0-20241227204702-9c586e8f14bc

  # 本地 protobuf (macOS Intel) - 可选
  - name: google/protobuf
    url: /usr/local/include/google/protobuf
    optional: true

  # 本地 protobuf (macOS ARM/Homebrew) - 可选
  - name: google/protobuf
    url: /opt/homebrew/include/google/protobuf
    optional: true

  # protoc-gen-validate 验证规则
  - name: validate
    url: github.com/envoyproxy/protoc-gen-validate
    path: /validate
    version: v1.2.1

  # OpenAPI v3 注解
  - name: openapiv3
    url: github.com/pubgo/protoc-gen-openapi
    path: /proto/openapiv3
    version: v0.7.9

  # 错误码定义
  - name: errorpb
    url: github.com/pubgo/funk/v2
    path: /proto/errorpb
    version: v2.0.0-beta.10

  # Lava 框架注解
  - name: lava
    url: github.com/pubgo/lava/v2
    path: /proto/lava
    version: v2.0.0-beta.3

plugins:
  # Go 基础生成 (继承 base 配置)
  - name: go

  # gRPC 服务生成
  - name: go-grpc
    opt:
      - require_unimplemented_servers=false

  # 错误码生成
  - name: go-errors2

  # CloudEvent 生成
  - name: go-cloudevent

  # 枚举增强生成
  - name: go-enum2

  # OpenAPI 文档生成 (独立输出目录)
  - name: openapi
    out: ./docs/swagger
    skip_base: true  # 跳过 base 配置

# 插件安装器
installers:
  - github.com/pubgo/protobuild@latest
  - google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  - github.com/envoyproxy/protoc-gen-validate/cmd/protoc-gen-validate-go@latest
  - github.com/pubgo/funk/v2/cmds/protoc-gen-go-errors@latest
  - github.com/pubgo/funk/v2/cmds/protoc-gen-go-enum@latest
  - storj.io/drpc/cmd/protoc-gen-go-drpc@latest
  - github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
  - connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
  - github.com/twitchtv/twirp/protoc-gen-twirp@latest
  - github.com/pubgo/lava/cmds/protoc-gen-lava@latest
  - github.com/pubgo/protoc-gen-openapi@latest
  - github.com/emicklei/proto-contrib/cmd/protofmt@latest

# Linter 配置
linter:
  rules:
    included_paths: []
    excluded_paths: []
    enabled_rules:
      - core::0131::http-method
      - core::0131::http-body
      - core::0235::plural-method-name
    disabled_rules:
      - all
  format_type: yaml
  ignore_comment_disables_flag: false
```

**配置要点说明：**

| 特性 | 说明 |
|------|------|

  ### 示例 8: 协议仓 + SDK 仓自动发布

  场景：所有 `.proto` 放在“协议仓”，生成的 SDK（多语言）放在“SDK 仓”。协议仓打 tag 后，CI 自动生成代码、提交到 SDK 仓并打同名 tag。

  #### 协议仓目录示例

  - `proto/`：协议源文件
  - `protobuf.yaml`：现有生成配置（复用 protobuild）
  - `scripts/generate.sh`：生成并推送 SDK 的脚本（见下）

  #### scripts/generate.sh（简化示例，保持分支对齐）

  ```bash
  #!/usr/bin/env bash
  set -euo pipefail

  PROTO_ROOT="proto"
  OUT_DIR="pkg"              # 生成物输出
  SDK_REPO_URL="${SDK_REPO_URL:-git@github.com:your-org/proto-sdk-repo.git}"
  SDK_REPO_DIR=".sdk-repo"
  SDK_SUBDIR="go"            # SDK 仓中的子目录（按语言分 go/js/python）
  SDK_BRANCH="${SDK_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}"  # 与协议仓当前分支保持一致

  # 安装工具（如已预装可跳过）
  command -v protoc >/dev/null || { echo "missing protoc"; exit 1; }
  command -v protoc-gen-go >/dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  command -v protoc-gen-go-grpc >/dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

  # 清理并生成
  rm -rf "${OUT_DIR}" && mkdir -p "${OUT_DIR}"
  go run . gen   # 复用 protobuild 的生成逻辑

  # 克隆 SDK 仓并同步生成物
  rm -rf "${SDK_REPO_DIR}"
  git clone "${SDK_REPO_URL}" "${SDK_REPO_DIR}"
  (
    cd "${SDK_REPO_DIR}"
    git checkout -B "${SDK_BRANCH}"  # 在 SDK 仓使用同名分支，避免覆盖他人分支
  )
  rsync -av --delete "${OUT_DIR}/" "${SDK_REPO_DIR}/${SDK_SUBDIR}/"

  # 提交并打 tag（TAG 由 CI 传入，与协议仓 tag 对齐）
  cd "${SDK_REPO_DIR}"
  if git status --porcelain | grep -q .; then
    COMMIT_MSG="chore: update proto SDK from $(git -C .. rev-parse --short HEAD)"
    git add "${SDK_SUBDIR}"
    git commit -m "${COMMIT_MSG}"
    if [ -n "${TAG:-}" ]; then git tag -f "${TAG}"; fi
    git push origin "${SDK_BRANCH}"
    if [ -n "${TAG:-}" ]; then git push origin "${TAG}" --force; fi
  else
    echo "No changes to commit."
  fi
  ```

  #### GitHub Actions（协议仓 .github/workflows/publish-sdk.yml）

  ```yaml
  name: Publish SDK

  on:
    push:
      tags:
        - 'v*'

  jobs:
    build-publish:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
          with:
            fetch-depth: 0

        - name: Setup Go
          uses: actions/setup-go@v5
          with:
            go-version: '1.21'

        - name: Set TAG and branch from ref
          run: |
            echo "TAG=${GITHUB_REF##*/}" >> $GITHUB_ENV
            echo "SDK_BRANCH=${GITHUB_REF_NAME}" >> $GITHUB_ENV

        - name: Publish SDK
          env:
            SDK_REPO_URL: git@github.com:your-org/proto-sdk-repo.git
            TAG: ${{ env.TAG }}
            SDK_BRANCH: ${{ env.SDK_BRANCH }}
          run: |
            chmod +x scripts/generate.sh
            scripts/generate.sh
          # 需要写 SDK 仓的 deploy key / PAT：
          # - uses: webfactory/ssh-agent@v0.9.0
          #   with:
          #     ssh-private-key: ${{ secrets.SSH_PRIVATE_KEY }}
  ```

  #### 实践要点

  - 协议仓 tag 与 SDK 仓 tag 一一对应，避免版本漂移。
  - 分支对齐：在 SDK 仓使用与协议仓相同的分支名（CI 传入 `SDK_BRANCH`），避免覆盖他人分支。
  - 多语言时在脚本中按语言追加生成命令和 `SDK_SUBDIR`，同一提交推送。
  - 固定生成工具版本（protoc/插件），避免无谓 diff。
  - 无变更不提交；有变更才生成 commit/tag。
| `base` | 全局插件配置，避免重复设置 `out`、`paths`、`module` |
| `optional: true` | 可选依赖，适用于不同系统的本地路径差异 |
| `skip_base: true` | 特定插件跳过全局配置，如 OpenAPI 单独输出到 docs 目录 |
| `version` 指定 | 锁定依赖版本，确保构建可重复性 |
| `installers` | 一键安装所有需要的 protoc 插件 |

---

## 高级用法 / Advanced Usage

### 目录级配置覆盖

在子目录中创建 `protobuf.plugin.yaml` 覆盖根配置：

```yaml
# proto/admin/protobuf.plugin.yaml
# 只对 proto/admin/ 目录下的文件生效

plugins:
  - name: go
    out: pkg/admin
    opt:
      - paths=source_relative

  # 只为管理 API 生成文档
  - name: doc
    out: docs/admin
    opt:
      - html,index.html
```

### 环境变量使用

```yaml
deps:
  # 使用环境变量
  - name: private-api
    source: git
    url: https://oauth2:${GITHUB_TOKEN}@github.com/company/api.git
    ref: ${API_VERSION:-main}  # 默认使用 main

plugins:
  - name: go
    out: ${OUTPUT_DIR:-gen}
    opt: paths=source_relative
```

### Linter 配置

```yaml
linter:
  rules:
    # 包含的路径
    included_paths:
      - proto/api
      
    # 排除的路径
    excluded_paths:
      - proto/internal
      - proto/third_party
      
    # 启用的规则
    enabled_rules:
      - core::0131::http-method
      - core::0131::http-body
      - core::0235::plural-method-name
      
    # 禁用的规则
    disabled_rules:
      - all  # 先禁用所有，再单独启用
      
  format_type: yaml  # 输出格式: yaml, json, text
  ignore_comment_disables_flag: false
```

### 多配置文件

```bash
# 开发环境
protobuild -c protobuf.dev.yaml gen

# 生产环境
protobuild -c protobuf.prod.yaml gen

# CI 环境
protobuild -c protobuf.ci.yaml gen
```

---

## 常见问题 / FAQ

### Q: 如何查看依赖状态？

```bash
protobuild deps
```

### Q: 如何强制重新下载依赖？

```bash
protobuild vendor -u
```

### Q: 如何清理缓存？

```bash
# 预览将被清理的内容
protobuild clean --dry-run

# 实际清理
protobuild clean
```

### Q: 如何使用私有 Git 仓库？

确保 SSH 密钥已配置，或使用 HTTPS + Token：

```yaml
deps:
  # SSH
  - name: private
    source: git
    url: git@github.com:company/repo.git
    ref: main

  # HTTPS + Token
  - name: private2
    source: git
    url: https://oauth2:${GITHUB_TOKEN}@github.com/company/repo.git
    ref: main
```

### Q: 错误提示 "reference not found" 怎么办？

检查 `ref` 字段的值是否正确：

```bash
# 查看远程仓库的所有标签
git ls-remote --tags https://github.com/user/repo.git

# 查看远程仓库的所有分支
git ls-remote --heads https://github.com/user/repo.git
```

---

## 更多资源 / More Resources

- [MULTI_SOURCE_DEPS.md](./MULTI_SOURCE_DEPS.md) - 多源依赖设计文档
- [DESIGN.md](./DESIGN.md) - 架构设计文档
- [AUDIT_REVIEW.md](./AUDIT_REVIEW.md) - 评估审计文档
- [README.md](../README.md) - 项目主文档
