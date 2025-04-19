package main

import (
	"context"
	"go.uber.org/zap"
	"hit.edu/framework/cmd/nodelet/app"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/server"
)

func main() {
	//从cmd/nodelet/app/server下的runCommand（）开始执行
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		stopCh := server.SetupSignalHandler()
		<-stopCh
		cancel()
	}()
	nl, err := app.Setup(ctx)
	if err != nil {
		logs.Error("fail to setup", zap.Error(err))
	}
	app.Run(ctx, nl)
}
