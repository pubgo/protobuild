package protobuild

import (
	"context"
	"fmt"
	"os"

	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/redant"
)

// newVerifyCommand creates the verify command.
// It checks the vendor directory against the manifest written by
// 'protobuild vendor', so corrupted or hand-edited vendored files are caught
// before they break builds. Designed to run as a CI gate.
func newVerifyCommand() *redant.Command {
	return &redant.Command{
		Use:        "verify",
		Short:      "校验 vendor 目录内容与 manifest 一致（CI 门禁）",
		Middleware: withParseConfig(),
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()
			return runVerify()
		},
	}
}

// runVerify implements the verify command against the parsed global config.
// A project with no configured deps and an empty (or absent) vendor dir has
// an empty set to validate — that passes: the new init templates ship zero
// deps, and a CI gate must not fail forever on "manifest missing" for them.
// Any stray vendor content without a manifest still fails (no blessing of
// hand-placed files).
func runVerify() error {
	fmt.Printf("🔎 校验 vendor 目录: %s\n", globalCfg.Vendor)

	if len(globalCfg.Depends) == 0 && vendorDirEmpty(globalCfg.Vendor) {
		fmt.Println("✅ 未配置依赖且 vendor 为空，校验通过（空集）")
		return nil
	}

	problems, stale, err := verifyVendor(globalCfg.Vendor, globalCfg.Checksum)
	if err != nil {
		return err
	}

	if stale {
		fmt.Println("⚠️  protobuf.yaml 已变更，vendor 可能过期，请运行 'protobuild vendor'")
	}

	if len(problems) == 0 {
		fmt.Println("✅ vendor 校验通过")
		return nil
	}

	fmt.Printf("❌ vendor 校验失败 (%d 个问题):\n", len(problems))
	for _, p := range problems {
		fmt.Printf("   • %s\n", p)
	}
	fmt.Println("\n💡 提示: 运行 'protobuild vendor' 重建 vendor 目录")
	return fmt.Errorf("vendor 校验失败")
}

// vendorDirEmpty reports whether the vendor directory holds nothing to verify:
// absent or no entries. Read errors on an existing dir count as non-empty so
// verification proceeds and surfaces them.
func vendorDirEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return os.IsNotExist(err)
	}
	return len(entries) == 0
}
