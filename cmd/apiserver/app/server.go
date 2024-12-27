package app

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apiserver"
	"hit.edu/framework/pkg/apiserver/options"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/component-base/version"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	
	"github.com/spf13/cobra"
)

const APIServerName = "api-server"

func init() {
}

func NewAPIServerCommand() *cobra.Command {
	// 创建Option
	opt := options.NewOptions()
	
	cmd := &cobra.Command{
		Use:  "api-genericserver",
		Long: `接口服务器`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 参数验证
			if errs := opt.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}
			
			return runCommand(cmd, opt)
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
	// 添加Flags
	opt.AddFlags(cmd.Flags())
	return cmd
}

func runCommand(cmd *cobra.Command, opt *options.Options) error {
	
	// TODO: 根据Config修改Option,然后传入API Server中
	
	return Run(cmd.Context(), opt)
}

func Run(ctx context.Context, opts *options.Options) error {
	//logs.Info("Starting API Server\t", "version\t", version.Get())
	logs.Init(APIServerName)
	logs.Info("Starting API Server", zap.String("version", version.Get()))
	// 参数配置
	config := apiserver.NewConfig(opts)
	
	// 创建服务器
	server := apiserver.NewAPIServer(config)
	
	// PreRun
	prepared := server.PrepareRun()
	
	// Run
	return prepared.RunWithContext(ctx)
}
