package protobuild

import (
	"context"
	"fmt"

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

			fmt.Printf("🔎 校验 vendor 目录: %s\n", globalCfg.Vendor)
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
		},
	}
}
