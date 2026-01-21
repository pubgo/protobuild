package protobuild

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/protobuild/cmd/formatcmd"
	linters "github.com/pubgo/protobuild/cmd/linters"
	"github.com/pubgo/protobuild/internal/depresolver"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/redant"
	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	yaml "gopkg.in/yaml.v3"
)

var (
	globalCfg Config

	protoCfg       = "protobuf.yaml"
	protoPluginCfg = "protobuf.plugin.yaml"
	pwd            = assert.Exit1(os.Getwd())
	logger         = log.GetLogger("protobuild")
	// binPath  = []string{os.ExpandEnv("$HOME/bin"), os.ExpandEnv("$HOME/.local/bin"), os.ExpandEnv("./bin")}
)

const (
	reTagPluginName = "retag"
)

// withParseConfig 返回一个解析配置的中间件
func withParseConfig() redant.MiddlewareFunc {
	return func(next redant.HandlerFunc) redant.HandlerFunc {
		return func(ctx context.Context, inv *redant.Invocation) error {
			if err := parseConfig(); err != nil {
				slog.Error("failed to parse config", "err", err)
				return err
			}
			return next(ctx, inv)
		}
	}
}

func Main(ver string) *redant.Command {
	var force bool
	var update bool
	cliArgs, options := linters.NewCli()
	app := &redant.Command{
		Use:   "protobuf",
		Short: "protobuf generation, configuration and management",
		Options: typex.Options{
			redant.Option{
				Flag:        "conf",
				Shorthand:   "c",
				Description: "protobuf config path",
				Default:     protoCfg,
				Value:       redant.StringOf(&protoCfg),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if shutil.IsHelp() {
				return nil
			}

			file := os.Stdin
			if term.IsTerminal(int(file.Fd())) {
				return nil
			}

			fi := assert.Exit1(file.Stat())
			if fi.Size() == 0 {
				return nil
			}

			in := assert.Must1(io.ReadAll(file))
			req := &pluginpb.CodeGeneratorRequest{}
			assert.Must(proto.Unmarshal(in, req))

			var opts protogen.Options
			plg := assert.Must1(opts.New(req))
			for _, f := range plg.Files {
				if !f.Generate {
					continue
				}

				log.Printf("%s\n", f.GeneratedFilenamePrefix)
			}

			var params []string
			var plgName string
			for _, p := range strings.Split(req.GetParameter(), ",") {
				if strings.HasPrefix(p, "__wrapper") {
					names := strings.Split(p, "=")
					plgName = strings.TrimSpace(names[len(names)-1])
				} else {
					params = append(params, p)
				}
			}

			if len(params) > 0 {
				req.Parameter = generic.Ptr(strings.Join(params, ","))
			}

			for _, p := range globalCfg.Plugins {
				if p.Name != plgName {
					continue
				}

				log.Printf("%#v\n", p)

				if p.Shell != "" {
					cmd := shutil.Shell(strings.TrimSpace(p.Shell))
					cmd.Stdin = bytes.NewBuffer(assert.Must1(proto.Marshal(req)))
					assert.Must(cmd.Run())
					break
				}

				if p.Docker != "" {
					cmd := shutil.Shell("docker run -i --rm " + p.Docker)
					cmd.Stdin = bytes.NewBuffer(assert.Must1(proto.Marshal(req)))
					assert.Must(cmd.Run())
					break
				}
			}

			return nil
		},
		Children: typex.Commands{
			&redant.Command{
				Use:        "gen",
				Short:      "编译 protobuf 文件",
				Middleware: withParseConfig(),
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					defer recovery.Exit()

					var pluginMap = make(map[string]*Config)
					for i := range globalCfg.Root {
						if pathutil.IsNotExist(globalCfg.Root[i]) {
							log.Printf("file %s not found", globalCfg.Root[i])
							continue
						}

						assert.Must(filepath.WalkDir(globalCfg.Root[i], func(path string, d fs.DirEntry, err error) error {
							if err != nil {
								return err
							}

							if !d.IsDir() {
								return nil
							}

							// check protobuf.plugin.yaml
							var pluginCfg *Config
							pluginCfgPath := filepath.Join(path, protoPluginCfg)
							if pathutil.IsExist(pluginCfgPath) {
								pluginCfg = parsePluginConfig(pluginCfgPath)
							} else {
								for dir, v := range pluginMap {
									if strings.HasPrefix(path, dir) {
										pluginCfg = v
										break
									}
								}
							}
							pluginCfg = mergePluginConfig(&globalCfg, pluginCfg)

							for _, e := range pluginCfg.Excludes {
								if strings.HasPrefix(path, e) {
									return filepath.SkipDir
								}
							}

							pluginMap[path] = pluginCfg
							return nil
						}))
					}

					for protoSourcePath, pp := range pluginMap {
						// check contains proto file in dir
						hasProto := lo.ContainsBy(
							assert.Must1(os.ReadDir(protoSourcePath)),
							func(item os.DirEntry) bool {
								return !item.IsDir() && strings.HasSuffix(item.Name(), ".proto")
							},
						)
						if !hasProto {
							continue
						}

						var doF = func(pluginCfg *Config, protoPath string) {
							data := ""

							includes := lo.Uniq(append(pluginCfg.Includes, pluginCfg.Vendor, pwd))
							base := "protoc"
							for i := range includes {
								base += fmt.Sprintf(" -I %s", includes[i])
							}

							reTagData := ""
							for i := range pluginCfg.Plugins {
								plg := pluginCfg.Plugins[i]
								if plg.SkipRun {
									continue
								}

								name := plg.Name

								// 指定plugin path
								if plg.Path != "" {
									plg.Path = assert.Must1(exec.LookPath(plg.Path))
									assert.If(pathutil.IsNotExist(plg.Path), "plugin path notfound, path=%s", plg.Path)
									data += fmt.Sprintf(" --plugin=protoc-gen-%s=%s", name, plg.Path)
								}

								out := func() string {
									// https://github.com/pseudomuto/protoc-gen-doc
									// 目录特殊处理
									if name == "doc" {
										out := filepath.Join(plg.Out, protoPath)
										assert.Must(pathutil.IsNotExistMkDir(out))
										return out
									}

									if plg.Out != "" {
										return plg.Out
									}

									if pluginCfg.BasePlugin.Out != "" {
										return pluginCfg.BasePlugin.Out
									}

									return "."
								}()

								assert.Exit(pathutil.IsNotExistMkDir(out))

								opts := append(plg.Opt, plg.Opts...)
								hasPath := lo.ContainsBy(opts, func(opt string) bool {
									return strings.HasPrefix(opt, "paths=")
								})

								hasModule := lo.ContainsBy(opts, func(opt string) bool {
									return strings.HasPrefix(opt, "module=")
								})

								if !hasPath && pluginCfg.BasePlugin.Paths != "" && !plg.SkipBase {
									opts = append(opts, fmt.Sprintf("paths=%s", pluginCfg.BasePlugin.Paths))
								}

								if !hasModule && pluginCfg.BasePlugin.Module != "" && !plg.SkipBase {
									opts = append(opts, fmt.Sprintf("module=%s", pluginCfg.BasePlugin.Module))
								}

								if plg.Shell != "" || plg.Docker != "" {
									opts = append(opts, "__wrapper="+name)
									data += fmt.Sprintf(" --plugin=protoc-gen-%s=%s", name, assert.Must1(exec.LookPath(os.Args[0])))
								}

								if name == reTagPluginName {
									reTagData = fmt.Sprintf(" --%s_out=%s", name, out)
									opts = append(opts, "__out="+out)

									reTagData += fmt.Sprintf(" --%s_opt=%s", name, strings.Join(opts, ","))
									if plg.Path != "" {
										reTagData += fmt.Sprintf(" --plugin=protoc-gen-%s=%s", name, plg.Path)
									}
									continue
								}

								data += fmt.Sprintf(" --%s_out=%s", name, out)

								if len(opts) > 0 {
									var protoOpt []string
									for _, opt := range opts {
										if !lo.ContainsBy(plg.ExcludeOpts, func(d string) bool { return strings.HasPrefix(opt, d) }) {
											protoOpt = append(protoOpt, opt)
										}
									}
									data += fmt.Sprintf(" --%s_opt=%s", name, strings.Join(protoOpt, ","))
								}
							}
							data = base + data + " " + filepath.Join(protoPath, "*.proto")
							logger.Info().Msg(data)
							assert.Exit(shutil.Shell(data).Run(), data)
							if reTagData != "" {
								data = base + reTagData + " " + filepath.Join(protoPath, "*.proto")
								logger.Info().Bool(reTagPluginName, true).Msg(data)
								assert.Exit(shutil.Shell(data).Run(), data)
							}
						}
						doF(pp, protoSourcePath)
					}
					return nil
				},
			},
			&redant.Command{
				Use:   "vendor",
				Short: "同步项目 protobuf 依赖到 .proto 目录中",
				Options: typex.Options{
					redant.Option{
						Flag:        "force",
						Shorthand:   "f",
						Description: "protobuf force vendor",
						Value:       redant.BoolOf(&force),
					},
					redant.Option{
						Flag:        "update",
						Shorthand:   "u",
						Description: "force re-download dependencies (ignore cache)",
						Value:       redant.BoolOf(&update),
					},
				},
				Middleware: withParseConfig(),
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					defer recovery.Exit()

					// Filter valid dependencies
					var validDeps []*depend
					for _, dep := range globalCfg.Depends {
						if dep.Name != "" && dep.Url != "" {
							validDeps = append(validDeps, dep)
						}
					}

					if len(validDeps) == 0 {
						fmt.Println("📦 No dependencies configured")
						return nil
					}

					fmt.Printf("\n🔍 Resolving %d dependencies...\n\n", len(validDeps))

					// Create dependency resolver manager
					resolver := depresolver.NewManager("", "")

					// Clean cache if --update flag is set
					if update {
						fmt.Println("🗑️  Cleaning dependency cache...")
						_ = resolver.CleanCache()
						fmt.Println()
					}

					var changed bool
					var resolvedPaths = make(map[string]string) // dep.Name -> localPath
					var failedDeps []string

					// Resolve all dependencies using the multi-source resolver
					for i, dep := range globalCfg.Depends {
						if dep.Name == "" || dep.Url == "" {
							continue
						}

						// Convert config depend to depresolver.Dependency
						resolverDep := &depresolver.Dependency{
							Name:     dep.Name,
							Source:   depresolver.Source(dep.Source), // "" = auto-detect
							URL:      dep.Url,
							Path:     dep.Path,
							Version:  dep.Version,
							Ref:      dep.Ref,
							Optional: dep.Optional,
						}

						// Resolve dependency
						result, err := resolver.Resolve(ctx, resolverDep)
						if err != nil {
							if dep.Optional != nil && *dep.Optional {
								fmt.Printf("  ⚠️  [optional] %s - skipped\n", dep.Name)
								continue
							}
							// Print detailed error and continue to show all failures
							fmt.Print(err.Error())
							failedDeps = append(failedDeps, dep.Name)
							continue
						}

						// Skip if empty result (optional not found)
						if result.LocalPath == "" {
							continue
						}

						if result.Changed {
							changed = true
							fmt.Printf("  ✅ %s (downloaded)\n", dep.Name)
						} else {
							fmt.Printf("  ✅ %s (cached)\n", dep.Name)
						}

						// Update version in config if resolved
						if result.Version != "" {
							globalCfg.Depends[i].Version = &result.Version
						}

						resolvedPaths[dep.Name] = result.LocalPath
					}

					// Check if any dependencies failed
					if len(failedDeps) > 0 {
						fmt.Printf("\n❌ Failed to resolve %d dependencies: %v\n", len(failedDeps), failedDeps)
						return fmt.Errorf("dependency resolution failed")
					}

					if !changed && !globalCfg.changed && !force {
						fmt.Println("\n✨ No changes detected")
						return nil
					}

					// Delete old vendor directory
					fmt.Printf("\n📁 Updating vendor directory: %s\n", globalCfg.Vendor)
					_ = os.RemoveAll(globalCfg.Vendor)

					// Count total .proto files first
					var totalFiles int
					for _, localPath := range resolvedPaths {
						_ = filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) error {
							if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
								totalFiles++
							}
							return nil
						})
					}

					// Create progress bar for copying
					bar := progressbar.NewOptions(totalFiles,
						progressbar.OptionSetDescription("  📋 Copying proto files"),
						progressbar.OptionShowCount(),
						progressbar.OptionSetWidth(30),
						progressbar.OptionOnCompletion(func() { fmt.Println() }),
					)

					// Copy resolved dependencies to vendor
					var copiedFiles int
					for name, localPath := range resolvedPaths {
						if pathutil.IsNotExist(localPath) {
							fmt.Printf("  ⚠️  Path not found: %s (%s)\n", name, localPath)
							continue
						}

						newUrl := filepath.Join(globalCfg.Vendor, name)
						assert.Must(filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) (gErr error) {
							if err != nil {
								return err
							}

							defer recovery.Err(&gErr, func(err error) error {
								return errors.WrapTag(err,
									errors.T("path", path),
									errors.T("name", info.Name()),
								)
							})

							if info.IsDir() {
								return nil
							}

							if !strings.HasSuffix(info.Name(), ".proto") {
								return nil
							}

							newPath := filepath.Join(newUrl, strings.TrimPrefix(path, localPath))
							assert.Must(pathutil.IsNotExistMkDir(filepath.Dir(newPath)))
							assert.Must1(copyFile(newPath, path))
							copiedFiles++
							_ = bar.Add(1)

							return nil
						}))
					}
					_ = bar.Finish()

					// Update config file with resolved versions
					{
						var buf bytes.Buffer
						enc := yaml.NewEncoder(&buf)
						enc.SetIndent(2)
						defer enc.Close()
						assert.Must(enc.Encode(globalCfg))
						assert.Must(os.WriteFile(protoCfg, buf.Bytes(), 0o666))

						err := writeChecksumData(globalCfg.Vendor, []byte(globalCfg.Checksum))
						if err != nil {
							fmt.Printf("  ⚠️  Failed to write checksum: %s\n", err)
						}
					}

					fmt.Printf("\n✅ Vendor complete! Copied %d proto files.\n", copiedFiles)
					return nil
				},
			},
			&redant.Command{
				Use:   "install",
				Short: "install protobuf plugin",
				Options: typex.Options{
					redant.Option{
						Flag:        "force",
						Shorthand:   "f",
						Description: "force update protobuf plugin",
						Value:       redant.BoolOf(&force),
					},
				},
				Middleware: withParseConfig(),
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					defer recovery.Exit()

					for _, plg := range globalCfg.Installers {
						if !strings.Contains(plg, "@") {
							pluginPaths := strings.Split(plg, "@")
							plg = strings.Join(pluginPaths[:len(pluginPaths)-1], "@") + "@latest"
						}

						plgName := strings.Split(lo.LastOrEmpty(strings.Split(plg, "/")), "@")[0]
						path, err := exec.LookPath(plgName)
						if err != nil {
							slog.Error("command not found", slog.Any("name", plgName))
						}

						if err == nil && !globalCfg.changed && !force {
							slog.Info("no changes", slog.Any("path", path))
							continue
						}

						slog.Info("install command", slog.Any("name", plg))
						assert.Must(shutil.Shell("go", "install", plg).Run())
					}
					return nil
				},
			},
			&redant.Command{
				Use:        "lint",
				Short:      "lint protobuf https://linter.aip.dev/rules/",
				Options:    options,
				Middleware: withParseConfig(),
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					var protoPaths []string
					for i := range globalCfg.Root {
						if pathutil.IsNotExist(globalCfg.Root[i]) {
							log.Printf("file %s not found", globalCfg.Root[i])
							continue
						}

						assert.Must(filepath.WalkDir(globalCfg.Root[i], func(path string, d fs.DirEntry, err error) error {
							if err != nil {
								return err
							}

							if d.IsDir() {
								protoPaths = append(protoPaths, path)
							}

							return nil
						}))
					}

					protoPaths = lo.Uniq(protoPaths)
					for _, path := range protoPaths {
						// check contains proto file in dir
						protoFiles := lo.Map(assert.Must1(os.ReadDir(path)), func(item os.DirEntry, index int) string {
							return filepath.Join(path, item.Name())
						})
						protoFiles = lo.Filter(protoFiles, func(item string, index int) bool { return strings.HasSuffix(item, ".proto") })
						if len(protoFiles) == 0 {
							continue
						}

						includes := lo.Uniq(append(globalCfg.Includes, globalCfg.Vendor))
						err := linters.Linter(cliArgs, globalCfg.Linter, includes, protoFiles)
						if err != nil {
							return err
						}
					}

					return nil
				},
			},
			formatcmd.New("format"),
			&redant.Command{
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

						// Detect source type
						source := depresolver.Source(dep.Source)
						if source == "" {
							source = depresolver.DetectSource(dep.Url)
						}

						// Get version
						version := "-"
						if dep.Version != nil && *dep.Version != "" {
							version = *dep.Version
						} else if dep.Ref != "" {
							version = dep.Ref
						}

						// Check cache status
						status := "⚪ not cached"
						resolverDep := &depresolver.Dependency{
							Name:    dep.Name,
							Source:  source,
							URL:     dep.Url,
							Path:    dep.Path,
							Version: dep.Version,
							Ref:     dep.Ref,
						}

						// Try to check if it's in cache
						result, err := resolver.Resolve(context.Background(), resolverDep)
						if err == nil && result.LocalPath != "" {
							if pathutil.IsExist(result.LocalPath) {
								status = "🟢 cached"
							}
						}

						// Check if optional
						optFlag := ""
						if dep.Optional != nil && *dep.Optional {
							optFlag = " (optional)"
						}

						fmt.Printf("  %-35s %-10s %-12s %s%s\n",
							dep.Name,
							source.DisplayName(),
							version,
							status,
							optFlag,
						)
					}

					fmt.Println()
					fmt.Printf("  Total: %d dependencies\n\n", len(globalCfg.Depends))
					return nil
				},
			},
			&redant.Command{
				Use:   "clean",
				Short: "清理依赖缓存",
				Options: typex.Options{
					redant.Option{
						Flag:        "dry-run",
						Description: "只显示要删除的内容，不实际删除",
						Value:       redant.BoolOf(&force), // reuse force as dry-run
					},
				},
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					resolver := depresolver.NewManager("", "")
					cacheDir := resolver.CacheDir()

					// Calculate cache size
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

					if fileCount == 0 {
						fmt.Println("📭 Cache is empty, nothing to clean.")
						return nil
					}

					// Format size
					sizeStr := formatBytes(totalSize)
					fmt.Printf("🗑️  Cache directory: %s\n", cacheDir)
					fmt.Printf("   Files: %d, Size: %s\n\n", fileCount, sizeStr)

					if force { // dry-run mode
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
			},
			&redant.Command{
				Use:   "version",
				Short: "version info",
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					defer recovery.Exit()
					fmt.Printf("Project:   %s\n", running.Project)
					fmt.Printf("Version:   %s\n", running.Version)
					fmt.Printf("GitCommit: %s\n", running.CommitID)
					return nil
				},
			},
		},
	}
	return app
}

func copyFile(dstFilePath, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	assert.Must(err, "打开源文件错误", srcFilePath)
	defer srcFile.Close()

	dstFile, err := os.Create(dstFilePath)
	assert.Must(err, "打开目标文件错误", dstFilePath)
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

// formatBytes formats bytes into a human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
