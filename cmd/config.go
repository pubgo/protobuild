package cmd

import (
	"github.com/pubgo/funk/errors"
	"gopkg.in/yaml.v2"
)

var _ yaml.Unmarshaler = (*pluginOpts)(nil)

type pluginOpts []string

func (p *pluginOpts) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dt interface{}
	if err := unmarshal(&dt); err != nil {
		return err
	}

	switch _dt := dt.(type) {
	case string:
		if _dt != "" {
			*p = []string{_dt}
			return nil
		}
		return nil
	case []string:
		*p = _dt
		return nil
	case []interface{}:
		var dtList []string
		for i := range _dt {
			dtList = append(dtList, _dt[i].(string))
		}
		*p = dtList
		return nil
	default:
		return errors.New("yaml kind type error, data=%#v", dt)
	}
}

type Cfg struct {
	Checksum      string   `yaml:"checksum,omitempty" hash:"-"`
	Vendor        string   `yaml:"vendor,omitempty"`
	BasePluginOut plugin   `yaml:"base_plugin_out" hash:"-"`
	Root          []string `yaml:"root,omitempty" hash:"-"`
	Includes      []string `yaml:"includes,omitempty" hash:"-"`
	Plugins       []plugin `yaml:"plugins,omitempty" hash:"-"`
	Depends       []depend `yaml:"deps,omitempty"`
	changed       bool
}

type plugin struct {
	Name string     `yaml:"name,omitempty"`
	Path string     `yaml:"path,omitempty"`
	Out  string     `yaml:"out,omitempty"`
	Opt  pluginOpts `yaml:"opt,omitempty"`
}

type depend struct {
	Name    string `yaml:"name,omitempty"`
	Url     string `yaml:"url,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Version string `yaml:"version,omitempty"`
}
