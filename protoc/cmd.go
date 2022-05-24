package protoc

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

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/pubgo/protobuild/pkg/modutil"
	"github.com/pubgo/protobuild/pkg/shutil"
	"github.com/pubgo/protobuild/pkg/typex"
	"github.com/pubgo/protobuild/pkg/utils"
	"github.com/pubgo/protobuild/version"
)

var (
	cfg      Cfg
	protoCfg = "protobuf.yaml"
	modPath  = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	pwd      = xerror.ExitErr(os.Getwd()).(string)
)

func Main() {
	var app = &cli.App{
		Name:    "prototool",
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
			defer xerror.RespExit()

			content := xerror.PanicBytes(ioutil.ReadFile(protoCfg))
			xerror.Panic(yaml.Unmarshal(content, &cfg))

			cfg.ProtoPath = utils.FirstFnNotEmpty(func() string {
				return cfg.ProtoPath
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

			xerror.Panic(pathutil.IsNotExistMkDir(cfg.ProtoPath))

			// protobuf文件检查
			for _, dep := range cfg.Depends {
				xerror.Assert(dep.Name == "" || dep.Url == "", "name和url都不能为空")
			}
			return nil
		},
		Commands: cli.Commands{
			&cli.Command{
				Name:  "gen",
				Usage: "编译protobuf文件",
				Action: func(ctx *cli.Context) error {
					defer xerror.RespExit()

					var protoList sync.Map

					for i := range cfg.Root {
						if pathutil.IsNotExist(cfg.Root[i]) {
							log.Printf("file %s not flund", cfg.Root[i])
							continue
						}

						xerror.Panic(filepath.Walk(cfg.Root[i], func(path string, info fs.FileInfo, err error) error {
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
						var base = fmt.Sprintf("protoc -I %s -I %s", cfg.ProtoPath, pwd)
						for i := range cfg.Includes {
							base += fmt.Sprintf(" -I %s", cfg.Includes[i])
						}
						var retagOut = ""
						var retagOpt = ""
						for i := range cfg.Plugins {
							var plg = cfg.Plugins[i]

							var name = plg.Name

							var out = func() string {
								// https://github.com/pseudomuto/protoc-gen-doc
								// 目录特殊处理
								if name == "doc" {
									var out = filepath.Join(plg.Out, in)
									xerror.Panic(pathutil.IsNotExistMkDir(out))
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
						fmt.Println(data + "\n")
						xerror.Panic(shutil.Shell(data).Run(), data)
						if retagOut != "" && retagOpt != "" {
							data = base + retagOut + retagOpt + " " + filepath.Join(in, "*.proto")
						}
						fmt.Println(data + "\n")
						xerror.Panic(shutil.Shell(data).Run(), data)
						return true
					})
					return nil
				},
			},
			&cli.Command{
				Name:  "vendor",
				Usage: "同步项目protobuf依赖到.proto中",
				Action: func(ctx *cli.Context) error {
					defer xerror.RespExit()

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
							xerror.Panic(err)

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

						if v == "" || pathutil.IsNotExist(fmt.Sprintf("%s@%s", url, v)) {
							fmt.Println("go", "get", "-d", url+"/...")
							xerror.Panic(shutil.Shell("go", "get", "-d", url+"/...").Run())

							// 再次解析go.mod然后获取版本信息
							versions = modutil.LoadVersions()
							v = versions[url]

							xerror.Assert(v == "", "%s version为空", url)
						}

						cfg.Depends[i].Version = v
					}
					xerror.Panic(ioutil.WriteFile(protoCfg, xerror.PanicBytes(yaml.Marshal(cfg)), 0755))

					// 删除老的protobuf文件
					_ = os.RemoveAll(cfg.ProtoPath)

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

						url = xerror.PanicStr(filepath.Abs(url))
						var newUrl = filepath.Join(cfg.ProtoPath, dep.Name)
						xerror.Panic(filepath.Walk(url, func(path string, info fs.FileInfo, err error) (gErr error) {
							if err != nil {
								return err
							}

							defer xerror.RespErr(&gErr)

							if info.IsDir() {
								return nil
							}

							if !strings.HasSuffix(info.Name(), ".proto") {
								return nil
							}

							var newPath = filepath.Join(newUrl, strings.TrimPrefix(path, url))
							xerror.Panic(pathutil.IsNotExistMkDir(filepath.Dir(newPath)))
							xerror.PanicErr(copyFile(newPath, path))

							return nil
						}))
					}
					return nil
				},
			},
		},
	}
	xerror.Exit(app.Run(os.Args))
}

func copyFile(dstFilePath string, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	xerror.Panic(err, "打开源文件错误", srcFilePath)

	dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0444)
	xerror.Panic(err, "打开目标文件错误", dstFilePath)

	return io.Copy(dstFile, srcFile)
}
