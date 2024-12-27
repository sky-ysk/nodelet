package app

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/component-base/version"
	"hit.edu/framework/pkg/nodelet"
	"hit.edu/framework/pkg/server"
)

const NodeletName = "nodelet"

func NewNodeletCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "nodelet",
		Long: `节点资源管理`,
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

	nl, err := Setup(ctx)
	if err != nil {
		return err
	}
	return Run(ctx, nl)
}

func Run(ctx context.Context, nl *nodelet.Nodelet) error {
	logs.Init(NodeletName)
	logs.Info("Starting Resourcelet\t", "version\t", version.Get())

	nl.Run(ctx)
	logs.Error("Failed to start resourcelet")
	return fmt.Errorf("")
}

func Setup(ctx context.Context) (*nodelet.Nodelet, error) {
	nl, err := nodelet.New(ctx)
	return nl, err
}
