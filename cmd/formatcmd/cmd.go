// Package formatcmd provides the format command for CLI.
package formatcmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/redant"

	"github.com/pubgo/protobuild/cmd/format"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/protobuild/internal/typex"
)

// FormatConfig holds format command configuration.
type FormatConfig struct {
	// Paths to format (directories or files)
	Paths []string
	// Write changes to files (default: false, just check)
	Write bool
	// Show diff instead of writing
	Diff bool
	// Exit with non-zero code if files need formatting
	ExitCode bool
	// Use builtin formatter instead of buf
	Builtin bool
}

// ProjectConfig holds project configuration from protobuf.yaml.
type ProjectConfig struct {
	// Root directories containing proto files
	Root []string
	// Vendor directory
	Vendor string
	// Include paths
	Includes []string
}

// New creates a new format command.
// The configProvider function is called to get project configuration.
func New(name string, configProvider func() *ProjectConfig) *redant.Command {
	var cfg FormatConfig

	return &redant.Command{
		Use:   name,
		Short: "Format Protobuf files using buf format",
		Long: `Format Protobuf files using buf format command.

By default, this command formats files in the 'root' directories defined in protobuf.yaml.
You can also specify paths explicitly.

Examples:
  # Format files in configured root directories
  protobuild format

  # Format files in place (write changes)
  protobuild format -w

  # Format specific paths
  protobuild format -w proto/ api/

  # Show diff of changes
  protobuild format --diff

  # Format and exit with error if changes needed (useful for CI)
  protobuild format --exit-code
`,
		Options: typex.Options{
			redant.Option{
				Flag:        "write",
				Shorthand:   "w",
				Description: "Write changes to files",
				Value:       redant.BoolOf(&cfg.Write),
			},
			redant.Option{
				Flag:        "diff",
				Shorthand:   "d",
				Description: "Show diff of changes",
				Value:       redant.BoolOf(&cfg.Diff),
			},
			redant.Option{
				Flag:        "exit-code",
				Description: "Exit with non-zero code if files need formatting (useful for CI)",
				Value:       redant.BoolOf(&cfg.ExitCode),
			},
			redant.Option{
				Flag:        "builtin",
				Description: "Use builtin formatter instead of buf",
				Value:       redant.BoolOf(&cfg.Builtin),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			// Get paths from args or use configured root directories
			cfg.Paths = inv.Args

			if len(cfg.Paths) == 0 && configProvider != nil {
				projectCfg := configProvider()
				if projectCfg != nil && len(projectCfg.Root) > 0 {
					cfg.Paths = projectCfg.Root
					slog.Info("using root directories from config", "paths", cfg.Paths)
				}
			}

			// Fallback to default if still empty
			if len(cfg.Paths) == 0 {
				cfg.Paths = []string{"proto"}
			}

			if cfg.Builtin {
				return runBuiltinFormat(cfg)
			}
			return runBufFormat(cfg)
		},
	}
}

// runBufFormat runs the buf format command.
func runBufFormat(cfg FormatConfig) error {
	bufPath, err := exec.LookPath("buf")
	if err != nil {
		slog.Error("buf not found in PATH")
		fmt.Println("\n💡 To install buf, run one of the following:")
		fmt.Println("   • brew install bufbuild/buf/buf")
		fmt.Println("   • go install github.com/bufbuild/buf/cmd/buf@latest")
		fmt.Println("   • See https://github.com/bufbuild/buf/releases")
		return errors.Wrap(err, "buf not found")
	}

	slog.Debug("using buf", "path", bufPath)

	for _, path := range cfg.Paths {
		// Check if path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			slog.Warn("path not found, skipping", "path", path)
			continue
		}

		args := []string{bufPath, "format", path}

		if cfg.Write {
			args = append(args, "-w")
		}
		if cfg.Diff {
			args = append(args, "--diff")
		}
		if cfg.ExitCode {
			args = append(args, "--exit-code")
		}

		slog.Info("formatting", "path", path, "write", cfg.Write)

		output, err := shutil.Run(args...)
		if err != nil {
			// If exit-code flag is set and files need formatting, buf exits with 1
			if cfg.ExitCode {
				fmt.Println("❌ Some files need formatting")
				if output != "" {
					fmt.Println(output)
				}
			}
			return err
		}

		if output != "" {
			fmt.Println(output)
		}
	}

	if cfg.Write {
		fmt.Println("✅ Files formatted successfully")
	} else if !cfg.Diff {
		fmt.Println("✅ All files are properly formatted")
	}

	return nil
}

// runBuiltinFormat runs the builtin formatter.
func runBuiltinFormat(cfg FormatConfig) error {
	slog.Info("using builtin formatter")

	for _, path := range cfg.Paths {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			slog.Warn("path not found, skipping", "path", path)
			continue
		}

		if info.IsDir() {
			err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(filePath, ".proto") {
					slog.Debug("formatting", "file", filePath)
					// Note: builtin format always writes to file
					format.Format(filePath)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(path, ".proto") {
			format.Format(path)
		}
	}

	fmt.Println("✅ Files formatted successfully")
	return nil
}
