package protobuild

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
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

	defaultSkillContent = `---
name: protobuild
description: "Manage protobuf projects: vendor dependencies, generate code, lint/format. Use only when protobuf.yaml is present in the project root."
license: MIT
metadata:
	author: pubgo
	version: "1.0.0"
compatibility: "Requires protobuild CLI, protoc, plugins; run in repo root with protobuf.yaml."
allowed-tools: "Bash(protobuild:*)"
---

## When to use
- 项目存在 ` + "`protobuf.yaml`" + `，需要下载依赖、生成代码、lint/format proto。
- 先 vendor 再 gen；只检查格式/规范，使用 lint 或 format --exit-code。

## Safety / Constraints
- 会写盘的命令：gen、format -w、install（需告知用户）。
- 必须在包含 protobuf.yaml 的项目根目录执行。
- 如未 vendor 过依赖，先运行 protobuild vendor。

## Inputs
- command: one of [gen, vendor, lint, format, clean, deps, install, doctor, web]
- args: optional string list, e.g. ["-w", "--diff"]
- working_dir: project root (defaults to current CWD if已在项目根)

## Steps
1) 确认 working_dir 内存在 protobuf.yaml，否则提示用户补全。
2) 如 command=gen 且依赖未就绪，先运行 protobuild vendor。
3) 执行：protobuild <command> <args...>（在 working_dir）。
4) 收集 stdout/stderr/exit_code，若命令会写盘（gen/format -w/install），提示用户查看变更。
5) 失败时给出 stderr 摘要与下一步建议（如检查路径/依赖/插件）。

## Examples
- protobuild vendor
- protobuild gen
- protobuild format --exit-code
- protobuild lint
`

	defaultOpenAIContent = `interface:
	display_name: "Protobuild"
	short_description: "Proto build/lint/format tool"
	icon_small: "./assets/protobuild-16.png"
	icon_large: "./assets/protobuild-64.png"
	brand_color: "#0F9D58"
	default_prompt: "Use protobuild to vendor deps and generate proto code."

dependencies:
	tools: []
`
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
			newSkillsCommand(),
			webcmd.New(&protoCfg),
			newVersionCommand(),
			upgradecmd.New("pubgo", "protobuild"),
		},
	}

	return app
}

// newSkillsCommand installs the Agent Skills template into the current repository.
func newSkillsCommand() *redant.Command {
	var outDir = ".agents/skills/protobuild"
	var withOpenAI bool
	var force bool

	return &redant.Command{
		Use:   "skills",
		Short: "install Agent Skills template (.agents/skills/protobuild)",
		Options: typex.Options{
			redant.Option{
				Flag:        "path",
				Shorthand:   "p",
				Description: "skills output directory",
				Value:       redant.StringOf(&outDir),
			},
			redant.Option{
				Flag:        "openai",
				Description: "also generate agents/openai.yaml",
				Value:       redant.BoolOf(&withOpenAI),
			},
			redant.Option{
				Flag:        "force",
				Shorthand:   "f",
				Description: "overwrite existing files",
				Value:       redant.BoolOf(&force),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			skillPath := filepath.Join(outDir, "SKILL.md")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}

			if err := writeFileIfNeeded(skillPath, []byte(defaultSkillContent), force); err != nil {
				return err
			}

			if withOpenAI {
				openaiDir := filepath.Join(outDir, "agents")
				if err := os.MkdirAll(openaiDir, 0o755); err != nil {
					return err
				}
				openaiPath := filepath.Join(openaiDir, "openai.yaml")
				if err := writeFileIfNeeded(openaiPath, []byte(defaultOpenAIContent), force); err != nil {
					return err
				}
			}

			fmt.Printf("Skill installed at %s\n", skillPath)
			if withOpenAI {
				fmt.Printf("OpenAI metadata at %s\n", filepath.Join(outDir, "agents", "openai.yaml"))
			}
			fmt.Println("You can validate with: skills-ref validate", outDir)
			return nil
		},
	}
}

func writeFileIfNeeded(path string, data []byte, force bool) error {
	if _, err := os.Stat(path); err == nil {
		if !force {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return os.WriteFile(path, data, 0o644)
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
