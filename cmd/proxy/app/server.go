package app

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/component-base/version"
	"hit.edu/framework/pkg/proxy"
	"hit.edu/framework/pkg/server"
)

const ProxyName = "Proxy"

func NewProxyCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "api-proxy",
		Long: `接口代理`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(cmd)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	return cmd
}

func runCommand(cmd *cobra.Command) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		stopCh := server.SetupSignalHandler()
		<-stopCh
		cancel()
	}()

	r, err := Setup(ctx)
	if err != nil {
		return err
	}
	return Run(ctx, r)
}

func Run(ctx context.Context, p *proxy.Proxy) error {
	// 初始化日志模块
	logs.Init(ProxyName)
	logs.Infof("Starting Proxy, version %s", version.Get())

	p.Run(ctx)

	logs.Error("Failed to start Proxy")
	return fmt.Errorf("")
}

func Setup(ctx context.Context) (*proxy.Proxy, error) {
	// 创建配置
	c := proxy.NewConfig()
	//
	p, err := proxy.NewProxy(c)
	return p, err
}
