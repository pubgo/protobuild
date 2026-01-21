# protobuild

[![Go Report Card](https://goreportcard.com/badge/github.com/pubgo/protobuild)](https://goreportcard.com/report/github.com/pubgo/protobuild)
[![License](https://img.shields.io/github/license/pubgo/protobuild)](LICENSE)

> A powerful Protocol Buffers build and management tool

[中文文档](./README_CN.md)

## Features

- 🚀 **Unified Build** - One command to compile all proto files
- 📦 **Multi-Source Dependencies** - Support Go modules, Git, HTTP, S3, GCS, and local paths
- 🔌 **Plugin Support** - Flexible protoc plugin configuration
- 🔍 **Linting** - Built-in proto file linting with AIP rules
- 📝 **Formatting** - Auto-format proto files
- ⚙️ **Configuration-driven** - YAML-based project configuration
- 📊 **Progress Display** - Visual progress bars and detailed error messages
- 🗑️ **Cache Management** - Clean and manage dependency cache
- 🌐 **Web UI** - Visual configuration editor with proto file browser
- 🏥 **Environment Check** - Doctor command to diagnose development environment
- 🎯 **Project Initialization** - Quick project setup with templates

## Installation

```bash
go install github.com/pubgo/protobuild@latest
```

## Quick Start

1. Create a `protobuf.yaml` configuration file in your project root:

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

2. Vendor dependencies:

```bash
protobuild vendor
```

3. Generate code:

```bash
protobuild gen
```

## Commands

| Command | Description |
|---------|-------------|
| `gen` | Compile protobuf files |
| `vendor` | Sync proto dependencies to vendor directory |
| `vendor -u` | Force re-download all dependencies (ignore cache) |
| `deps` | Show dependency list and status |
| `install` | Install protoc plugins |
| `lint` | Lint proto files using AIP rules |
| `format` | Format proto files using buf |
| `format -w` | Format and write changes to files |
| `format --diff` | Show diff of formatting changes |
| `format --builtin` | Use builtin formatter instead of buf |
| `web` | Start web-based configuration UI |
| `web --port 9090` | Start web UI on custom port |
| `clean` | Clean dependency cache |
| `clean --dry-run` | Show what would be cleaned without deleting |
| `init` | Initialize a new protobuild project |
| `init --template grpc` | Initialize with specific template (basic, grpc, minimal) |
| `doctor` | Check development environment and dependencies |
| `doctor --fix` | Auto-install missing Go plugins |
| `version` | Show version information |

## Configuration

### Configuration File Structure

```yaml
# Checksum for tracking changes (auto-generated)
checksum: ""

# Vendor directory for proto dependencies
vendor: .proto

# Base plugin configuration (applied to all plugins)
base:
  out: pkg
  paths: source_relative
  module: github.com/your/module

# Proto source directories
root:
  - proto
  - api

# Include paths for protoc
includes:
  - proto
  - .proto

# Exclude paths from compilation
excludes:
  - proto/internal

# Proto dependencies
deps:
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf
    version: v21.0
    optional: false

# Protoc plugins configuration
plugins:
  - name: go
    out: pkg
    opt:
      - paths=source_relative
  - name: go-grpc
    out: pkg
    opt:
      - paths=source_relative

# Plugin installers (go install)
installers:
  - google.golang.org/protobuf/cmd/protoc-gen-go@latest
  - google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Linter configuration
linter:
  rules:
    enabled_rules:
      - core::0131::http-method
    disabled_rules:
      - all
  format_type: yaml
```

### Plugin Configuration

Each plugin supports the following options:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Plugin name (used as protoc-gen-{name}) |
| `path` | string | Custom plugin binary path |
| `out` | string | Output directory |
| `opt` | string/list | Plugin options |
| `shell` | string | Run via shell command |
| `docker` | string | Run via Docker container |
| `skip_base` | bool | Skip base configuration |
| `skip_run` | bool | Skip this plugin |
| `exclude_opts` | list | Options to exclude |

### Dependency Configuration

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Local name/path in vendor directory |
| `url` | string | Source URL (Go module, Git URL, HTTP archive, S3, GCS, or local path) |
| `path` | string | Subdirectory within the source |
| `version` | string | Specific version (for Go modules) |
| `ref` | string | Git ref (branch, tag, commit) for Git sources |
| `source` | string | Source type: `gomod`, `git`, `http`, `s3`, `gcs`, `local` (auto-detected if not specified) |
| `optional` | bool | Skip if not found |

#### Supported Dependency Sources

```yaml
deps:
  # Go Module (default)
  - name: google/protobuf
    url: github.com/protocolbuffers/protobuf
    path: src/google/protobuf

  # Git Repository
  - name: googleapis
    url: https://github.com/googleapis/googleapis.git
    ref: master

  # HTTP Archive
  - name: envoy
    url: https://github.com/envoyproxy/envoy/archive/v1.28.0.tar.gz
    path: api

  # Local Path
  - name: local-protos
    url: ./third_party/protos

  # S3 Bucket
  - name: internal-protos
    url: s3://my-bucket/protos.tar.gz

  # GCS Bucket
  - name: shared-protos
    url: gs://my-bucket/protos.tar.gz
```

## Usage Examples

### Custom Config File

```bash
protobuild -c protobuf.custom.yaml gen
```

### Lint Proto Files

```bash
protobuild lint
protobuild lint --list-rules  # Show available rules
protobuild lint --debug       # Debug mode
```

### Format Proto Files

```bash
# Format and preview changes (dry run)
protobuild format

# Format and write changes to files
protobuild format -w

# Show diff of formatting changes
protobuild format --diff

# Exit with error if files need formatting (useful for CI)
protobuild format --exit-code

# Use builtin formatter instead of buf
protobuild format --builtin

# Format specific directories
protobuild format -w proto/ api/
```

### Web Configuration UI

```bash
# Start web UI on default port (8080)
protobuild web

# Start web UI on custom port
protobuild web --port 9090
```

The web interface provides:
- 📝 Visual configuration editor
- 📦 Dependency management
- 🔌 Plugin configuration
- 🚀 One-click build, lint, format operations
- 📄 Real-time YAML preview
- 📊 Project statistics dashboard
- 🔍 Proto file browser with syntax highlighting
- 📚 Configuration examples reference

### Initialize New Project

```bash
# Interactive initialization
protobuild init

# Use specific template
protobuild init --template basic    # Basic Go + gRPC project
protobuild init --template grpc     # Full gRPC-Gateway project
protobuild init --template minimal  # Minimal configuration

# Specify output directory
protobuild init -o ./my-project
```

### Check Development Environment

```bash
# Diagnose environment issues
protobuild doctor

# Auto-install missing Go plugins
protobuild doctor --fix
```

Example output:
```
🏥 Protobuild Doctor

  Checking development environment...

  ✅ protoc                 installed (v25.1)
  ✅ protoc-gen-go          installed
  ✅ protoc-gen-go-grpc     installed
  ✅ buf                    installed (v1.28.1)
  ✅ api-linter             installed
  ✅ go                     installed (go1.21.5)
  ✅ Configuration          protobuf.yaml found
  ⚠️  Vendor directory       not found (run 'protobuild vendor')

  ✅ Environment check passed!
```

### Force Vendor Update

```bash
protobuild vendor -f      # Force update even if no changes detected
protobuild vendor -u      # Re-download all dependencies (ignore cache)
```

### Show Dependency Status

```bash
protobuild deps
```

Example output:
```
📦 Dependencies:

  NAME                                SOURCE     VERSION      STATUS
  ----                                ------     -------      ------
  google/protobuf                     Go Module  v21.0        🟢 cached
  googleapis                          Git        master       ⚪ not cached

  Total: 2 dependencies
```

### Clean Dependency Cache

```bash
protobuild clean           # Clean all cached dependencies
protobuild clean --dry-run # Preview what will be cleaned
```

### Install Plugins

```bash
protobuild install
protobuild install -f  # Force reinstall
```

## Directory-level Configuration

You can place a `protobuf.plugin.yaml` file in any proto directory to override the root configuration for that directory and its subdirectories.

```yaml
# proto/api/protobuf.plugin.yaml
plugins:
  - name: go
    out: pkg/api
    opt:
      - paths=source_relative
```

## Supported Protoc Plugins

- `google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest`
- `github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest`
- `github.com/bufbuild/protoc-gen-validate/cmd/protoc-gen-validate@latest`
- And many more...

## Error Handling

When dependency resolution fails, protobuild provides detailed error messages with suggestions:

```
❌ Failed to download dependency: google/protobuf
   Source:  Git
   URL:     git::https://github.com/protocolbuffers/protobuf.git?ref=v99.0
   Ref:     v99.0
   Error:   reference not found

💡 Suggestions:
   • Check if the repository URL is correct and accessible
   • Verify the ref (tag/branch/commit) exists
   • Ensure you have proper authentication (SSH key or token)
```

## Cache Location

Dependencies are cached in:
- **macOS/Linux**: `~/.cache/protobuild/deps/`
- **Go modules**: Standard Go module cache (`$GOPATH/pkg/mod`)

## Architecture

```
protobuild
├── cmd/
│   ├── protobuild/          # Main CLI application
│   │   ├── cmd.go           # Entry point and core handlers
│   │   ├── commands.go      # Command factory functions
│   │   ├── config.go        # Configuration structs
│   │   ├── proto_walker.go  # Proto file walking utilities
│   │   ├── protoc_builder.go# Protoc command builder
│   │   ├── vendor_service.go# Dependency vendoring
│   │   ├── util.go          # Shared utilities
│   │   └── yaml_types.go    # YAML type definitions
│   ├── format/              # Proto file formatting (builtin)
│   ├── formatcmd/           # Format command (buf integration)
│   ├── linters/             # AIP linting rules
│   └── webcmd/              # Web configuration UI
│       ├── cmd.go           # Web command entry
│       ├── server.go        # HTTP server and API
│       └── templates/       # HTML templates (Alpine.js + Tailwind)
└── internal/
    ├── depresolver/         # Multi-source dependency resolver
    ├── modutil/             # Go module utilities
    ├── plugin/              # Plugin management
    ├── protoutil/           # Protobuf utilities
    ├── shutil/              # Shell utilities
    └── typex/               # Type extensions
```

## Documentation

- [Configuration Examples](./docs/EXAMPLES.md) - Detailed configuration examples for various use cases
- [Multi-Source Dependencies](./docs/MULTI_SOURCE_DEPS.md) - Design document for multi-source dependency resolution
- [Design Document](./docs/DESIGN.md) - Architecture and design documentation

## Roadmap

Upcoming features planned for future releases:

| Feature | Description | Status |
|---------|-------------|--------|
| 🔗 **Dependency Graph** | Visualize proto file import dependencies | Planned |
| ⚠️ **Breaking Change Detection** | Detect incompatible changes between versions | Planned |
| 📚 **API Documentation Generator** | Auto-generate Markdown/HTML docs from proto comments | Planned |
| 🎭 **Mock Server** | Auto-start mock gRPC/HTTP server for testing | Planned |
| 📝 **Proto Templates** | Quick generation of common proto patterns (CRUD, pagination) | Planned |
| 📊 **Field Statistics** | Analyze field naming conventions and type distribution | Planned |
| ✏️ **Online Editor** | Edit proto files directly in Web UI | Planned |

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.