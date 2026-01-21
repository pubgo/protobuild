// Package config provides shared configuration types for protobuild.
package config

// Config represents the protobuild project configuration.
type Config struct {
	Checksum   string         `yaml:"checksum,omitempty" json:"checksum,omitempty" hash:"-"`
	Vendor     string         `yaml:"vendor,omitempty" json:"vendor"`
	BasePlugin *BasePluginCfg `yaml:"base,omitempty" json:"base,omitempty" hash:"-"`

	// Root path, default is proto path (source path)
	Root []string `yaml:"root,omitempty" json:"root" hash:"-"`

	// Includes protoc include path, default is proto path and .proto path
	Includes   []string  `yaml:"includes,omitempty" json:"includes" hash:"-"`
	Excludes   []string  `yaml:"excludes,omitempty" json:"excludes" hash:"-"`
	Depends    []*Depend `yaml:"deps,omitempty" json:"deps"`
	Plugins    []*Plugin `yaml:"plugins,omitempty" json:"plugins" hash:"-"`
	Installers []string  `yaml:"installers,omitempty" json:"installers" hash:"-"`
	Linter     *Linter   `yaml:"linter,omitempty" json:"linter,omitempty" hash:"-"`

	// Changed is used internally to track if config has been modified (lowercase for internal use)
	Changed bool `yaml:"-" json:"-"`
}

// BasePluginCfg represents base plugin configuration applied to all plugins.
type BasePluginCfg struct {
	Out    string `yaml:"out,omitempty" json:"out"`
	Paths  string `yaml:"paths,omitempty" json:"paths"`
	Module string `yaml:"module,omitempty" json:"module"`
}

// Plugin represents a protoc plugin configuration.
type Plugin struct {
	// Name protoc plugin name (used as protoc-gen-{name})
	Name string `yaml:"name,omitempty" json:"name"`

	// Path custom plugin binary path
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Out output directory
	Out string `yaml:"out,omitempty" json:"out,omitempty"`

	// Shell run via shell command
	Shell string `yaml:"shell,omitempty" json:"shell,omitempty"`

	// Docker run via Docker container
	Docker string `yaml:"docker,omitempty" json:"docker,omitempty"`

	// Remote remote plugin URL
	Remote string `yaml:"remote,omitempty" json:"remote,omitempty"`

	// SkipBase skip base config
	SkipBase bool `yaml:"skip_base,omitempty" json:"skip_base,omitempty"`

	// SkipRun skip run plugin
	SkipRun bool `yaml:"skip_run,omitempty" json:"skip_run,omitempty"`

	// ExcludeOpts options to exclude
	ExcludeOpts PluginOpts `yaml:"exclude_opts,omitempty" json:"exclude_opts,omitempty"`

	// Opt plugin options
	Opt PluginOpts `yaml:"opt,omitempty" json:"opt,omitempty"`

	// Opts alias for Opt
	Opts PluginOpts `yaml:"opts,omitempty" json:"opts,omitempty"`
}

// Depend represents a proto dependency.
type Depend struct {
	// Name local name/path in vendor directory
	Name string `yaml:"name,omitempty" json:"name"`

	// Source type: gomod(default), git, http, s3, gcs, local
	Source string `yaml:"source,omitempty" json:"source,omitempty"`

	// Url source URL
	Url string `yaml:"url,omitempty" json:"url"`

	// Path subdirectory within the source
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Version specific version (for Go modules)
	Version *string `yaml:"version,omitempty" json:"version,omitempty"`

	// Ref git ref (branch, tag, commit) for Git sources
	Ref string `yaml:"ref,omitempty" json:"ref,omitempty"`

	// Optional skip if not found
	Optional *bool `yaml:"optional,omitempty" json:"optional,omitempty"`
}

// Linter represents linter configuration.
type Linter struct {
	Rules                     *LinterRules `yaml:"rules,omitempty" json:"rules,omitempty" hash:"-"`
	FormatType                string       `yaml:"format_type,omitempty" json:"format_type,omitempty"`
	IgnoreCommentDisablesFlag bool         `yaml:"ignore_comment_disables_flag,omitempty" json:"ignore_comment_disables_flag,omitempty"`
}

// LinterRules represents linter rules configuration.
type LinterRules struct {
	EnabledRules  []string `yaml:"enabled_rules,omitempty" json:"enabled_rules,omitempty"`
	DisabledRules []string `yaml:"disabled_rules,omitempty" json:"disabled_rules,omitempty"`
}

// GetVersion returns the version string or empty if nil.
func (d *Depend) GetVersion() string {
	if d.Version == nil {
		return ""
	}
	return *d.Version
}

// IsOptional returns true if the dependency is optional.
func (d *Depend) IsOptional() bool {
	if d.Optional == nil {
		return false
	}
	return *d.Optional
}

// GetAllOpts returns combined Opt and Opts.
func (p *Plugin) GetAllOpts() []string {
	result := make([]string, 0, len(p.Opt)+len(p.Opts))
	result = append(result, p.Opt...)
	result = append(result, p.Opts...)
	return result
}
