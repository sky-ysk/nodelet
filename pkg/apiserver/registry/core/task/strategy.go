package task

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"math/rand"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/legacyscheme"
	"hit.edu/framework/pkg/apiserver/registry/storage/field"
	run "runtime"
)

type Strategy struct {
	runtime.ObjectTyper
}

var thisStrategy = &Strategy{legacyscheme.Scheme}

func (t Strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {

	task, ok := obj.(*apis.Task)
	if !ok {
		return
	}
	rand.Seed(time.Now().UnixNano() + int64(run.NumGoroutine()) + time.Now().UnixMicro())

	// 生成范围 30000-32760 的随机数
	min := 30000
	max := 32760
	randomNum := min + rand.Intn(max-min+1) // +1 包含上限值
	if task.Spec.RandomNum == 0 {
		task.Spec.RandomNum = int32(randomNum)
	}
	////生成 taskid
	//if task.Status.TaskID == "" {
	//	task.Status.TaskID = string(task.ObjectMeta.UID)
	//}
}
func (t Strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}
func (t Strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return nil
}
func (t Strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}
func (t Strategy) Canonicalize(obj runtime.Object) {}

// 资源的命名空间级别
func (t Strategy) NamespaceScoped() bool {
	// true表示资源必须配备命名空间信息
	return true
}
