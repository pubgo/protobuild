package protobuild

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pubgo/funk/v2/pathutil"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/redant"
)

// safeGitRef guards refs used in shell pipelines (git archive | tar).
var safeGitRef = regexp.MustCompile(`^[A-Za-z0-9/_.-]+$`)

// newBreakingCommand creates the breaking command: a buf breaking wrapper
// tailored to protobuild's workspace layout.
//
// Vendored deps (.proto) are gitignored, so comparing against a git ref
// directly would fail to compile the against side. Instead the command
// materializes a temporary base tree: git archive of the base ref plus the
// current vendor directory, and uses it as the against input. Dep updates
// therefore don't create noise — only your proto files are checked.
func newBreakingCommand() *redant.Command {
	var against string
	var paths, excludePaths []string
	var errorFormat string

	return &redant.Command{
		Use:   "breaking",
		Short: "检测 proto 兼容性（buf breaking 包装，自动排除 vendor）",
		Long: `基于 buf breaking 检测 proto 接口兼容性。

against 输入默认自动检测 base 分支（origin/main -> master -> main -> master），
并物化为临时目录：base 分支的 tracked 文件 + 当前 vendor 目录，
因此依赖升级不会产生 breaking 噪音，只有你自己的 proto 变更会被检查。

示例:
  protobuild breaking                          # 对比 origin/main
  protobuild breaking --against main           # 对比本地 main
  protobuild breaking --against .git#tag=v1.0.0
  protobuild breaking --error-format github-actions`,
		Options: typex.Options{
			redant.Option{
				Flag:        "against",
				Description: "buf against 输入，如 .git#branch=main；缺省自动物化 base 分支临时目录",
				Value:       redant.StringOf(&against),
			},
			redant.Option{
				Flag:        "path",
				Description: "只检查指定文件/目录（可多次）",
				Value:       redant.StringArrayOf(&paths),
			},
			redant.Option{
				Flag:        "exclude-path",
				Description: "排除指定文件/目录（可多次）",
				Value:       redant.StringArrayOf(&excludePaths),
			},
			redant.Option{
				Flag:        "error-format",
				Description: "输出格式 [text,json,msvs,junit,github-actions,gitlab-code-quality]",
				Default:     "text",
				Value:       redant.StringOf(&errorFormat),
			},
		},
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			if _, err := shutil.Run("buf", "--version"); err != nil {
				return fmt.Errorf("buf 未安装: go install github.com/bufbuild/buf/cmd/buf@latest")
			}

			if !pathutil.IsExist(filepath.Join(pwd, bufCfg)) {
				return fmt.Errorf("%s 不存在，buf breaking 需要它解析 import；请先运行 'protobuild buf'", bufCfg)
			}

			args := []string{"buf", "breaking"}

			againstInput, cleanup, err := resolveAgainstInput(against)
			if err != nil {
				return err
			}
			defer cleanup()
			args = append(args, "--against", againstInput)

			// no --path needed: the materialized base tree carries the same
			// vendor content as the working tree, so identical vendored files
			// produce no differences and only your protos are compared
			for _, p := range paths {
				args = append(args, "--path", p)
			}
			for _, p := range excludePaths {
				args = append(args, "--exclude-path", p)
			}
			args = append(args, "--error-format", errorFormat)

			fmt.Printf("🔎 运行 buf breaking (against: %s)\n", againstInput)
			if err := shutil.Shell(args...).Run(); err != nil {
				return fmt.Errorf("检测到 breaking changes（详见上方 buf 输出）")
			}

			fmt.Println("✅ 无 breaking changes")
			return nil
		},
	}
}

// resolveAgainstInput resolves the buf against input. When the user provides
// one it is used verbatim; otherwise the base branch is materialized into a
// temp directory (tracked files + current vendor). The returned cleanup
// removes the temp dir when necessary.
func resolveAgainstInput(against string) (string, func(), error) {
	noop := func() {}

	if strings.TrimSpace(against) != "" {
		return against, noop, nil
	}

	ref := detectBaseRef()
	if ref == "" {
		return "", noop, fmt.Errorf("无法自动检测 base 分支，请通过 --against 指定（buf input 格式，如 .git#branch=main）")
	}

	return materializeBaseTree(ref)
}

// detectBaseRef finds the most likely base branch as a locally resolvable ref.
func detectBaseRef() string {
	for _, ref := range []string{
		"refs/remotes/origin/main", "refs/remotes/origin/master",
		"refs/heads/main", "refs/heads/master",
	} {
		if out, err := shutil.Run("git", "rev-parse", "--verify", "--quiet", ref); err == nil && out != "" {
			return ref
		}
	}
	return ""
}

// materializeBaseTree creates a temp directory containing the base ref's
// tracked files plus the current vendor directory and the rendered buf.yaml,
// producing a compilable against input for buf breaking.
func materializeBaseTree(ref string) (string, func(), error) {
	noop := func() {}
	if !safeGitRef.MatchString(ref) {
		return "", noop, fmt.Errorf("非法的 git ref: %q", ref)
	}

	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("protobuild-breaking-%d", os.Getpid()))
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return "", noop, err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }

	// export the base tree (respects .gitattributes export-ignore)
	export := shutil.Shell("git", "archive", ref, "|", "tar", "-x", "-C", tmp)
	if err := export.Run(); err != nil {
		cleanup()
		return "", noop, fmt.Errorf("导出 %s 失败: %w", ref, err)
	}

	// inject the current vendor so imports resolve on the base side too
	if globalCfg.Vendor != "" && pathutil.IsExist(globalCfg.Vendor) {
		if err := copyDir(globalCfg.Vendor, filepath.Join(tmp, globalCfg.Vendor)); err != nil {
			cleanup()
			return "", noop, fmt.Errorf("注入 vendor 失败: %w", err)
		}
	}

	// ensure the base side uses the same workspace layout
	out, err := renderBufConfig(&globalCfg)
	if err != nil {
		cleanup()
		return "", noop, err
	}
	if err := os.WriteFile(filepath.Join(tmp, bufCfg), out, 0o644); err != nil {
		cleanup()
		return "", noop, err
	}

	return tmp, cleanup, nil
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(target, data, 0o644)
	})
}
