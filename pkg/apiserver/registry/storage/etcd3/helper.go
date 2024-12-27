// 这个文件包含了一些提供某些辅助功能的函数
package etcd3

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
)

// 根据类型设置对应零值
func SetZeroValue(objPtr runtime.Object) error {
	v, err := EnforcePtr(objPtr)
	if err != nil {
		return err
	}
	v.Set(reflect.Zero(v.Type()))
	return nil
}

// EnforcePtr 确保传入的对象是一个指针，并返回对应值
func EnforcePtr(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}

func GetItemsPtr(list runtime.Object) (interface{}, error) {
	obj, err := getItemsPtr(list)
	if err != nil {
		return nil, fmt.Errorf("%T is not a list: %v", list, err)
	}
	return obj, nil
}

var (
	errExpectFieldItems = errors.New("no Items field in this object")
	errExpectSliceItems = errors.New("Items field must be a slice of objects")
)

// 从list中获取Items字段
func getItemsPtr(list runtime.Object) (interface{}, error) {
	v, err := EnforcePtr(list)
	if err != nil {
		return nil, err
	}
	
	items := v.FieldByName("Items")
	if !items.IsValid() {
		return nil, errExpectFieldItems
	}
	switch items.Kind() {
	case reflect.Interface, reflect.Pointer:
		target := reflect.TypeOf(items.Interface()).Elem()
		if target.Kind() != reflect.Slice {
			return nil, errExpectSliceItems
		}
		return items.Interface(), nil
	case reflect.Slice:
		return items.Addr().Interface(), nil
	default:
		return nil, errExpectSliceItems
	}
}

// initializationSignalFrom 用于信号通知，用于优先级调度
type InitializationSignal interface {
	// Signal 发送初始化完成的信号
	Signal()
	// Wait 用于等待初始化完成
	Wait()
}

type priorityAndFairnessKeyType int

const (
	// priorityAndFairnessInitializationSignalKey 用于存储和检索与Watch初始化信号相关的值的键
	priorityAndFairnessInitializationSignalKey priorityAndFairnessKeyType = iota
)

// initializationSignalFrom提取InitializationSignal
func initializationSignalFrom(ctx context.Context) (InitializationSignal, bool) {
	signal, ok := ctx.Value(priorityAndFairnessInitializationSignalKey).(InitializationSignal)
	return signal, ok && signal != nil
}

// WatchInitialized 发送Watch请求初始化完成的信号
func WatchInitialized(ctx context.Context) {
	if signal, ok := initializationSignalFrom(ctx); ok {
		signal.Signal()
	}
}
