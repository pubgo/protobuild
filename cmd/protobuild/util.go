package protobuild

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/a8m/envsubst"
	"github.com/cnf/structhash"
	"github.com/huandu/go-clone"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
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

	oldChecksum, err := getChecksumData(globalCfg.Vendor)
	if err != nil {
		slog.Warn("failed to get checksum data", slog.Any("err", err.Error()))
	}
	if oldChecksum != checksum {
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

var checkSumPath = func(vendorPath string) string {
	return filepath.Join(vendorPath, "checksum")
}

func getChecksumData(vendorPath string) (string, error) {
	var path = checkSumPath(vendorPath)
	if pathutil.IsNotExist(vendorPath) {
		return "", errors.NewFmt("file not found")
	}

	if pathutil.IsNotExist(path) {
		return "", errors.NewFmt("file not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", errors.WrapCaller(err)
	}
	return string(data), nil
}

func writeChecksumData(vendorPath string, data []byte) error {
	_ = os.MkdirAll(vendorPath, 0755)
	var path = checkSumPath(vendorPath)
	return errors.WrapCaller(os.WriteFile(path, data, 0644))
}
