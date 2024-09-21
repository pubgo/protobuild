package protobuild

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a8m/envsubst"
	"github.com/cnf/structhash"
	_ "github.com/deckarep/golang-set/v2"
	"github.com/huandu/go-clone"
	_ "github.com/huandu/go-clone"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/strutil"
	"github.com/pubgo/protobuild/internal/modutil"
	"gopkg.in/yaml.v3"
)

func mergePluginConfig(base *Config, pluginConfigs ...*Config) *Config {
	base = clone.Clone(base).(*Config)
	for _, cfg := range pluginConfigs {
		if cfg == nil {
			continue
		}

		if cfg.BasePlugin != nil {
			base.BasePlugin = cfg.BasePlugin
		}

		if len(cfg.Root) > 0 {
			base.Root = cfg.Root
		}

		if len(cfg.Includes) > 0 {
			base.Includes = append(base.Includes, cfg.Includes...)
		}

		if len(cfg.Excludes) > 0 {
			base.Excludes = cfg.Excludes
		}

		if len(cfg.Plugins) > 0 {
			base.Plugins = cfg.Plugins
		}
	}

	if base.BasePlugin == nil {
		base.BasePlugin = &basePluginCfg{}
	}
	return base
}

func parseConfig() error {
	content := assert.Must1(os.ReadFile(protoCfg))
	content = assert.Must1(envsubst.Bytes(content))
	assert.Must(yaml.Unmarshal(content, &globalCfg))

	globalCfg.Vendor = strutil.FirstFnNotEmpty(
		func() string {
			return globalCfg.Vendor
		},
		func() string {
			protoPath := filepath.Join(pwd, ".proto")
			if pathutil.IsExist(protoPath) {
				return protoPath
			}
			return ""
		},
		func() string {
			goModPath := filepath.Dir(modutil.GoModPath())
			if goModPath == "" {
				panic("没有找到项目go.mod文件")
			}

			return filepath.Join(goModPath, ".proto")
		},
	)

	assert.Must(pathutil.IsNotExistMkDir(globalCfg.Vendor))

	// protobuf文件检查
	for _, dep := range globalCfg.Depends {
		assert.If(dep.Name == "" || dep.Url == "", "name和url都不能为空")
	}

	checksum := fmt.Sprintf("%x", structhash.Sha1(globalCfg, 1))
	if globalCfg.Checksum != checksum {
		globalCfg.Checksum = checksum
		globalCfg.changed = true
	}

	return nil
}

func parsePluginConfig(path string) (cfg *Config) {
	content := assert.Must1(os.ReadFile(path))
	content = assert.Must1(envsubst.Bytes(content))
	assert.Must(yaml.Unmarshal(content, &cfg))
	return
}
