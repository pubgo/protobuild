package cmd

import (
	"github.com/pubgo/funk/errors"
	"gopkg.in/yaml.v3"
)

var _ yaml.Unmarshaler = (*pluginOpts)(nil)

type pluginOpts []string

func (p *pluginOpts) UnmarshalYAML(value *yaml.Node) error {
	if value.IsZero() {
		return nil
	}

	switch value.Kind {
	case yaml.ScalarNode:
		if value.Value != "" {
			*p = []string{value.Value}
			return nil
		}
		return nil
	case yaml.SequenceNode:
		var data []string
		if err := value.Decode(&data); err != nil {
			return err
		}
		*p = data
		return nil
	default:
		return errors.New("yaml kind type error, data=%s", value.Value)
	}
}

type Cfg struct {
	Checksum   string         `yaml:"checksum,omitempty" hash:"-"`
	Vendor     string         `yaml:"vendor,omitempty"`
	BasePlugin *basePluginCfg `yaml:"base,omitempty" hash:"-"`
	Root       []string       `yaml:"root,omitempty" hash:"-"`
	Includes   []string       `yaml:"includes,omitempty" hash:"-"`
	Excludes   []string       `yaml:"excludes,omitempty" hash:"-"`
	Depends    []depend       `yaml:"deps,omitempty"`
	Plugins    []plugin       `yaml:"plugins,omitempty" hash:"-"`
	changed    bool
}

type basePluginCfg struct {
	Out string `yaml:"out,omitempty"`
	Opt string `yaml:"opt,omitempty"`
}

type plugin struct {
	Name   string     `yaml:"name,omitempty"`
	Path   string     `yaml:"path,omitempty"`
	Out    string     `yaml:"out,omitempty"`
	Shell  string     `yaml:"shell,omitempty"`
	Docker string     `yaml:"docker,omitempty"`
	Remote string     `yaml:"remote,omitempty"`
	Opt    pluginOpts `yaml:"opt,omitempty"`
}

type depend struct {
	Name    string `yaml:"name,omitempty"`
	Url     string `yaml:"url,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Version string `yaml:"version,omitempty"`
}
