package protobuild

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/protobuild/internal/config"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/redant"
)

// initTemplates defines available project templates.
var initTemplates = map[string]*config.Config{
	"basic": {
		Vendor: ".proto",
		BasePlugin: &config.BasePluginCfg{
			Out:   "./pkg",
			Paths: "source_relative",
		},
		Root:     []string{"proto"},
		Includes: []string{"proto"},
		Plugins: []*config.Plugin{
			{Name: "go"},
			{Name: "go-grpc"},
		},
	},
	"grpc-gateway": {
		Vendor: ".proto",
		BasePlugin: &config.BasePluginCfg{
			Out:    "./pkg",
			Paths:  "import",
			Module: "",
		},
		Root:     []string{"proto"},
		Includes: []string{"proto"},
		Depends: []*config.Depend{
			{Name: "google/api", Url: "https://github.com/googleapis/googleapis", Path: "google/api"},
			{Name: "protoc-gen-openapiv2/options", Url: "https://github.com/grpc-ecosystem/grpc-gateway", Path: "protoc-gen-openapiv2/options"},
		},
		Linter: &config.Linter{
			Rules: &config.LinterRules{
				EnabledRules: []string{
					"core::0131::http-method",
					"core::0131::http-body",
					"core::0235::plural-method-name",
				},
				DisabledRules: []string{"all"},
			},
			FormatType: "yaml",
		},
		Plugins: []*config.Plugin{
			{Name: "go"},
			{Name: "go-grpc"},
			{Name: "grpc-gateway", Opts: config.PluginOpts{"generate_unbound_methods=true"}},
			{Name: "openapiv2", Out: "./docs/swagger"},
		},
		Installers: []string{
			"go install google.golang.org/protobuf/cmd/protoc-gen-go@latest",
			"go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
			"go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest",
			"go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest",
		},
	},
	"minimal": {
		Vendor: ".proto",
		BasePlugin: &config.BasePluginCfg{
			Out:   "./pkg",
			Paths: "source_relative",
		},
		Root:     []string{"proto"},
		Includes: []string{"proto"},
		Plugins: []*config.Plugin{
			{Name: "go"},
		},
	},
}

// newInitCommand creates the init command.
func newInitCommand() *redant.Command {
	var template string
	var force bool

	return &redant.Command{
		Use:   "init",
		Short: "初始化 protobuf 项目配置",
		Options: typex.Options{
			redant.Option{
				Flag:        "template",
				Shorthand:   "t",
				Description: "项目模板 (basic, grpc-gateway, minimal)",
				Default:     "",
				Value:       redant.StringOf(&template),
			},
			redant.Option{
				Flag:        "force",
				Shorthand:   "f",
				Description: "覆盖已存在的配置文件",
				Value:       redant.BoolOf(&force),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()
			return runInit(template, force)
		},
	}
}

// runInit executes the init command logic.
func runInit(template string, force bool) error {
	// Check if config already exists
	if _, err := os.Stat(protoCfg); err == nil && !force {
		return fmt.Errorf("配置文件 %s 已存在，使用 --force 覆盖", protoCfg)
	}

	var cfg *config.Config

	if template != "" {
		// Use specified template
		tmpl, ok := initTemplates[template]
		if !ok {
			fmt.Println("可用的模板:")
			for name := range initTemplates {
				fmt.Printf("  - %s\n", name)
			}
			return fmt.Errorf("未知的模板: %s", template)
		}
		cfg = tmpl
		fmt.Printf("📦 使用模板: %s\n", template)
	} else {
		// Interactive mode
		cfg = interactiveInit()
	}

	// Try to detect Go module
	if cfg.BasePlugin != nil && cfg.BasePlugin.Module == "" {
		if mod := detectGoModule(); mod != "" {
			cfg.BasePlugin.Module = mod + "/pkg"
			fmt.Printf("🔍 检测到 Go 模块: %s\n", mod)
		}
	}

	// Create proto directory if not exists
	for _, dir := range cfg.Root {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
		fmt.Printf("📁 创建目录: %s\n", dir)
	}

	// Seed an example proto when the root dir is still empty (best-effort;
	// existing user protos are never touched — check is inside the function).
	hasExample := false
	if len(cfg.Root) > 0 {
		created, err := createExampleProto(cfg.Root[0])
		if err != nil {
			fmt.Printf("  ⚠️  示例 proto 生成失败: %v\n", err)
		}
		hasExample = created
	}

	// Save config
	if err := config.Save(protoCfg, cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("✅ 配置文件已创建: %s\n", protoCfg)

	// Generate buf.yaml so buf CLI / buf lsp work from the very beginning.
	syncBufConfig(cfg)

	fmt.Println("\n下一步:")
	if len(cfg.Depends) > 0 {
		fmt.Println("  1. 运行 'protobuild vendor' 同步依赖")
	}
	if hasExample {
		fmt.Println("  · 示例 proto 已生成于 proto/example/v1/，直接 'protobuild gen' 即可验证链路")
	} else {
		fmt.Println("  · 在 proto/ 目录下创建 .proto 文件")
	}
	fmt.Println("  · 运行 'protobuild gen' 生成代码")

	return nil
}

// interactiveInit runs interactive configuration setup.
func interactiveInit() *config.Config {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Protobuild 项目初始化")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Proto source directory
	fmt.Print("Proto 源目录 [proto]: ")
	protoDir, _ := reader.ReadString('\n')
	protoDir = strings.TrimSpace(protoDir)
	if protoDir == "" {
		protoDir = "proto"
	}

	// Output directory
	fmt.Print("代码输出目录 [./pkg]: ")
	outDir, _ := reader.ReadString('\n')
	outDir = strings.TrimSpace(outDir)
	if outDir == "" {
		outDir = "./pkg"
	}

	// Path mode
	fmt.Print("路径模式 (source_relative/import) [source_relative]: ")
	pathMode, _ := reader.ReadString('\n')
	pathMode = strings.TrimSpace(pathMode)
	if pathMode == "" {
		pathMode = "source_relative"
	}

	// gRPC support
	fmt.Print("是否需要 gRPC 支持? (y/n) [y]: ")
	grpcInput, _ := reader.ReadString('\n')
	grpcInput = strings.TrimSpace(strings.ToLower(grpcInput))
	needGrpc := grpcInput == "" || grpcInput == "y" || grpcInput == "yes"

	// Build config
	cfg := &config.Config{
		Vendor: ".proto",
		BasePlugin: &config.BasePluginCfg{
			Out:   outDir,
			Paths: pathMode,
		},
		Root:     []string{protoDir},
		Includes: []string{protoDir},
		Plugins: []*config.Plugin{
			{Name: "go"},
		},
	}

	if needGrpc {
		cfg.Plugins = append(cfg.Plugins, &config.Plugin{Name: "go-grpc"})
	}

	return cfg
}

// detectGoModule tries to detect the Go module name from go.mod.
func detectGoModule() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	return ""
}

// createExampleProto creates an example proto file if the directory is empty.
// It reports whether an example was actually written.
func createExampleProto(protoDir string) (bool, error) {
	// Check if directory is empty
	entries, err := os.ReadDir(protoDir)
	if err != nil {
		return false, err
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".proto") {
			return false, nil // Already has proto files
		}
	}

	// Create example proto
	exampleDir := filepath.Join(protoDir, "example", "v1")
	if err := os.MkdirAll(exampleDir, 0o755); err != nil {
		return false, err
	}

	module := detectGoModule()
	if module == "" {
		module = "github.com/yourorg/yourproject"
	}

	example := fmt.Sprintf(`syntax = "proto3";

package example.v1;

option go_package = "%s/pkg/example/v1;examplev1";

// HelloRequest is the request message for Hello.
message HelloRequest {
  string name = 1;
}

// HelloResponse is the response message for Hello.
message HelloResponse {
  string message = 1;
}

// GreeterService is a simple greeting service.
service GreeterService {
  // SayHello returns a greeting.
  rpc SayHello(HelloRequest) returns (HelloResponse);
}
`, module)

	exampleFile := filepath.Join(exampleDir, "greeter.proto")
	if err := os.WriteFile(exampleFile, []byte(example), 0o644); err != nil {
		return false, err
	}

	fmt.Printf("📝 创建示例文件: %s\n", exampleFile)
	return true, nil
}
