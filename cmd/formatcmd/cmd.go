// Package formatcmd provides the format command for CLI.
package formatcmd

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pubgo/funk/v2/errors"
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
	// Use clang-format instead of buf
	ClangFormat bool
	// clang-format style (e.g. file, google, llvm)
	ClangStyle string
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
	var cfg = FormatConfig{ClangStyle: "file"}

	return &redant.Command{
		Use:   name,
		Short: "Format Protobuf files (buf, builtin, or clang-format)",
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

	# Use clang-format engine
	protobuild format --clang-format -w

	# Use clang-format with specific style
	protobuild format --clang-format --clang-style google -w
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
			redant.Option{
				Flag:        "clang-format",
				Description: "Use clang-format instead of buf",
				Value:       redant.BoolOf(&cfg.ClangFormat),
			},
			redant.Option{
				Flag:        "clang-style",
				Description: "clang-format style (file, google, llvm, ...)",
				Default:     "file",
				Value:       redant.StringOf(&cfg.ClangStyle),
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

			if cfg.Builtin && cfg.ClangFormat {
				return fmt.Errorf("--builtin and --clang-format cannot be used together")
			}

			if cfg.ClangFormat {
				return runClangFormat(cfg)
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

// runClangFormat runs the clang-format command.
func runClangFormat(cfg FormatConfig) error {
	clangPath, err := exec.LookPath("clang-format")
	if err != nil {
		slog.Error("clang-format not found in PATH")
		fmt.Println("\n💡 To install clang-format, run one of the following:")
		fmt.Println("   • brew install clang-format")
		fmt.Println("   • apt-get install clang-format")
		fmt.Println("   • See https://clang.llvm.org/docs/ClangFormat.html")
		return errors.Wrap(err, "clang-format not found")
	}

	files, err := collectProtoFiles(cfg.Paths)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		slog.Warn("no proto files found to format", "paths", cfg.Paths)
		return nil
	}

	slog.Debug("using clang-format", "path", clangPath, "style", cfg.ClangStyle)

	var changedFiles []string

	for _, file := range files {
		original, err := os.ReadFile(file)
		if err != nil {
			return errors.Wrap(err, "read file failed")
		}

		formatted, err := getClangFormattedContent(clangPath, file, cfg.ClangStyle)
		if err != nil {
			return err
		}

		if bytes.Equal(original, formatted) {
			continue
		}

		changedFiles = append(changedFiles, file)

		if cfg.Write {
			if err := os.WriteFile(file, formatted, 0o644); err != nil {
				return errors.Wrap(err, "write formatted file failed")
			}
			continue
		}

		if cfg.Diff {
			diffOut, err := unifiedDiff(file, original, formatted)
			if err != nil {
				return err
			}
			if diffOut != "" {
				fmt.Print(diffOut)
			}
		}
	}

	if len(changedFiles) == 0 {
		fmt.Println("✅ All files are properly formatted")
		return nil
	}

	if cfg.Write {
		fmt.Printf("✅ Files formatted successfully (%d files)\n", len(changedFiles))
		return nil
	}

	if !cfg.Diff {
		fmt.Printf("⚠️  %d files need formatting\n", len(changedFiles))
	}

	if cfg.ExitCode {
		return fmt.Errorf("some files need formatting")
	}

	return nil
}

func getClangFormattedContent(clangPath, filePath, style string) ([]byte, error) {
	args := make([]string, 0, 2)
	if strings.TrimSpace(style) != "" {
		args = append(args, "--style="+style)
	}
	args = append(args, filePath)

	cmd := exec.Command(clangPath, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(ee.Stderr))
			if stderr != "" {
				return nil, fmt.Errorf("clang-format failed for %s: %s", filePath, stderr)
			}
		}
		return nil, errors.Wrap(err, "clang-format failed")
	}
	return out, nil
}

func collectProtoFiles(paths []string) ([]string, error) {
	files := make([]string, 0)
	seen := make(map[string]struct{})

	for _, p := range paths {
		info, err := os.Stat(p)
		if os.IsNotExist(err) {
			slog.Warn("path not found, skipping", "path", p)
			continue
		}
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			if strings.HasSuffix(p, ".proto") {
				if _, ok := seen[p]; !ok {
					seen[p] = struct{}{}
					files = append(files, p)
				}
			}
			continue
		}

		err = filepath.WalkDir(p, func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(filePath, ".proto") {
				if _, ok := seen[filePath]; !ok {
					seen[filePath] = struct{}{}
					files = append(files, filePath)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(files)
	return files, nil
}

func unifiedDiff(filePath string, original, formatted []byte) (string, error) {
	tmp, err := os.CreateTemp("", "protobuild-clang-format-*.proto")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(formatted); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	cmd := exec.Command("diff", "-u", "-L", filePath, "-L", filePath+" (clang-format)", filePath, tmpPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), nil
	}

	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		// diff exits with 1 when files differ, which is expected.
		return string(out), nil
	}

	return "", errors.Wrap(err, "failed to generate diff output")
}
