// Package webcmd provides CLI command for the web UI.
package webcmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/redant"
)

// New creates the web command.
func New(configPath *string) *redant.Command {
	var portStr string

	return &redant.Command{
		Use:   "web",
		Short: "启动 Web 配置管理界面",
		Long: `启动一个本地 Web 服务器，提供可视化的配置管理界面。

通过浏览器可以：
  - 查看和编辑项目配置
  - 管理 Proto 依赖
  - 配置 Protoc 插件
  - 执行构建、检查、格式化等操作
  - 实时预览 YAML 配置

示例:
  protobuild web
  protobuild web --port 9090`,
		Options: typex.Options{
			redant.Option{
				Flag:        "port",
				Shorthand:   "p",
				Description: "Web 服务器端口",
				Default:     "8080",
				Value:       redant.StringOf(&portStr),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			port, _ := strconv.Atoi(portStr)
			if port == 0 {
				port = 8080
			}

			server, err := NewServer(*configPath)
			if err != nil {
				return err
			}

			// Handle signals for graceful shutdown
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigCh
				cancel()
			}()

			return server.Start(ctx, port)
		},
	}
}
