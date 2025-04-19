package app

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/component-base/version"
	"hit.edu/framework/pkg/registry"
	"hit.edu/framework/pkg/server"
)

const RegistryName = "registry"

func NewRegistryCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "registry",
		Long: `资源仓库`,
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

func Run(ctx context.Context, r *registry.Registry) error {
	// 初始化日志模块
	logs.Init(RegistryName)
	logs.Infof("Starting Registry, version %s", version.Get())

	r.Run(ctx)

	logs.Error("Failed to start Registry")
	return fmt.Errorf("")
}

func Setup(ctx context.Context) (*registry.Registry, error) {
	// 创建配置
	c := registry.NewConfig()
	//
	r, err := registry.NewRegistry(c)
	return r, err
}
