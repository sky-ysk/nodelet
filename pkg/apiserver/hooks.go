package apiserver

import (
	"context"
	"fmt"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"log"
)

// PostStartHookFunc 在服务器启动之后调用的函数
type PostStartHookFunc func(context PostStartHookContext) error

// PreShutdownHookFunc 服务器关闭之前调用的函数
type PreShutdownHookFunc func() error

type PostStartHookContext context.Context

// PostStartHookProvider 为apiServer提供PostStart钩子函数的接口
type PostStartHookProvider interface {
	PostStartHook() (string, PostStartHookFunc, error)
}

type postStartHookEntry struct {
	hook PostStartHookFunc

	// done postStart完成后关闭通道
	done chan struct{}
}

type preShutdownHookEntry struct {
	hook PreShutdownHookFunc
}

// AddPostStartHook 向APIServer中添加PostStart钩子函数
func (s *APIServer) AddPostStartHook(name string, hook PostStartHookFunc) error {
	if len(name) == 0 {
		return fmt.Errorf("missing name")
	}
	if hook == nil {
		return fmt.Errorf("hook func may not be nil: %q", name)
	}
	if s.disabledPostStartHooks.Has(name) {
		return nil
	}

	s.postStartHookLock.Lock()
	defer s.postStartHookLock.Unlock()

	if s.postStartHooksCalled {
		return fmt.Errorf("unable to add %q because PostStartHooks have already been called", name)
	}
	if _, exists := s.postStartHooks[name]; exists {
		// this is programmer error, but it can be hard to debug
		return fmt.Errorf("unable to add %q because it was already registered ", name)
	}

	done := make(chan struct{})
	s.postStartHooks[name] = postStartHookEntry{hook: hook, done: done}

	return nil
}

// AddPostStartHookOrDie 添加不允许失败的PostStartHook,
func (s *APIServer) AddPostStartHookOrDie(name string, hook PostStartHookFunc) {
	if err := s.AddPostStartHook(name, hook); err != nil {
		log.Fatalf("Error registering PostStartHook %q: %v", name, err)
	}
}

// AddPreShutdownHook allows you to add a PreShutdownHook.
func (s *APIServer) AddPreShutdownHook(name string, hook PreShutdownHookFunc) error {
	if len(name) == 0 {
		return fmt.Errorf("missing name")
	}
	if hook == nil {
		return nil
	}

	s.preShutdownHookLock.Lock()
	defer s.preShutdownHookLock.Unlock()

	if s.preShutdownHooksCalled {
		return fmt.Errorf("unable to add %q because PreShutdownHooks have already been called", name)
	}
	if _, exists := s.preShutdownHooks[name]; exists {
		return fmt.Errorf("unable to add %q because it is already registered", name)
	}

	s.preShutdownHooks[name] = preShutdownHookEntry{hook: hook}

	return nil
}

// AddPreShutdownHookOrDie 添加不允许失败的PreShutdownHook
func (s *APIServer) AddPreShutdownHookOrDie(name string, hook PreShutdownHookFunc) {
	if err := s.AddPreShutdownHook(name, hook); err != nil {
		log.Fatalf("Error registering PreShutdownHook %q: %v", name, err)
	}
}

// RunPostStartHooks 为服务器运行PostStartHooks。
func (s *APIServer) RunPostStartHooks(ctx context.Context) {
	s.postStartHookLock.Lock()
	defer s.postStartHookLock.Unlock()
	s.postStartHooksCalled = true

	context := PostStartHookContext(ctx)

	for hookName, hookEntry := range s.postStartHooks {
		go runPostStartHook(hookName, hookEntry, context)
	}
}

// RunPreShutdownHooks 为服务器运行PreShutdownHooks
func (s *APIServer) RunPreShutdownHooks() error {
	var errorList []error

	s.preShutdownHookLock.Lock()
	defer s.preShutdownHookLock.Unlock()
	s.preShutdownHooksCalled = true

	for hookName, hookEntry := range s.preShutdownHooks {
		if err := runPreShutdownHook(hookName, hookEntry); err != nil {
			errorList = append(errorList, err)
		}
	}
	return utilerrors.NewAggregate(errorList)
}

// isPostStartHookRegistered 检查给定的PostStartHook是否已注册
func (s *APIServer) isPostStartHookRegistered(name string) bool {
	s.postStartHookLock.Lock()
	defer s.postStartHookLock.Unlock()
	_, exists := s.postStartHooks[name]
	return exists
}

func runPostStartHook(name string, entry postStartHookEntry, context PostStartHookContext) {
	var err error
	func() {
		err = entry.hook(context)
	}()

	if err != nil {
		log.Fatalf("PostStartHook %q failed: %v", name, err)
	}
	close(entry.done)
}

func runPreShutdownHook(name string, entry preShutdownHookEntry) error {
	var err error
	func() {
		err = entry.hook()
	}()
	if err != nil {
		return fmt.Errorf("PreShutdownHook %q failed: %v", name, err)
	}
	return nil
}
