package protobuild

type Cfg struct {
	Checksum   string        `yaml:"checksum,omitempty" hash:"-"`
	Vendor     string        `yaml:"vendor,omitempty"`
	BasePlugin basePluginCfg `yaml:"base,omitempty" hash:"-"`
	Root       []string      `yaml:"root,omitempty" hash:"-"`
	Includes   []string      `yaml:"includes,omitempty" hash:"-"`
	Excludes   []string      `yaml:"excludes,omitempty" hash:"-"`
	Depends    []*depend     `yaml:"deps,omitempty"`
	Plugins    []*plugin     `yaml:"plugins,omitempty" hash:"-"`
	changed    bool
}

type basePluginCfg struct {
	Out    string `yaml:"out,omitempty"`
	Paths  string `yaml:"paths,omitempty"`
	Module string `yaml:"module,omitempty"`
}

type plugin struct {
	// Name protoc plugin name
	Name string `yaml:"name,omitempty"`

	// Path protoc plugin path
	Path string `yaml:"path,omitempty"`

	Out    string `yaml:"out,omitempty"`
	Shell  string `yaml:"shell,omitempty"`
	Docker string `yaml:"docker,omitempty"`
	Remote string `yaml:"remote,omitempty"`

	// SkipBase skip base config
	SkipBase bool `yaml:"skip_base,omitempty"`

	// ExcludeOpts exclude plugin opts
	ExcludeOpts pluginOpts `yaml:"exclude_opts,omitempty"`
	Opt         pluginOpts `yaml:"opt,omitempty"`
}

type depend struct {
	Name     string  `yaml:"name,omitempty"`
	Url      string  `yaml:"url,omitempty"`
	Path     string  `yaml:"path,omitempty"`
	Version  *string `yaml:"version,omitempty"`
	Optional *bool   `yaml:"optional,omitempty"`
}
