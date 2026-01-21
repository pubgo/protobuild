package protobuild

import (
	"github.com/pubgo/protobuild/internal/config"
)

// Type aliases for backward compatibility
type (
	Config        = config.Config
	basePluginCfg = config.BasePluginCfg
	plugin        = config.Plugin
	depend        = config.Depend
	pluginOpts    = config.PluginOpts
)
