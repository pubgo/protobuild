package cmd

type Cfg struct {
	Version  string   `yaml:"version,omitempty" hash:"-"`
	Checksum string   `yaml:"checksum,omitempty" hash:"-"`
	BaseDir  string   `yaml:"base_dir" hash:"-"`
	Root     []string `yaml:"root,omitempty" hash:"-"`
	Includes []string `yaml:"includes,omitempty" hash:"-"`
	Plugins  []plugin `yaml:"plugins,omitempty" hash:"-"`
	Depends  []depend `yaml:"deps,omitempty"`
	Vendor   string   `yaml:"vendor,omitempty"`
	changed  bool
}

type plugin struct {
	Name string      `yaml:"name,omitempty"`
	Path string      `yaml:"path,omitempty"`
	Out  string      `yaml:"out,omitempty"`
	Opt  interface{} `yaml:"opt,omitempty"`
}

type depend struct {
	Name    string `yaml:"name,omitempty"`
	Url     string `yaml:"url,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Version string `yaml:"version,omitempty"`
}
