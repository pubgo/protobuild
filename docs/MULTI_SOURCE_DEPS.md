# Multi-Source Dependency Design / 多源依赖设计

## Problem / 问题

Current implementation relies heavily on Go modules for dependency management:
- Uses `go mod graph` to resolve versions
- Uses `go get` to download dependencies  
- Stores in `$GOPATH/pkg/mod`

This makes it difficult for non-Go projects (Python, Java, JavaScript, etc.) to use protobuild.

当前实现严重依赖 Go 模块进行依赖管理：
- 使用 `go mod graph` 解析版本
- 使用 `go get` 下载依赖
- 存储在 `$GOPATH/pkg/mod`

这使得非 Go 项目（Python、Java、JavaScript 等）难以使用 protobuild。

## Proposed Solution / 建议方案

Implement a multi-source dependency resolver that supports various dependency sources.

实现支持多种依赖源的多源依赖解析器。

### Configuration Schema / 配置模式

```yaml
deps:
  # Source type: gomod (default, backward compatible)
  # 源类型：gomod（默认，向后兼容）
  - name: google/api
    source: gomod
    url: github.com/googleapis/googleapis
    path: google/api
    version: v0.0.0-20230822172742-b8732ec3820d
    
  # Source type: git
  # 源类型：git
  - name: google/protobuf
    source: git
    url: https://github.com/protocolbuffers/protobuf.git
    path: src/google/protobuf
    ref: v21.0           # tag, branch, or commit hash
    depth: 1             # shallow clone depth (optional)
    
  # Source type: http (tar.gz, zip)
  # 源类型：http（tar.gz, zip）
  - name: envoy/api
    source: http
    url: https://github.com/envoyproxy/envoy/archive/refs/tags/v1.25.0.tar.gz
    path: api/envoy
    strip: 1             # strip leading directory components
    
  # Source type: buf (Buf Schema Registry)
  # 源类型：buf（Buf Schema 注册表）
  - name: buf/validate
    source: buf
    url: buf.build/bufbuild/protovalidate
    version: v0.5.0
    
  # Source type: local (already supported)
  # 源类型：local（已支持）
  - name: local/proto
    source: local
    url: /path/to/proto
```

### Architecture / 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Dependency Manager                        │
├─────────────────────────────────────────────────────────────┤
│                    Source Router                             │
├─────────┬─────────┬─────────┬─────────┬────────────────────┤
│  GoMod  │   Git   │  HTTP   │   Buf   │   Local            │
│ Resolver│ Resolver│ Resolver│ Resolver│  Resolver          │
├─────────┴─────────┴─────────┴─────────┴────────────────────┤
│                    Cache Manager                             │
│              (~/.cache/protobuild/deps)                      │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Plan / 实现计划

#### Phase 1: Abstraction Layer / 抽象层

```go
// internal/depresolver/resolver.go

type Source string

const (
    SourceGoMod Source = "gomod"
    SourceGit   Source = "git"
    SourceHTTP  Source = "http"
    SourceBuf   Source = "buf"
    SourceLocal Source = "local"
)

type Dependency struct {
    Name     string  `yaml:"name"`
    Source   Source  `yaml:"source,omitempty"` // default: auto-detect or gomod
    URL      string  `yaml:"url"`
    Path     string  `yaml:"path,omitempty"`
    Version  string  `yaml:"version,omitempty"`
    Ref      string  `yaml:"ref,omitempty"`      // for git
    Depth    int     `yaml:"depth,omitempty"`    // for git
    Strip    int     `yaml:"strip,omitempty"`    // for http archives
    Optional bool    `yaml:"optional,omitempty"`
}

type Resolver interface {
    // Resolve resolves the dependency and returns the local path
    Resolve(dep *Dependency) (string, error)
    
    // Supports checks if this resolver supports the given dependency
    Supports(dep *Dependency) bool
}

type ResolverChain struct {
    resolvers []Resolver
    cache     *CacheManager
}

func (r *ResolverChain) Resolve(dep *Dependency) (string, error) {
    for _, resolver := range r.resolvers {
        if resolver.Supports(dep) {
            return resolver.Resolve(dep)
        }
    }
    return "", fmt.Errorf("no resolver found for source: %s", dep.Source)
}
```

#### Phase 2: Individual Resolvers / 独立解析器

**Using go-getter library / 使用 go-getter 库:**

The implementation uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) for unified downloading across multiple protocols:

实现使用 [hashicorp/go-getter](https://github.com/hashicorp/go-getter) 统一处理多种协议的下载：

- **Git** - Clone via `git::` prefix with `?ref=` query param for version/tag/branch
- **HTTP** - Automatic archive extraction (tar.gz, zip, etc.)
- **S3** - AWS S3 bucket access via `s3://` or `s3::` prefix
- **GCS** - Google Cloud Storage via `gcs://` or `gs://` prefix

```go
// Example: Git source with version
// url: git::https://github.com/protocolbuffers/protobuf.git?ref=v21.0

// Example: S3 source  
// url: s3://bucket-name/path/to/proto.tar.gz

// Example: HTTP archive (auto-extracted)
// url: https://github.com/user/repo/archive/v1.0.0.tar.gz
```

**GoMod Resolver (backward compatible):**
```go
// Uses existing go mod download mechanism
// Stores in $GOPATH/pkg/mod
```

**Buf Resolver:**
```go
type BufResolver struct {
    cacheDir string
}

func (r *BufResolver) Resolve(dep *Dependency) (string, error) {
    // Use buf CLI: buf export buf.build/owner/repo -o /cache/path
    cmd := exec.Command("buf", "export", dep.URL, "-o", cachePath)
    // ...
}
```

#### Phase 3: Cache Management / 缓存管理

```go
type CacheManager struct {
    baseDir string // ~/.cache/protobuild
}

func (c *CacheManager) GetPath(source Source, key string) string {
    return filepath.Join(c.baseDir, string(source), hashString(key))
}

func (c *CacheManager) Clean() error {
    return os.RemoveAll(c.baseDir)
}
```

### Migration Path / 迁移路径

1. **Backward Compatibility**: If `source` is not specified, auto-detect:
   - If URL looks like Go module path → use `gomod`
   - If URL ends with `.git` → use `git`
   - If URL starts with `http://` or `https://` and is archive → use `http`
   - If URL starts with `buf.build/` → use `buf`
   - If URL is local path → use `local`

2. **Deprecation Warning**: Show warning for implicit `gomod` detection

1. **向后兼容**：如果未指定 `source`，自动检测：
   - 如果 URL 看起来像 Go 模块路径 → 使用 `gomod`
   - 如果 URL 以 `.git` 结尾 → 使用 `git`
   - 如果 URL 以 `http://` 或 `https://` 开头且是归档 → 使用 `http`
   - 如果 URL 以 `buf.build/` 开头 → 使用 `buf`
   - 如果 URL 是本地路径 → 使用 `local`

2. **弃用警告**：对隐式 `gomod` 检测显示警告

### Benefits / 优点

1. **Language Agnostic**: Works with any language/framework
2. **Flexible**: Multiple source types for different use cases
3. **Cacheable**: Unified cache management
4. **Backward Compatible**: Existing configs still work
5. **Extensible**: Easy to add new source types

1. **语言无关**：适用于任何语言/框架
2. **灵活**：多种源类型满足不同用例
3. **可缓存**：统一缓存管理
4. **向后兼容**：现有配置仍然有效
5. **可扩展**：易于添加新的源类型

### Example Configurations / 配置示例

**For Python Project / Python 项目:**
```yaml
vendor: .proto
deps:
  - name: google/protobuf
    source: git
    url: https://github.com/protocolbuffers/protobuf.git
    ref: v21.0
    path: src/google/protobuf
    
  - name: googleapis
    source: git
    url: https://github.com/googleapis/googleapis.git
    ref: master
    path: google
```

**For TypeScript Project / TypeScript 项目:**
```yaml
vendor: proto
deps:
  - name: validate
    source: buf
    url: buf.build/bufbuild/protovalidate
    
  - name: google/api
    source: http
    url: https://github.com/googleapis/googleapis/archive/master.tar.gz
    strip: 1
    path: google/api
```

**For Go Project (current behavior) / Go 项目（当前行为）:**
```yaml
vendor: .proto
deps:
  - name: google/api
    url: github.com/googleapis/googleapis
    path: google/api
    # source: gomod (implied)
```
