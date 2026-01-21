// Package formatcmd provides the format command for CLI.
package formatcmd

import (
	"context"
	"log/slog"
	"os/exec"

	"github.com/pubgo/redant"

	"github.com/pubgo/protobuild/internal/shutil"
)

// New creates a new format command.
func New(name string) *redant.Command {
	return &redant.Command{
		Use:   name,
		Short: "Format Protobuf files",
		Handler: func(_ context.Context, _ *redant.Invocation) error {
			bufPath, err := exec.LookPath("buf")
			if err != nil {
				slog.Info("buf not found, please install https://github.com/bufbuild/buf/releases")
				return err
			}

			shutil.MustRun(bufPath, "format proto -w")
			return nil
		},
	}
}
