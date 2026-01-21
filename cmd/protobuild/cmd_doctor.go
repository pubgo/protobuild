package protobuild

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/redant"
)

// checkItem represents a single environment check.
type checkItem struct {
	Name        string
	Description string
	Check       func() checkResult
}

// checkResult represents the result of a check.
type checkResult struct {
	OK      bool
	Message string
	Help    string
}

// newDoctorCommand creates the doctor command.
func newDoctorCommand() *redant.Command {
	var fix bool

	return &redant.Command{
		Use:   "doctor",
		Short: "检查开发环境配置",
		Options: typex.Options{
			redant.Option{
				Flag:        "fix",
				Description: "尝试自动修复问题",
				Value:       redant.BoolOf(&fix),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()
			return runDoctor(fix)
		},
	}
}

// runDoctor executes the doctor command logic.
func runDoctor(fix bool) error {
	fmt.Println("🩺 Protobuild 环境检查")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	checks := []checkItem{
		{
			Name:        "protoc",
			Description: "Protocol Buffers 编译器",
			Check:       checkProtoc,
		},
		{
			Name:        "protoc-gen-go",
			Description: "Go Protobuf 插件",
			Check:       checkProtocGenGo,
		},
		{
			Name:        "protoc-gen-go-grpc",
			Description: "Go gRPC 插件",
			Check:       checkProtocGenGoGrpc,
		},
		{
			Name:        "buf",
			Description: "Buf CLI (可选，用于格式化)",
			Check:       checkBuf,
		},
		{
			Name:        "api-linter",
			Description: "API Linter (可选，用于代码检查)",
			Check:       checkApiLinter,
		},
		{
			Name:        "go",
			Description: "Go 编译器",
			Check:       checkGo,
		},
		{
			Name:        "config",
			Description: "项目配置文件",
			Check:       checkConfig,
		},
		{
			Name:        "vendor",
			Description: "Proto 依赖目录",
			Check:       checkVendor,
		},
	}

	var issues []checkItem
	var warnings []checkItem

	for _, item := range checks {
		result := item.Check()
		printCheckResult(item.Name, item.Description, result)

		if !result.OK {
			if isRequired(item.Name) {
				issues = append(issues, item)
			} else {
				warnings = append(warnings, item)
			}
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(issues) == 0 && len(warnings) == 0 {
		fmt.Println("✅ 所有检查通过！环境配置正确。")
		return nil
	}

	if len(issues) > 0 {
		fmt.Printf("❌ 发现 %d 个问题需要修复:\n", len(issues))
		for _, item := range issues {
			result := item.Check()
			fmt.Printf("   • %s: %s\n", item.Name, result.Message)
			if result.Help != "" {
				fmt.Printf("     💡 %s\n", result.Help)
			}
		}
	}

	if len(warnings) > 0 {
		fmt.Printf("⚠️  发现 %d 个可选组件未安装:\n", len(warnings))
		for _, item := range warnings {
			result := item.Check()
			fmt.Printf("   • %s: %s\n", item.Name, result.Message)
			if result.Help != "" {
				fmt.Printf("     💡 %s\n", result.Help)
			}
		}
	}

	if fix && len(issues) > 0 {
		fmt.Println("\n🔧 尝试自动修复...")
		autoFix()
	}

	return nil
}

// printCheckResult prints a formatted check result.
func printCheckResult(name, desc string, result checkResult) {
	status := "✅"
	if !result.OK {
		if isRequired(name) {
			status = "❌"
		} else {
			status = "⚠️ "
		}
	}

	fmt.Printf("%s %-20s %s\n", status, name, result.Message)
}

// isRequired returns true if the check is required (not optional).
func isRequired(name string) bool {
	optional := map[string]bool{
		"buf":        true,
		"api-linter": true,
	}
	return !optional[name]
}

// checkProtoc checks if protoc is installed.
func checkProtoc() checkResult {
	path, err := exec.LookPath("protoc")
	if err != nil {
		return checkResult{
			OK:      false,
			Message: "未安装",
			Help:    getProtocInstallHelp(),
		}
	}

	// Get version
	out, err := exec.Command("protoc", "--version").Output()
	if err != nil {
		return checkResult{OK: true, Message: fmt.Sprintf("已安装 (%s)", path)}
	}

	version := strings.TrimSpace(string(out))
	return checkResult{OK: true, Message: version}
}

// checkProtocGenGo checks if protoc-gen-go is installed.
func checkProtocGenGo() checkResult {
	path, err := exec.LookPath("protoc-gen-go")
	if err != nil {
		return checkResult{
			OK:      false,
			Message: "未安装",
			Help:    "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest",
		}
	}

	return checkResult{OK: true, Message: fmt.Sprintf("已安装 (%s)", path)}
}

// checkProtocGenGoGrpc checks if protoc-gen-go-grpc is installed.
func checkProtocGenGoGrpc() checkResult {
	path, err := exec.LookPath("protoc-gen-go-grpc")
	if err != nil {
		return checkResult{
			OK:      false,
			Message: "未安装",
			Help:    "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
		}
	}

	return checkResult{OK: true, Message: fmt.Sprintf("已安装 (%s)", path)}
}

// checkBuf checks if buf is installed.
func checkBuf() checkResult {
	path, err := exec.LookPath("buf")
	if err != nil {
		return checkResult{
			OK:      false,
			Message: "未安装 (可选)",
			Help:    "go install github.com/bufbuild/buf/cmd/buf@latest",
		}
	}

	out, err := exec.Command("buf", "--version").Output()
	if err != nil {
		return checkResult{OK: true, Message: fmt.Sprintf("已安装 (%s)", path)}
	}

	version := strings.TrimSpace(string(out))
	return checkResult{OK: true, Message: fmt.Sprintf("v%s", version)}
}

// checkApiLinter checks if api-linter is installed.
func checkApiLinter() checkResult {
	path, err := exec.LookPath("api-linter")
	if err != nil {
		return checkResult{
			OK:      false,
			Message: "未安装 (可选)",
			Help:    "go install github.com/googleapis/api-linter/cmd/api-linter@latest",
		}
	}

	return checkResult{OK: true, Message: fmt.Sprintf("已安装 (%s)", path)}
}

// checkGo checks if Go is installed.
func checkGo() checkResult {
	path, err := exec.LookPath("go")
	if err != nil {
		return checkResult{
			OK:      false,
			Message: "未安装",
			Help:    "请从 https://go.dev/dl/ 下载安装",
		}
	}

	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return checkResult{OK: true, Message: fmt.Sprintf("已安装 (%s)", path)}
	}

	// Extract version from "go version go1.21.0 darwin/amd64"
	parts := strings.Split(string(out), " ")
	if len(parts) >= 3 {
		return checkResult{OK: true, Message: parts[2]}
	}

	return checkResult{OK: true, Message: "已安装"}
}

// checkConfig checks if project config file exists.
func checkConfig() checkResult {
	if _, err := os.Stat(protoCfg); os.IsNotExist(err) {
		return checkResult{
			OK:      false,
			Message: fmt.Sprintf("%s 不存在", protoCfg),
			Help:    "运行 'protobuild init' 初始化项目",
		}
	}

	// Try to parse config
	if err := parseConfig(); err != nil {
		return checkResult{
			OK:      false,
			Message: fmt.Sprintf("配置文件解析错误: %v", err),
			Help:    "检查 YAML 语法是否正确",
		}
	}

	return checkResult{OK: true, Message: fmt.Sprintf("已配置 (%s)", protoCfg)}
}

// checkVendor checks if vendor directory exists and has dependencies.
func checkVendor() checkResult {
	if globalCfg.Vendor == "" {
		return checkResult{
			OK:      false,
			Message: "未配置 vendor 目录",
			Help:    "在配置文件中设置 vendor 字段",
		}
	}

	if _, err := os.Stat(globalCfg.Vendor); os.IsNotExist(err) {
		return checkResult{
			OK:      false,
			Message: fmt.Sprintf("%s 目录不存在", globalCfg.Vendor),
			Help:    "运行 'protobuild vendor' 同步依赖",
		}
	}

	// Count proto files in vendor
	count := 0
	filepath.Walk(globalCfg.Vendor, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			count++
		}
		return nil
	})

	if count == 0 {
		return checkResult{
			OK:      false,
			Message: fmt.Sprintf("%s 目录为空", globalCfg.Vendor),
			Help:    "运行 'protobuild vendor' 同步依赖",
		}
	}

	return checkResult{OK: true, Message: fmt.Sprintf("%s (%d 个 proto 文件)", globalCfg.Vendor, count)}
}

// getProtocInstallHelp returns platform-specific install instructions for protoc.
func getProtocInstallHelp() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install protobuf"
	case "linux":
		return "apt install -y protobuf-compiler 或从 https://github.com/protocolbuffers/protobuf/releases 下载"
	case "windows":
		return "从 https://github.com/protocolbuffers/protobuf/releases 下载"
	default:
		return "从 https://github.com/protocolbuffers/protobuf/releases 下载"
	}
}

// autoFix attempts to automatically fix common issues.
func autoFix() {
	// Check and install protoc-gen-go
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		fmt.Println("  安装 protoc-gen-go...")
		cmd := exec.Command("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("  ❌ 安装失败: %v\n", err)
		} else {
			fmt.Println("  ✅ protoc-gen-go 安装成功")
		}
	}

	// Check and install protoc-gen-go-grpc
	if _, err := exec.LookPath("protoc-gen-go-grpc"); err != nil {
		fmt.Println("  安装 protoc-gen-go-grpc...")
		cmd := exec.Command("go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("  ❌ 安装失败: %v\n", err)
		} else {
			fmt.Println("  ✅ protoc-gen-go-grpc 安装成功")
		}
	}

	// Run vendor if needed
	if globalCfg.Vendor != "" {
		if _, err := os.Stat(globalCfg.Vendor); os.IsNotExist(err) {
			fmt.Println("  同步依赖...")
			// This would need to call the vendor command
			fmt.Println("  💡 请手动运行 'protobuild vendor'")
		}
	}
}
