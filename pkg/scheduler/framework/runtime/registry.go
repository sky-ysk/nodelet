package runtime

import (
	"context"
	"fmt"

	"hit.edu/framework/pkg/scheduler/framework"
)

// 注册函数由插件传入
type PluginFactory = func(ctx context.Context, f framework.Handle) (framework.Plugin, error)

// 包含所有的可用的插件
// 插件必须先注册，才能被调度器使用
type Registry map[string]PluginFactory

// 注册插件
func (r Registry) Register(name string, factory PluginFactory) error {
	if _, ok := r[name]; ok {
		return fmt.Errorf("a plugin named %v already exists", name)
	}
	r[name] = factory
	return nil
}

// 取消注册插件
func (r Registry) Unregister(name string) error {
	if _, ok := r[name]; !ok {
		return fmt.Errorf("no plugin named %v exists", name)
	}
	delete(r, name)
	return nil
}

// 合并所有的插件
func (r Registry) Merge(in Registry) error {
	for name, factory := range in {
		if err := r.Register(name, factory); err != nil {
			return err
		}
	}
	return nil
}
