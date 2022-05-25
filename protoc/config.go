package protoc

type Cfg struct {
	Version   string   `yaml:"version,omitempty" hash:"-"`
	ProtoPath string   `yaml:"proto-path,omitempty"`
	Checksum  string   `yaml:"checksum,omitempty" hash:"-"`
	Root      []string `yaml:"root,omitempty" hash:"-"`
	Depends   []depend `yaml:"deps,omitempty"`
	Plugins   []plugin `yaml:"plugins,omitempty" hash:"-"`
	Includes  []string `yaml:"includes,omitempty" hash:"-"`
	changed   bool
}

type plugin struct {
	Name string      `yaml:"name,omitempty"`
	Out  string      `yaml:"out,omitempty"`
	Opt  interface{} `yaml:"opt,omitempty"`
}

type depend struct {
	Name    string `yaml:"name,omitempty"`
	Url     string `yaml:"url,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Version string `yaml:"version,omitempty"`
}
