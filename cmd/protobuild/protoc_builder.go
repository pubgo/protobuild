package protobuild

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/samber/lo"
)

// ProtocBuilder builds protoc commands for code generation.
type ProtocBuilder struct {
	includes []string
	vendor   string
	pwd      string
}

// NewProtocBuilder creates a new ProtocBuilder.
func NewProtocBuilder(includes []string, vendor, pwd string) *ProtocBuilder {
	return &ProtocBuilder{
		includes: includes,
		vendor:   vendor,
		pwd:      pwd,
	}
}

// BuildCommand builds a protoc command for the given config and proto path.
func (b *ProtocBuilder) BuildCommand(cfg *Config, protoPath string) *ProtocCommand {
	return &ProtocCommand{
		cfg:       cfg,
		protoPath: protoPath,
		includes:  lo.Uniq(append(b.includes, cfg.Includes...)),
		vendor:    b.vendor,
		pwd:       b.pwd,
	}
}

// ProtocCommand represents a protoc command to execute.
type ProtocCommand struct {
	cfg       *Config
	protoPath string
	includes  []string
	vendor    string
	pwd       string
}

// Execute runs the protoc command.
func (c *ProtocCommand) Execute() error {
	mainCmd, retagCmd := c.build()

	logger := log.GetLogger("protobuild")
	logger.Info().Msg(mainCmd)

	if err := shutil.Shell(mainCmd).Run(); err != nil {
		return fmt.Errorf("protoc failed: %w", err)
	}

	// Run retag plugin separately if configured
	if retagCmd != "" {
		logger.Info().Bool(reTagPluginName, true).Msg(retagCmd)
		if err := shutil.Shell(retagCmd).Run(); err != nil {
			return fmt.Errorf("retag failed: %w", err)
		}
	}

	return nil
}

// build constructs the protoc command strings.
func (c *ProtocCommand) build() (mainCmd, retagCmd string) {
	includes := lo.Uniq(append(c.includes, c.vendor, c.pwd))

	// Build base command with includes
	var base strings.Builder
	base.WriteString("protoc")
	for _, inc := range includes {
		base.WriteString(fmt.Sprintf(" -I %s", inc))
	}

	var pluginArgs strings.Builder
	var retagArgs strings.Builder

	for _, plg := range c.cfg.Plugins {
		if plg.SkipRun {
			continue
		}

		args := c.buildPluginArgs(plg)

		if plg.Name == reTagPluginName {
			retagArgs.WriteString(args)
		} else {
			pluginArgs.WriteString(args)
		}
	}

	protoFiles := filepath.Join(c.protoPath, "*.proto")

	mainCmd = base.String() + pluginArgs.String() + " " + protoFiles
	if retagArgs.Len() > 0 {
		retagCmd = base.String() + retagArgs.String() + " " + protoFiles
	}

	return
}

// buildPluginArgs builds command arguments for a single plugin.
func (c *ProtocCommand) buildPluginArgs(plg *plugin) string {
	var args strings.Builder
	name := plg.Name

	// Plugin path
	if plg.Path != "" {
		plgPath := assert.Must1(exec.LookPath(plg.Path))
		assert.If(pathutil.IsNotExist(plgPath), "plugin path not found: %s", plgPath)
		args.WriteString(fmt.Sprintf(" --plugin=protoc-gen-%s=%s", name, plgPath))
	}

	// Output directory
	out := c.resolveOutputDir(plg)
	assert.Exit(pathutil.IsNotExistMkDir(out))

	// Build options
	opts := c.buildPluginOpts(plg, out)

	// Handle wrapper plugins (shell/docker)
	if plg.Shell != "" || plg.Docker != "" {
		opts = append(opts, "__wrapper="+name)
		wrapperPath := assert.Must1(exec.LookPath("protobuild"))
		args.WriteString(fmt.Sprintf(" --plugin=protoc-gen-%s=%s", name, wrapperPath))
	}

	// Handle retag plugin specially - run after main compilation to modify generated files
	if name == reTagPluginName {
		args.WriteString(fmt.Sprintf(" --%s_out=%s", name, out))
		if len(opts) > 0 {
			filteredOpts := c.filterExcludedOpts(opts, plg.ExcludeOpts)
			if len(filteredOpts) > 0 {
				args.WriteString(fmt.Sprintf(" --%s_opt=%s", name, strings.Join(filteredOpts, ",")))
			}
		}
		return args.String()
	}

	// Output
	args.WriteString(fmt.Sprintf(" --%s_out=%s", name, out))

	// Options
	if len(opts) > 0 {
		filteredOpts := c.filterExcludedOpts(opts, plg.ExcludeOpts)
		if len(filteredOpts) > 0 {
			args.WriteString(fmt.Sprintf(" --%s_opt=%s", name, strings.Join(filteredOpts, ",")))
		}
	}

	return args.String()
}

// resolveOutputDir determines the output directory for a plugin.
func (c *ProtocCommand) resolveOutputDir(plg *plugin) string {
	// Special handling for doc plugin
	if plg.Name == "doc" {
		out := filepath.Join(plg.Out, c.protoPath)
		assert.Must(pathutil.IsNotExistMkDir(out))
		return out
	}

	if plg.Out != "" {
		return plg.Out
	}

	if c.cfg.BasePlugin != nil && c.cfg.BasePlugin.Out != "" {
		return c.cfg.BasePlugin.Out
	}

	return "."
}

// buildPluginOpts builds the options for a plugin.
func (c *ProtocCommand) buildPluginOpts(plg *plugin, out string) []string {
	opts := append(plg.Opt, plg.Opts...)

	// Add base paths option if not set
	hasPath := lo.ContainsBy(opts, func(opt string) bool {
		return strings.HasPrefix(opt, "paths=")
	})
	if !hasPath && c.cfg.BasePlugin != nil && c.cfg.BasePlugin.Paths != "" && !plg.SkipBase {
		opts = append(opts, fmt.Sprintf("paths=%s", c.cfg.BasePlugin.Paths))
	}

	// Add base module option if not set
	hasModule := lo.ContainsBy(opts, func(opt string) bool {
		return strings.HasPrefix(opt, "module=")
	})
	if !hasModule && c.cfg.BasePlugin != nil && c.cfg.BasePlugin.Module != "" && !plg.SkipBase {
		opts = append(opts, fmt.Sprintf("module=%s", c.cfg.BasePlugin.Module))
	}

	return opts
}

// filterExcludedOpts filters out excluded options.
func (c *ProtocCommand) filterExcludedOpts(opts []string, excludes pluginOpts) []string {
	return lo.Filter(opts, func(opt string, _ int) bool {
		return !lo.ContainsBy(excludes, func(ex string) bool {
			return strings.HasPrefix(opt, ex)
		})
	})
}
