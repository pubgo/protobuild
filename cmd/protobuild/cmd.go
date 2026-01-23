package protobuild

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/cmds/upgradecmd"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/protobuild/cmd/formatcmd"
	"github.com/pubgo/protobuild/cmd/linters"
	"github.com/pubgo/protobuild/cmd/webcmd"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/redant"
	"github.com/samber/lo"
	"golang.org/x/term"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	globalCfg      Config
	protoCfg       = "protobuf.yaml"
	protoPluginCfg = "protobuf.plugin.yaml"
	pwd            = assert.Exit1(os.Getwd())
	logger         = log.GetLogger("protobuild")
)

const (
	reTagPluginName = "retag"
)

// withParseConfig returns a middleware that parses the config file.
func withParseConfig() redant.MiddlewareFunc {
	return func(next redant.HandlerFunc) redant.HandlerFunc {
		return func(ctx context.Context, inv *redant.Invocation) error {
			if err := parseConfig(); err != nil {
				slog.Error("failed to parse config", "err", err)
				return err
			}
			return next(ctx, inv)
		}
	}
}

// Main creates and returns the root CLI command with all subcommands.
func Main() *redant.Command {
	var force, update, dryRun bool
	cliArgs, options := linters.NewCli()

	app := &redant.Command{
		Use:   "protobuild",
		Short: "Protobuf generation, configuration and management tool",
		Options: typex.Options{
			redant.Option{
				Flag:        "conf",
				Shorthand:   "c",
				Description: "protobuf config path",
				Default:     protoCfg,
				Value:       redant.StringOf(&protoCfg),
			},
		},
		Handler: handleStdinPlugin,
		Children: typex.Commands{
			newInitCommand(),
			newDoctorCommand(),
			newGenCommand(),
			newVendorCommand(&force, &update),
			newInstallCommand(&force),
			newLintCommand(cliArgs, options),
			newFormatCommand(),
			newDepsCommand(),
			newCleanCommand(&dryRun),
			webcmd.New(&protoCfg),
			newVersionCommand(),
			upgradecmd.New("pubgo", "protobuild"),
		},
	}

	return app
}

// handleStdinPlugin handles protoc plugin mode when invoked via stdin.
func handleStdinPlugin(ctx context.Context, inv *redant.Invocation) error {
	file := os.Stdin
	if term.IsTerminal(int(file.Fd())) {
		return redant.DefaultHelpFn()(ctx, inv)
	}

	fi := assert.Exit1(file.Stat())
	if fi.Size() == 0 {
		return redant.DefaultHelpFn()(ctx, inv)
	}

	in := assert.Must1(io.ReadAll(file))
	req := &pluginpb.CodeGeneratorRequest{}
	assert.Must(proto.Unmarshal(in, req))

	var opts protogen.Options
	plg := assert.Must1(opts.New(req))
	for _, f := range plg.Files {
		if !f.Generate {
			continue
		}
		log.Printf("%s\n", f.GeneratedFilenamePrefix)
	}

	plgName, params := parsePluginParams(req.GetParameter())
	if len(params) > 0 {
		req.Parameter = lo.ToPtr(strings.Join(params, ","))
	}

	return executeWrapperPlugin(plgName, req)
}

// parsePluginParams extracts wrapper plugin name and remaining params.
func parsePluginParams(param string) (plgName string, params []string) {
	for _, p := range strings.Split(param, ",") {
		if strings.HasPrefix(p, "__wrapper") {
			names := strings.Split(p, "=")
			plgName = strings.TrimSpace(names[len(names)-1])
		} else {
			params = append(params, p)
		}
	}
	return
}

// executeWrapperPlugin executes shell or docker wrapper plugin.
func executeWrapperPlugin(plgName string, req *pluginpb.CodeGeneratorRequest) error {
	for _, p := range globalCfg.Plugins {
		if p.Name != plgName {
			continue
		}

		log.Printf("%#v\n", p)
		reqData := assert.Must1(proto.Marshal(req))

		if p.Shell != "" {
			cmd := shutil.Shell(strings.TrimSpace(p.Shell))
			cmd.Stdin = bytes.NewBuffer(reqData)
			return cmd.Run()
		}

		if p.Docker != "" {
			cmd := shutil.Shell("docker run -i --rm " + p.Docker)
			cmd.Stdin = bytes.NewBuffer(reqData)
			return cmd.Run()
		}
	}
	return nil
}

// newInstallCommand creates the install command.
func newInstallCommand(force *bool) *redant.Command {
	return &redant.Command{
		Use:   "install",
		Short: "install protobuf plugin",
		Options: typex.Options{
			redant.Option{
				Flag:        "force",
				Shorthand:   "f",
				Description: "force update protobuf plugin",
				Value:       redant.BoolOf(force),
			},
		},
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			for _, plg := range globalCfg.Installers {
				installPlugin(plg, *force)
			}
			return nil
		},
	}
}

// installPlugin installs a single protobuf plugin.
func installPlugin(plg string, force bool) {
	// Ensure @version suffix
	if !strings.Contains(plg, "@") {
		parts := strings.Split(plg, "@")
		plg = strings.Join(parts[:len(parts)-1], "@") + "@latest"
	}

	plgName := strings.Split(lo.LastOrEmpty(strings.Split(plg, "/")), "@")[0]
	path, err := exec.LookPath(plgName)

	if err != nil {
		slog.Error("command not found", slog.Any("name", plgName))
	}

	if err == nil && !globalCfg.Changed && !force {
		slog.Info("no changes", slog.Any("path", path))
		return
	}

	slog.Info("install command", slog.Any("name", plg))
	assert.Must(shutil.Shell("go", "install", plg).Run())
}

// newVersionCommand creates the version command.
func newVersionCommand() *redant.Command {
	return &redant.Command{
		Use:   "version",
		Short: "version info",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()
			fmt.Printf("Project:   %s\n", running.Project())
			fmt.Printf("Version:   %s\n", running.Version())
			fmt.Printf("GitCommit: %s\n", running.CommitID())
			return nil
		},
	}
}

// newFormatCommand creates the format command with project config integration.
func newFormatCommand() *redant.Command {
	cmd := formatcmd.New("format", func() *formatcmd.ProjectConfig {
		return &formatcmd.ProjectConfig{
			Root:     globalCfg.Root,
			Vendor:   globalCfg.Vendor,
			Includes: globalCfg.Includes,
		}
	})
	cmd.Middleware = withParseConfig()
	return cmd
}

// newLintCommand creates the lint command.
func newLintCommand(cliArgs *linters.CliArgs, options typex.Options) *redant.Command {
	return &redant.Command{
		Use:        "lint",
		Short:      "lint protobuf https://linter.aip.dev/rules/",
		Options:    options,
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			walker := NewProtoWalker(globalCfg.Root, globalCfg.Excludes)
			protoDirs := lo.Uniq(walker.GetAllProtoDirs())

			for _, dir := range protoDirs {
				protoFiles := walker.GetProtoFiles(dir)
				if len(protoFiles) == 0 {
					continue
				}

				includes := lo.Uniq(append(globalCfg.Includes, globalCfg.Vendor))
				linterCfg := toLinterConfig(globalCfg.Linter)
				if err := linters.Linter(cliArgs, linterCfg, includes, protoFiles); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
