package protobuild

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/protobuild/internal/depresolver"
	"github.com/pubgo/redant"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/protobuild/internal/typex"
)

// newGenCommand creates the gen command.
func newGenCommand() *redant.Command {
	return &redant.Command{
		Use:        "gen",
		Short:      "编译 protobuf 文件",
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			walker := NewProtoWalker(globalCfg.Root, globalCfg.Excludes)
			pluginMap := walker.CollectPluginConfigs(&globalCfg, protoPluginCfg)

			builder := NewProtocBuilder(globalCfg.Includes, globalCfg.Vendor, pwd)

			for protoPath, cfg := range pluginMap {
				if !walker.HasProtoFiles(protoPath) {
					continue
				}

				cmd := builder.BuildCommand(cfg, protoPath)
				if err := cmd.Execute(); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// newVendorCommand creates the vendor command.
func newVendorCommand(force, update *bool) *redant.Command {
	return &redant.Command{
		Use:   "vendor",
		Short: "同步项目 protobuf 依赖到 .proto 目录中",
		Options: typex.Options{
			redant.Option{
				Flag:        "force",
				Shorthand:   "f",
				Description: "protobuf force vendor",
				Value:       redant.BoolOf(force),
			},
			redant.Option{
				Flag:        "update",
				Shorthand:   "u",
				Description: "force re-download dependencies (ignore cache)",
				Value:       redant.BoolOf(update),
			},
		},
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			svc := NewVendorService(&globalCfg)
			result, err := svc.ResolveDependencies(ctx, *update)
			if err != nil {
				return err
			}

			if len(result.FailedDeps) > 0 {
				fmt.Printf("\n❌ Failed to resolve %d dependencies: %v\n", len(result.FailedDeps), result.FailedDeps)
				return fmt.Errorf("dependency resolution failed")
			}

			if len(result.ResolvedPaths) == 0 {
				fmt.Println("📦 No dependencies configured")
				return nil
			}

			if !result.Changed && !globalCfg.changed && !*force {
				fmt.Println("\n✨ No changes detected")
				return nil
			}

			copiedFiles, err := svc.CopyToVendor(result.ResolvedPaths)
			if err != nil {
				return err
			}

			// Update config file
			if err := saveConfig(); err != nil {
				return err
			}

			fmt.Printf("\n✅ Vendor complete! Copied %d proto files.\n", copiedFiles)
			return nil
		},
	}
}

// newDepsCommand creates the deps command.
func newDepsCommand() *redant.Command {
	return &redant.Command{
		Use:        "deps",
		Short:      "显示依赖列表及状态",
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(globalCfg.Depends) == 0 {
				fmt.Println("📭 No dependencies configured")
				return nil
			}

			resolver := depresolver.NewManager("", "")

			fmt.Println()
			fmt.Println("📦 Dependencies:")
			fmt.Println()
			fmt.Printf("  %-35s %-10s %-12s %s\n", "NAME", "SOURCE", "VERSION", "STATUS")
			fmt.Printf("  %-35s %-10s %-12s %s\n", "----", "------", "-------", "------")

			for _, dep := range globalCfg.Depends {
				if dep.Name == "" || dep.Url == "" {
					continue
				}

				source := depresolver.Source(dep.Source)
				if source == "" {
					source = depresolver.DetectSource(dep.Url)
				}

				version := getDepVersion(dep)
				status := getDepStatus(ctx, resolver, dep, source)
				optFlag := getOptionalFlag(dep)

				fmt.Printf("  %-35s %-10s %-12s %s%s\n",
					dep.Name, source.DisplayName(), version, status, optFlag)
			}

			fmt.Println()
			fmt.Printf("  Total: %d dependencies\n\n", len(globalCfg.Depends))
			return nil
		},
	}
}

// newCleanCommand creates the clean command.
func newCleanCommand(dryRun *bool) *redant.Command {
	return &redant.Command{
		Use:   "clean",
		Short: "清理依赖缓存",
		Options: typex.Options{
			redant.Option{
				Flag:        "dry-run",
				Description: "只显示要删除的内容，不实际删除",
				Value:       redant.BoolOf(dryRun),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			resolver := depresolver.NewManager("", "")
			cacheDir := resolver.CacheDir()

			totalSize, fileCount := calculateCacheSize(cacheDir)

			if fileCount == 0 {
				fmt.Println("📭 Cache is empty, nothing to clean.")
				return nil
			}

			sizeStr := formatBytes(totalSize)
			fmt.Printf("🗑️  Cache directory: %s\n", cacheDir)
			fmt.Printf("   Files: %d, Size: %s\n\n", fileCount, sizeStr)

			if *dryRun {
				fmt.Println("🔍 Dry-run mode: no files will be deleted.")
				return nil
			}

			fmt.Print("Cleaning...")
			if err := resolver.CleanCache(); err != nil {
				fmt.Println(" ❌")
				return fmt.Errorf("failed to clean cache: %w", err)
			}
			fmt.Println(" ✅")
			fmt.Printf("\n✨ Cleaned %d files (%s)\n", fileCount, sizeStr)
			return nil
		},
	}
}

// Helper functions

func getDepVersion(dep *depend) string {
	if dep.Version != nil && *dep.Version != "" {
		return *dep.Version
	}
	if dep.Ref != "" {
		return dep.Ref
	}
	return "-"
}

func getDepStatus(ctx context.Context, resolver *depresolver.Manager, dep *depend, source depresolver.Source) string {
	resolverDep := &depresolver.Dependency{
		Name:    dep.Name,
		Source:  source,
		URL:     dep.Url,
		Path:    dep.Path,
		Version: dep.Version,
		Ref:     dep.Ref,
	}

	result, err := resolver.Resolve(ctx, resolverDep)
	if err == nil && result.LocalPath != "" && pathutil.IsExist(result.LocalPath) {
		return "🟢 cached"
	}
	return "⚪ not cached"
}

func getOptionalFlag(dep *depend) string {
	if dep.Optional != nil && *dep.Optional {
		return " (optional)"
	}
	return ""
}

func calculateCacheSize(cacheDir string) (int64, int) {
	var totalSize int64
	var fileCount int

	if pathutil.IsDir(cacheDir) {
		_ = filepath.Walk(cacheDir, func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
				fileCount++
			}
			return nil
		})
	}

	return totalSize, fileCount
}

func saveConfig() error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	defer enc.Close()

	if err := enc.Encode(globalCfg); err != nil {
		return err
	}

	if err := os.WriteFile(protoCfg, buf.Bytes(), 0o666); err != nil {
		return err
	}

	if err := writeChecksumData(globalCfg.Vendor, []byte(globalCfg.Checksum)); err != nil {
		fmt.Printf("  ⚠️  Failed to write checksum: %s\n", err)
	}

	return nil
}
