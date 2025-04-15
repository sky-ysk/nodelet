package test

import (
	"context"
	"fmt"
	"testing"

	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/test/etcd_sync/informer"
)

func TestTask(t *testing.T) {

	// 跨域时需要将这个修改为跨域的域名
	target := informer.CreateTarget[*apis.Task]("broker", "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.120", 14399)
	target.URL = "http://0.0.0.0:14399" //用于域内初步测试，跨域时去掉
	task := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-tasks",
			Namespace: "Test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: "demo-task",
		},
	}
	obj, err := target.Create(context.TODO(), task, metav1.CreateOptions{})
	obj, err = target.Get(context.TODO(), "Test", "demo-tasks", "tasks", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get task: %v", err)
	}
	task.Spec.Name = "1111"
	obj, err = target.Update(context.TODO(), task, metav1.UpdateOptions{})
	obj, err = target.Get(context.TODO(), "Test", "demo-tasks", "tasks", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get task: %v", err)
	}

	err = target.Delete(context.TODO(), "Test", "demo-tasks", "tasks", metav1.DeleteOptions{})
	_, name, err := informer.EncodeResourceType(obj)
	if err != nil {
		logs.Infof("Failed to encode task: %v", err)
	}
	fmt.Println(name)
}
