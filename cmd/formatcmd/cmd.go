package formatcmd

import (
	"context"
	"log/slog"
	"os/exec"

	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/urfave/cli/v3"
)

func New(name string) *cli.Command {
	return &cli.Command{
		Name:  name,
		Usage: "Format Protobuf files",
		Action: func(ctx context.Context, command *cli.Command) error {
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
