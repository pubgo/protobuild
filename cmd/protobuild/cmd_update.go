package protobuild

import (
	"context"
	"fmt"
	"strings"

	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/protobuild/internal/depresolver"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/redant"
	"golang.org/x/mod/semver"
)

// newUpdateCommand creates the update command: bumps dependency versions to
// the latest discoverable release, then re-runs the vendor flow.
func newUpdateCommand(dryRun *bool) *redant.Command {
	return &redant.Command{
		Use:   "update [name...]",
		Short: "升级依赖到最新发布版本并重新 vendor",
		Long: `升级依赖版本并重新 vendor。

支持自动发现版本的来源：
  git   通过 git ls-remote 取最新 semver tag
  gomod 通过 go list -m -versions 取最新版本
http/s3/gcs/local 来源无法自动发现版本，会被跳过。

示例:
  protobuild update               # 升级全部依赖
  protobuild update retag         # 只升级指定依赖
  protobuild update --dry-run     # 只显示可升级项`,
		Options: typex.Options{
			redant.Option{
				Flag:        "dry-run",
				Description: "只显示可升级项，不修改配置",
				Value:       redant.BoolOf(dryRun),
			},
		},
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			want := map[string]bool{}
			for _, name := range inv.Args {
				want[name] = true
			}

			fmt.Println("⬆️  检查依赖更新")
			changed := 0
			for _, dep := range globalCfg.Depends {
				if dep.Name == "" || dep.Url == "" {
					continue
				}
				if len(want) > 0 && !want[dep.Name] {
					continue
				}

				source := depresolver.DetectSource(dep.Url)
				latest, err := latestVersion(source, dep.Url)
				if err != nil {
					fmt.Printf("  ⚠️  %s: 无法发现最新版本 (%v)\n", dep.Name, err)
					continue
				}

				current := ""
				if dep.Version != nil {
					current = *dep.Version
				}
				if latest == current {
					fmt.Printf("  ✅ %s 已是最新 (%s)\n", dep.Name, current)
					continue
				}

				fmt.Printf("  ⬆️  %s: %s -> %s\n", dep.Name, displayVersion(current), latest)
				if *dryRun {
					continue
				}

				v := latest
				dep.Version = &v
				globalCfg.Changed = true
				changed++
			}

			if *dryRun {
				fmt.Println("\n🔍 Dry-run mode: 配置未修改。")
				return nil
			}

			if changed == 0 {
				fmt.Println("\n✨ 没有需要升级的依赖")
				return nil
			}

			fmt.Printf("\n📦 %d 个依赖已升级，重新 vendor...\n", changed)
			if err := saveConfig(); err != nil {
				return err
			}

			return runVendorSync(ctx, false, true)
		},
	}
}

// displayVersion renders an empty version as unset.
func displayVersion(v string) string {
	if v == "" {
		return "(未固定)"
	}
	return v
}

// latestVersion discovers the latest release version for a dependency.
func latestVersion(source depresolver.Source, url string) (string, error) {
	switch source {
	case depresolver.SourceGit:
		return latestGitTag(url)
	case depresolver.SourceGoMod:
		return latestGoModVersion(url)
	default:
		return "", fmt.Errorf("source %s 不支持自动发现版本", source.DisplayName())
	}
}

// latestGitTag returns the highest semver tag of a git repository.
func latestGitTag(url string) (string, error) {
	out, err := shutil.Run("git", "ls-remote", "--tags", "--sort=-v:refname", toGitURL(url))
	if err != nil {
		return "", fmt.Errorf("git ls-remote 失败: %w", err)
	}

	var tags []string
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		tag := strings.TrimSuffix(strings.TrimSpace(parts[1]), "^{}")
		tag = strings.TrimPrefix(tag, "refs/tags/")
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return pickLatestSemver(tags)
}

// latestGoModVersion returns the latest version of a Go module.
func latestGoModVersion(module string) (string, error) {
	out, err := shutil.Run("go", "list", "-m", "-versions", module)
	if err != nil {
		return "", fmt.Errorf("go list -m -versions 失败: %w", err)
	}

	fields := strings.Fields(out)
	if len(fields) < 2 {
		return "", fmt.Errorf("%s 没有可用版本", module)
	}

	return pickLatestSemver(fields[1:])
}

// toGitURL normalizes bare host/path URLs (e.g. github.com/x/y) to https.
func toGitURL(url string) string {
	u := strings.TrimSpace(url)
	u = strings.TrimPrefix(u, "git::")
	if strings.Contains(u, "://") || strings.HasPrefix(u, "git@") {
		return u
	}
	return "https://" + u
}

// pickLatestSemver picks the highest semver tag, preferring stable releases.
// Tags without semver shape are ignored.
func pickLatestSemver(tags []string) (string, error) {
	bestStable := ""
	bestAny := ""

	for _, tag := range tags {
		v := tag
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		if !semver.IsValid(v) {
			continue
		}

		if bestAny == "" || semver.Compare(v, "v"+strings.TrimPrefix(bestAny, "v")) > 0 {
			bestAny = tag
		}
		if semver.Prerelease(v) == "" {
			if bestStable == "" || semver.Compare(v, "v"+strings.TrimPrefix(bestStable, "v")) > 0 {
				bestStable = tag
			}
		}
	}

	if bestStable != "" {
		return bestStable, nil
	}
	if bestAny != "" {
		return bestAny, nil
	}
	return "", fmt.Errorf("未找到 semver 版本")
}
