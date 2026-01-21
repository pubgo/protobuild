# protobuild Design Document

## Overview

protobuild is a command-line tool designed to simplify Protocol Buffers development workflow. It provides unified configuration management, dependency handling, code generation, linting, and formatting capabilities.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      protobuild CLI                         │
├─────────────────────────────────────────────────────────────┤
│  Commands: gen | vendor | install | lint | format | version │
├─────────────────────────────────────────────────────────────┤
│                    Configuration Layer                       │
│              (protobuf.yaml / protobuf.plugin.yaml)          │
├─────────────────────────────────────────────────────────────┤
│   Dependency    │    Plugin      │   Linter    │  Formatter │
│   Manager       │    Manager     │   Engine    │   Engine   │
├─────────────────────────────────────────────────────────────┤
│                      protoc / Go Modules                     │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Configuration System

The configuration system supports hierarchical configuration with inheritance:

- **Root Configuration** (`protobuf.yaml`): Project-level configuration
- **Directory Configuration** (`protobuf.plugin.yaml`): Directory-level overrides

Configuration loading flow:

```
1. Load root protobuf.yaml
2. Walk proto directories
3. Check for protobuf.plugin.yaml in each directory
4. Merge configurations with inheritance
5. Apply base plugin settings
```

### 2. Dependency Manager

Responsible for managing proto file dependencies:

**Features:**
- Automatic version resolution via `go mod graph`
- Go module cache integration (`$GOPATH/pkg/mod`)
- Local path support
- Checksum-based change detection
- Optional dependencies

**Workflow:**
```
1. Parse dependencies from config
2. Resolve versions from go.mod or specified
3. Download/locate proto files
4. Copy to vendor directory
5. Update checksum
```

### 3. Plugin Manager

Manages protoc plugin execution:

**Plugin Types:**
- Standard protoc plugins (protoc-gen-*)
- Shell-based plugins
- Docker-based plugins

**Execution Flow:**
```
1. Load plugin configuration
2. Apply base settings
3. Build protoc command with options
4. Execute for each proto directory
5. Handle retag plugin specially (post-processing)
```

### 4. Linter Engine

Integrates with [api-linter](https://github.com/googleapis/api-linter) for proto file validation:

**Features:**
- AIP rule enforcement
- Custom rule enable/disable
- Multiple output formats (YAML, JSON, GitHub Actions)
- Comment-based disable support

### 5. Formatter Engine

Formats proto files using:
- [protocompile](https://github.com/bufbuild/protocompile) parser
- Custom formatting rules

## Data Flow

### Generation Flow

```
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│ protobuf.yaml│───▶│   Config    │───▶│   Walk Dir   │
└──────────────┘    │   Parser    │    │   (*.proto)  │
                    └─────────────┘    └──────────────┘
                                              │
                                              ▼
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│  Generated   │◀───│   protoc    │◀───│   Build Cmd  │
│    Code      │    │   Execute   │    │   with Opts  │
└──────────────┘    └─────────────┘    └──────────────┘
```

### Vendor Flow

```
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│     deps     │───▶│   Resolve   │───▶│   Download   │
│   config     │    │   Version   │    │   /Locate    │
└──────────────┘    └─────────────┘    └──────────────┘
                                              │
                                              ▼
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│   Update     │◀───│    Copy     │◀───│   Filter     │
│   Checksum   │    │   to Vendor │    │   .proto     │
└──────────────┘    └─────────────┘    └──────────────┘
```

## Configuration Schema

### Main Configuration (protobuf.yaml)

```yaml
# Auto-generated checksum for change detection
checksum: string

# Vendor directory path
vendor: string (default: .proto)

# Base plugin configuration
base:
  out: string      # Default output directory
  paths: string    # paths option (source_relative|import)
  module: string   # module prefix

# Proto source directories
root: []string

# Include paths for protoc -I
includes: []string

# Exclude paths from processing
excludes: []string

# Proto dependencies
deps:
  - name: string     # Local path in vendor
    url: string      # Module path or local path
    path: string     # Subdirectory in module
    version: string  # Specific version
    optional: bool   # Skip if not found

# Plugin configurations
plugins:
  - name: string         # Plugin name
    path: string         # Custom binary path
    out: string          # Output directory
    opt: string|[]string # Plugin options
    shell: string        # Shell command
    docker: string       # Docker image
    skip_base: bool      # Skip base config
    skip_run: bool       # Skip execution
    exclude_opts: []string # Excluded options

# Plugin installers
installers: []string

# Linter configuration
linter:
  rules:
    enabled_rules: []string
    disabled_rules: []string
  format_type: string
  ignore_comment_disables_flag: bool
```

## Key Design Decisions

### 1. YAML Configuration

**Rationale:** YAML provides a human-readable format with support for comments, making it easy to document and maintain configuration.

### 2. Go Module Integration

**Rationale:** Leverages existing Go toolchain for dependency resolution, avoiding the need for a separate dependency management system.

### 3. Middleware Pattern for Commands

**Rationale:** The middleware pattern (via redant) allows clean separation of concerns:
- Configuration parsing middleware
- Error handling
- Recovery mechanisms

### 4. Checksum-based Change Detection

**Rationale:** Avoids unnecessary vendor updates by tracking configuration changes via SHA1 checksums.

### 5. Hierarchical Configuration

**Rationale:** Allows directory-specific overrides while maintaining project-wide defaults, useful for monorepo structures.

## Error Handling

The project uses a consistent error handling approach:

1. **Assertions** (`assert.Must`, `assert.Exit`): For unrecoverable errors
2. **Recovery** (`recovery.Exit`, `recovery.Err`): For panic recovery
3. **Error Wrapping** (`errors.WrapTag`): For context-rich error messages

## Extension Points

### Custom Plugins

Support for three types of custom plugins:

1. **Binary Plugins**: Standard protoc plugins
2. **Shell Plugins**: Execute via shell commands
3. **Docker Plugins**: Execute via Docker containers

### Custom Linter Rules

Via the linter configuration:
- Enable specific AIP rules
- Disable rules globally or per-file
- Custom output formats

## Dependencies

Key external dependencies:

| Package | Purpose |
|---------|---------|
| `github.com/pubgo/redant` | CLI framework |
| `github.com/googleapis/api-linter` | Proto linting |
| `github.com/bufbuild/protocompile` | Proto parsing/formatting |
| `github.com/samber/lo` | Utility functions |
| `gopkg.in/yaml.v3` | YAML parsing |

## Future Considerations

1. **Remote Plugin Support**: Execute plugins via remote services
2. **Parallel Execution**: Concurrent proto compilation
3. **Watch Mode**: File watching for automatic regeneration
4. **Plugin Caching**: Cache plugin binaries for faster execution
5. **Proto Registry**: Integration with Buf Schema Registry
