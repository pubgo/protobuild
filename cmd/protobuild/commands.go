package protobuild

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pubgo/funk/v2/pathutil"
	"github.com/pubgo/funk/v2/recovery"
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

// writeLockfile persists the resolved dependency state next to protobuf.yaml.
// Best-effort: failures warn but never break the vendor flow.
func writeLockfile(resolvedPaths map[string]string) {
	if err := saveLockfile(buildLockfile(&globalCfg, resolvedPaths)); err != nil {
		fmt.Printf("  ⚠️  Failed to write lockfile: %s\n", err)
		return
	}
	fmt.Printf("🔒 已更新 %s\n", lockfileName)
}

// runVendorSync resolves dependencies, syncs the vendor directory, and keeps
// the generated artifacts (buf.yaml, lockfile) consistent.
func runVendorSync(ctx context.Context, update, force bool) error {
	svc := NewVendorService(&globalCfg)
	result, err := svc.ResolveDependencies(ctx, update)
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

	if !result.Changed && !globalCfg.Changed && !force {
		fmt.Println("\n✨ No changes detected")
		// keep buf.yaml in sync even when nothing was re-copied
		syncBufConfig(&globalCfg)
		writeLockfile(result.ResolvedPaths)
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

	// Keep buf.yaml in sync with the vendor layout for buf CLI / buf lsp.
	syncBufConfig(&globalCfg)
	writeLockfile(result.ResolvedPaths)

	fmt.Printf("\n✅ Vendor complete! Copied %d proto files.\n", copiedFiles)
	return nil
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

			return runVendorSync(ctx, *update, *force)
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
func newCleanCommand(dryRun *bool, maxAge, maxSize *string) *redant.Command {
	return &redant.Command{
		Use:   "clean",
		Short: "清理依赖缓存",
		Long: `清理依赖缓存。

默认清空全部缓存；通过 --max-age / --max-size 可按策略选择性清理：
  protobuild clean --max-age 30d           # 只清理 30 天未使用的缓存
  protobuild clean --max-size 500MB        # 从最旧开始清理直到缓存小于 500MB
  protobuild clean --max-age 30d --max-size 1GB --dry-run`,
		Options: typex.Options{
			redant.Option{
				Flag:        "dry-run",
				Description: "只显示要删除的内容，不实际删除",
				Value:       redant.BoolOf(dryRun),
			},
			redant.Option{
				Flag:        "max-age",
				Description: "只清理超过该时长未使用的缓存，如 30d / 720h",
				Value:       redant.StringOf(maxAge),
			},
			redant.Option{
				Flag:        "max-size",
				Description: "把缓存清理到该大小以内，如 500MB / 2GB",
				Value:       redant.StringOf(maxSize),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			resolver := depresolver.NewManager("", "")

			// selective mode
			if strings.TrimSpace(*maxAge) != "" || strings.TrimSpace(*maxSize) != "" {
				return cleanByPolicy(resolver.CacheDir(), *maxAge, *maxSize, *dryRun)
			}

			// full clean mode
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

// cleanByPolicy removes cache entries by age and/or size policy.
func cleanByPolicy(cacheDir, maxAge, maxSize string, dryRun bool) error {
	var cutoff time.Time
	if strings.TrimSpace(maxAge) != "" {
		d, err := parseCleanDuration(maxAge)
		if err != nil {
			return err
		}
		cutoff = time.Now().Add(-d)
	}

	var limit int64
	if strings.TrimSpace(maxSize) != "" {
		size, err := parseCleanSize(maxSize)
		if err != nil {
			return err
		}
		limit = size
	}

	entries := collectCacheEntries(cacheDir)
	if len(entries) == 0 {
		fmt.Println("📭 Cache is empty, nothing to clean.")
		return nil
	}

	remove := planCacheRemoval(entries, cutoff, limit)
	if len(remove) == 0 {
		fmt.Printf("📭 没有符合清理条件的缓存 (%d 个缓存项)\n", len(entries))
		return nil
	}

	var removeSize int64
	fmt.Printf("🗑️  将清理 %d 个缓存项:\n", len(remove))
	for _, e := range remove {
		fmt.Printf("   • %s (%s, %s)\n", e.name, formatBytes(e.size), e.modTime.Format("2006-01-02"))
		removeSize += e.size
	}

	if dryRun {
		fmt.Println("\n🔍 Dry-run mode: no files will be deleted.")
		return nil
	}

	fmt.Println()
	for _, e := range remove {
		if err := os.RemoveAll(e.path); err != nil {
			return fmt.Errorf("删除 %s 失败: %w", e.path, err)
		}
	}
	fmt.Printf("✨ 已清理 %d 个缓存项 (%s)\n", len(remove), formatBytes(removeSize))
	return nil
}

// Helper functions

func getDepVersion(dep *depend) string {
	if dep.Version != nil && *dep.Version != "" {
		return *dep.Version
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
