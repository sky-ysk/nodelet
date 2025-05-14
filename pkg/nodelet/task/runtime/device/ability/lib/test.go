package lib

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
	"time"
)

type TestStrategy struct{}

func (ts *TestStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	logs.Infof("[DEVICE RUNTIME] this is a test, runtime is %v", runtime.Name)
	time.Sleep(time.Second * 5)
	return "test", nil
}
