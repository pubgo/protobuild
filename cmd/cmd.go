package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cnf/structhash"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/logx"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
	"github.com/pubgo/x/pathutil"
	"github.com/urfave/cli/v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/pubgo/protobuild/internal/modutil"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/protobuild/internal/utils"
	"github.com/pubgo/protobuild/version"
)

var (
	cfg      Cfg
	protoCfg = "protobuf.yaml"
	modPath  = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	pwd      = assert.Exit1(os.Getwd())
	logger   = logx.WithName("proto-build")
	//binPath  = []string{os.ExpandEnv("$HOME/bin"), os.ExpandEnv("$HOME/.local/bin"), os.ExpandEnv("./bin")}
)

func Main() *cli.App {
	var force bool
	var app = &cli.App{
		Name:    "protobuild",
		Usage:   "protobuf generation, configuration and management",
		Version: version.Version,
		Flags: typex.Flags{
			&cli.StringFlag{
				Name:        "conf",
				Usage:       "protobuf config path",
				Value:       protoCfg,
				Destination: &protoCfg,
			},
		},
		Before: func(ctx *cli.Context) error {
			defer recovery.Exit()

			content := assert.Must1(ioutil.ReadFile(protoCfg))
			assert.Must(yaml.Unmarshal(content, &cfg))

			cfg.Vendor = utils.FirstFnNotEmpty(func() string {
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
		},
		Commands: cli.Commands{
			fmtCmd(),
			&cli.Command{
				Name:  "gen",
				Usage: "编译protobuf文件",
				Action: func(ctx *cli.Context) error {
					defer recovery.Exit()

					var protoList sync.Map

					for i := range cfg.Root {
						if pathutil.IsNotExist(cfg.Root[i]) {
							log.Printf("file %s not flund", cfg.Root[i])
							continue
						}

						assert.Must(filepath.Walk(cfg.Root[i], func(path string, info fs.FileInfo, err error) error {
							if err != nil {
								return err
							}

							if info.IsDir() {
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
						var in = key.(string)

						var data = ""
						var base = fmt.Sprintf("protoc -I %s -I %s", cfg.Vendor, pwd)
						logger.Info(fmt.Sprintf("%v", cfg.Includes))
						for i := range cfg.Includes {
							base += fmt.Sprintf(" -I %s", cfg.Includes[i])
						}

						var retagOut = ""
						var retagOpt = ""
						for i := range cfg.Plugins {
							var plg = cfg.Plugins[i]

							var name = plg.Name

							// 指定plugin path
							if plg.Path != "" {
								data += fmt.Sprintf(" --plugin=%s=%s", name, plg.Path)
							}

							var out = func() string {
								// https://github.com/pseudomuto/protoc-gen-doc
								// 目录特殊处理
								if name == "doc" {
									var out = filepath.Join(plg.Out, in)
									assert.Must(pathutil.IsNotExistMkDir(out))
									return out
								}

								if plg.Out != "" {
									return plg.Out
								}

								return "."
							}()

							_ = pathutil.IsNotExistMkDir(out)

							var opts = func(dt interface{}) []string {
								switch _dt := dt.(type) {
								case string:
									if _dt != "" {
										return []string{_dt}
									}
								case []string:
									return _dt
								case []interface{}:
									var dtList []string
									for i := range _dt {
										dtList = append(dtList, _dt[i].(string))
									}
									return dtList
								}
								return nil
							}(plg.Opt)

							if name == "retag" {
								retagOut = fmt.Sprintf(" --%s_out=%s", name, out)
								retagOpt = fmt.Sprintf(" --%s_opt=%s", name, strings.Join(opts, ","))
								continue
							}

							data += fmt.Sprintf(" --%s_out=%s", name, out)

							if len(opts) > 0 {
								data += fmt.Sprintf(" --%s_opt=%s", name, strings.Join(opts, ","))
							}
						}
						data = base + data + " " + filepath.Join(in, "*.proto")
						logger.Info(data)
						assert.Must(shutil.Shell(data).Run(), data)
						if retagOut != "" && retagOpt != "" {
							data = base + retagOut + retagOpt + " " + filepath.Join(in, "*.proto")
						}
						logger.Info(data)
						assert.Must(shutil.Shell(data).Run(), data)
						return true
					})
					return nil
				},
			},
			&cli.Command{
				Name:  "vendor",
				Usage: "同步项目protobuf依赖到.proto中",
				Flags: typex.Flags{
					&cli.BoolFlag{
						Name:        "force",
						Usage:       "protobuf force vendor",
						Aliases:     []string{"f"},
						Value:       force,
						Destination: &force,
					},
				},
				Action: func(ctx *cli.Context) error {
					defer recovery.Exit()

					var changed bool

					// 解析go.mod并获取所有pkg版本
					var versions = modutil.LoadVersions()
					for i, dep := range cfg.Depends {
						var url = os.ExpandEnv(dep.Url)

						// url是本地目录, 不做检查
						if pathutil.IsDir(url) {
							continue
						}

						var v = utils.FirstFnNotEmpty(func() string {
							return versions[url]
						}, func() string {
							return dep.Version
						}, func() string {
							// go.mod中version不存在, 并且protobuf.yaml也没有指定
							// go pkg缓存
							var localPkg, err = ioutil.ReadDir(filepath.Dir(filepath.Join(modPath, url)))
							assert.Must(err)

							var _, name = filepath.Split(url)
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
							fmt.Println("go", "get", "-d", url+"/...")
							assert.Must(shutil.Shell("go", "get", "-d", url+"/...").Run())

							// 再次解析go.mod然后获取版本信息
							versions = modutil.LoadVersions()
							v = versions[url]

							assert.If(v == "", "%s version为空", url)
						}

						cfg.Depends[i].Version = v
					}
					assert.Must(ioutil.WriteFile(protoCfg, assert.Must1(yaml.Marshal(cfg)), 0644))

					if !changed && !cfg.changed && !force {
						fmt.Println("No changes")
						return nil
					}

					// 删除老的protobuf文件
					_ = os.RemoveAll(cfg.Vendor)

					for _, dep := range cfg.Depends {
						if dep.Name == "" || dep.Url == "" {
							continue
						}

						var url = os.ExpandEnv(dep.Url)
						var v = dep.Version

						// 加载版本
						if v != "" {
							url = fmt.Sprintf("%s@%s", url, v)
						}

						// 加载路径
						url = filepath.Join(url, dep.Path)

						if !utils.DirExists(url) {
							url = filepath.Join(modPath, url)
						}

						fmt.Println(url)

						url = assert.Must1(filepath.Abs(url))
						var newUrl = filepath.Join(cfg.Vendor, dep.Name)
						assert.Must(filepath.Walk(url, func(path string, info fs.FileInfo, err error) (gErr error) {
							if err != nil {
								return err
							}

							defer recovery.Err(&gErr, func(err xerr.XErr) xerr.XErr {
								return err.WrapF("path=%s name=%s", path, info.Name())
							})

							if info.IsDir() {
								return nil
							}

							if !strings.HasSuffix(info.Name(), ".proto") {
								return nil
							}

							var newPath = filepath.Join(newUrl, strings.TrimPrefix(path, url))
							assert.Must(pathutil.IsNotExistMkDir(filepath.Dir(newPath)))
							assert.Must1(copyFile(newPath, path))

							return nil
						}))
					}
					return nil
				},
			},
		},
	}
	return app
}

func copyFile(dstFilePath string, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	defer srcFile.Close()
	assert.Must(err, "打开源文件错误", srcFilePath)

	dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0444)
	defer srcFile.Close()
	assert.Must(err, "打开目标文件错误", dstFilePath)

	return io.Copy(dstFile, srcFile)
}
