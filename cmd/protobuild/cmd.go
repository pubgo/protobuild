package protobuild

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/a8m/envsubst"
	"github.com/cnf/structhash"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/strutil"
	cli "github.com/urfave/cli/v3"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	yaml "gopkg.in/yaml.v3"

	"github.com/pubgo/protobuild/internal/modutil"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/protobuild/version"

	_ "golang.org/x/mod/module"
)

var (
	cfg Cfg

	protoCfg = "protobuf.yaml"
	modPath  = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	pwd      = assert.Exit1(os.Getwd())
	logger   = log.GetLogger("proto-build")
	// binPath  = []string{os.ExpandEnv("$HOME/bin"), os.ExpandEnv("$HOME/.local/bin"), os.ExpandEnv("./bin")}
)

func parseConfig() (gErr error) {
	defer recovery.Err(&gErr)

	content := assert.Must1(os.ReadFile(protoCfg))
	content = assert.Must1(envsubst.Bytes(content))
	assert.Must(yaml.Unmarshal(content, &cfg))

	cfg.Vendor = strutil.FirstFnNotEmpty(func() string {
		return cfg.Vendor
	}, func() string {
		protoPath := filepath.Join(pwd, ".proto")
		if pathutil.IsExist(protoPath) {
			return protoPath
		}
		return ""
	}, func() string {
		goModPath := filepath.Dir(modutil.GoModPath())
		if goModPath == "" {
			panic("没有找到项目go.mod文件")
		}

		return filepath.Join(goModPath, ".proto")
	})

	assert.Must(pathutil.IsNotExistMkDir(cfg.Vendor))

	// protobuf文件检查
	for _, dep := range cfg.Depends {
		assert.If(dep.Name == "" || dep.Url == "", "name和url都不能为空")
	}

	checksum := fmt.Sprintf("%x", structhash.Sha1(cfg, 1))
	if cfg.Checksum != checksum {
		cfg.Checksum = checksum
		cfg.changed = true
	}

	return nil
}

func Main() *cli.Command {
	var force bool
	app := &cli.Command{
		Name:                  "protobuf",
		Usage:                 "protobuf generation, configuration and management",
		Version:               version.Version,
		ShellComplete:         cli.DefaultAppComplete,
		EnableShellCompletion: true,
		Suggest:               true,
		Flags: typex.Flags{
			&cli.StringFlag{
				Name:        "conf",
				Aliases:     typex.Strs{"c", "f"},
				Usage:       "protobuf config path",
				Value:       protoCfg,
				Hidden:      false,
				Persistent:  true,
				Destination: &protoCfg,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if shutil.IsHelp() {
				return nil
			}

			file := os.Stdin
			fi := assert.Exit1(file.Stat())
			if fi.Size() == 0 {
				return cli.ShowAppHelp(c)
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

			for _, p := range cfg.Plugins {
				if p.Name != plgName {
					continue
				}

				log.Printf("%#v\n", p)

				if p.Shell != "" {
					cccc := shutil.Shell(strings.TrimSpace(p.Shell))
					cccc.Stdin = bytes.NewBuffer(assert.Must1(proto.Marshal(req)))
					assert.Must(cccc.Run())
					break
				}

				if p.Docker != "" {
					cccc := shutil.Shell("docker run -i --rm " + p.Docker)
					cccc.Stdin = bytes.NewBuffer(assert.Must1(proto.Marshal(req)))
					assert.Must(cccc.Run())
					break
				}
			}

			return nil
		},
		Commands: typex.Commands{
			&cli.Command{
				Name:  "gen",
				Usage: "编译 protobuf 文件",
				Before: func(ctx context.Context, c *cli.Command) error {
					return parseConfig()
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					defer recovery.Exit()

					var protoList sync.Map
					basePlugin := cfg.BasePlugin
					for i := range cfg.Root {
						if pathutil.IsNotExist(cfg.Root[i]) {
							log.Printf("file %s not found", cfg.Root[i])
							continue
						}

						assert.Must(filepath.Walk(cfg.Root[i], func(path string, info fs.FileInfo, err error) error {
							if err != nil {
								return err
							}

							if info.IsDir() {
								for _, e := range cfg.Excludes {
									if strings.HasPrefix(path, e) {
										return filepath.SkipDir
									}
								}
								return nil
							}

							if !strings.HasSuffix(info.Name(), ".proto") {
								return nil
							}

							protoList.Store(filepath.Dir(path), struct{}{})
							return nil
						}))
					}

					protoList.Range(func(key, _ interface{}) bool {
						in := key.(string)

						data := ""
						base := fmt.Sprintf("protoc -I %s -I %s", cfg.Vendor, pwd)
						logger.Info().Msgf("includes=%q", cfg.Includes)
						for i := range cfg.Includes {
							base += fmt.Sprintf(" -I %s", cfg.Includes[i])
						}

						retagOut := ""
						retagOpt := ""
						for i := range cfg.Plugins {
							plg := cfg.Plugins[i]

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
									out := filepath.Join(plg.Out, in)
									assert.Must(pathutil.IsNotExistMkDir(out))
									return out
								}

								if plg.Out != "" {
									return plg.Out
								}

								if basePlugin.Out != "" {
									return basePlugin.Out
								}

								return "."
							}()

							assert.Exit(pathutil.IsNotExistMkDir(out))

							opts := plg.Opt
							hasPath := func() bool {
								for _, opt := range opts {
									if strings.HasPrefix(opt, "paths=") {
										return true
									}
								}
								return false
							}

							hasModule := func() bool {
								for _, opt := range opts {
									if strings.HasPrefix(opt, "module=") {
										return true
									}
								}
								return false
							}

							if !hasPath() && basePlugin.Paths != "" && !plg.SkipBase {
								opts = append(opts, fmt.Sprintf("paths=%s", basePlugin.Paths))
							}

							if !hasModule() && basePlugin.Module != "" && !plg.SkipBase {
								opts = append(opts, fmt.Sprintf("module=%s", basePlugin.Module))
							}

							if plg.Shell != "" || plg.Docker != "" {
								opts = append(opts, "__wrapper="+name)
								data += fmt.Sprintf(" --plugin=protoc-gen-%s=%s", name, assert.Must1(exec.LookPath(os.Args[0])))
							}

							if name == "retag" {
								retagOut = fmt.Sprintf(" --%s_out=%s", name, out)
								retagOpt = fmt.Sprintf(" --%s_opt=%s", name, strings.Join(opts, ","))
								continue
							}

							data += fmt.Sprintf(" --%s_out=%s", name, out)

							if len(opts) > 0 {
								var protoOpt []string
								for _, opt := range opts {
									if !hasAny(plg.ExcludeOpts, func(d string) bool { return strings.HasPrefix(opt, d) }) {
										protoOpt = append(protoOpt, opt)
									}
								}
								data += fmt.Sprintf(" --%s_opt=%s", name, strings.Join(protoOpt, ","))
							}
						}
						data = base + data + " " + filepath.Join(in, "*.proto")
						logger.Info().Msg(data)
						assert.Must(shutil.Shell(data).Run(), data)
						if retagOut != "" && retagOpt != "" {
							data = base + retagOut + retagOpt + " " + filepath.Join(in, "*.proto")
							logger.Info().Bool("retag", true).Msg(data)
							assert.Must(shutil.Shell(data).Run(), data)
						}

						return true
					})
					return nil
				},
			},
			&cli.Command{
				Name:  "vendor",
				Usage: "同步项目 protobuf 依赖到 .proto 目录中",
				Before: func(ctx context.Context, c *cli.Command) error {
					return parseConfig()
				},
				Flags: typex.Flags{
					&cli.BoolFlag{
						Name:        "force",
						Usage:       "protobuf force vendor",
						Aliases:     []string{"f"},
						Value:       force,
						Destination: &force,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					defer recovery.Exit()

					var changed bool

					// 解析go.mod并获取所有pkg版本
					versions := modutil.LoadVersions()
					for i, dep := range cfg.Depends {
						pathVersion := strings.SplitN(dep.Url, "@", 2)
						if len(pathVersion) == 2 {
							dep.Version = generic.Ptr(pathVersion[1])
							dep.Path = pathVersion[0]
						}

						url := os.ExpandEnv(dep.Url)

						// url是本地目录, 不做检查
						if pathutil.IsDir(url) {
							continue
						}

						if pathutil.IsNotExist(url) && dep.Optional != nil && *dep.Optional {
							continue
						}

						v := strutil.FirstFnNotEmpty(func() string {
							return versions[url]
						}, func() string {
							return generic.DePtr(dep.Version)
						}, func() string {
							// go.mod中version不存在, 并且protobuf.yaml也没有指定
							// go pkg缓存
							localPkg, err := os.ReadDir(filepath.Dir(filepath.Join(modPath, url)))
							assert.Must(err)

							_, name := filepath.Split(url)
							for j := range localPkg {
								if !localPkg[j].IsDir() {
									continue
								}

								if strings.HasPrefix(localPkg[j].Name(), name+"@") {
									return strings.TrimPrefix(localPkg[j].Name(), name+"@")
								}
							}
							return ""
						})

						if v == "" || pathutil.IsNotExist(fmt.Sprintf("%s/%s@%s", modPath, url, v)) {
							changed = true
							if v == "" {
								fmt.Println("go", "get", "-d", url+"/...")
								assert.Must(shutil.Shell("go", "get", "-d", url+"/...").Run())

							} else if pathutil.IsNotExist(fmt.Sprintf("%s/%s@%s", modPath, url, v)) {
								fmt.Println("go", "get", "-d", fmt.Sprintf("%s@%s", url, v))
								assert.Must(shutil.Shell("go", "get", "-d", fmt.Sprintf("%s@%s", url, v)).Run())
							}

							// 再次解析go.mod然后获取版本信息
							versions = modutil.LoadVersions()
							v = versions[url]
							assert.If(v == "", "%s version为空", url)
						}

						cfg.Depends[i].Version = generic.Ptr(v)
					}

					if !changed && !cfg.changed && !force {
						fmt.Println("No changes")
						return nil
					}

					// 删除老的protobuf文件
					logger.Info().Str("vendor", cfg.Vendor).Msg("delete old vendor")
					_ = os.RemoveAll(cfg.Vendor)

					for _, dep := range cfg.Depends {
						if dep.Name == "" || dep.Url == "" {
							continue
						}

						url := os.ExpandEnv(dep.Url)
						v := generic.DePtr(dep.Version)

						// 加载版本
						if v != "" {
							url = fmt.Sprintf("%s@%s", url, v)
						}

						// 加载路径
						url = filepath.Join(url, dep.Path)

						if pathutil.IsNotExist(url) {
							url = filepath.Join(modPath, url)
						}

						fmt.Println(url)

						url = assert.Must1(filepath.Abs(url))

						if pathutil.IsNotExist(url) && dep.Optional != nil && *dep.Optional {
							continue
						}

						newUrl := filepath.Join(cfg.Vendor, dep.Name)
						assert.Must(filepath.Walk(url, func(path string, info fs.FileInfo, err error) (gErr error) {
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

							newPath := filepath.Join(newUrl, strings.TrimPrefix(path, url))
							assert.Must(pathutil.IsNotExistMkDir(filepath.Dir(newPath)))
							assert.Must1(copyFile(newPath, path))

							return nil
						}))
					}

					// TODO 强制更新配置文件, 可以考虑参数化
					{
						var buf bytes.Buffer
						enc := yaml.NewEncoder(&buf)
						enc.SetIndent(2)
						defer enc.Close()
						assert.Must(enc.Encode(cfg))
						assert.Must(os.WriteFile(protoCfg, buf.Bytes(), 0o666))
					}
					return nil
				},
			},
		},
	}
	return app
}

func copyFile(dstFilePath, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	defer srcFile.Close()
	assert.Must(err, "打开源文件错误", srcFilePath)

	dstFile, err := os.Create(dstFilePath)
	defer srcFile.Close()
	assert.Must(err, "打开目标文件错误", dstFilePath)

	return io.Copy(dstFile, srcFile)
}

func hasAny(data []string, fn func(d string) bool) bool {
	for _, d := range data {
		if fn(d) {
			return true
		}
	}
	return false
}
