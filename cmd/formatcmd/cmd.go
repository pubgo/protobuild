package formatcmd

import (
	"context"
	"log/slog"
	"os/exec"

	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/pubgo/redant"
)

func New(name string) *redant.Command {
	return &redant.Command{
		Use:   name,
		Short: "Format Protobuf files",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
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
