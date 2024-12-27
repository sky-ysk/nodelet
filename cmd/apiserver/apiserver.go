package main

import (
	"hit.edu/framework/cmd/apiserver/app"
	"hit.edu/framework/pkg/component-base/cli"
	"os"
)

func main() {
	command := app.NewAPIServerCommand()
	code := cli.Run(command)
	os.Exit(code)
}
