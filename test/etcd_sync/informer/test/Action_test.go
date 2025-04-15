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

func TestAction(t *testing.T) {
	// 跨域时需要将这个修改为跨域的域名
	target := informer.CreateTarget[*apis.Action]("broker", "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.120", 14399)
	target.URL = "http://0.0.0.0:14399" //用于域内初步测试，跨域时去掉
	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-actions",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "demo-action",
		},
	}
	obj, err := target.Create(context.TODO(), action, metav1.CreateOptions{})
	obj, err = target.Get(context.TODO(), "test", "demo-actions", "actions", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get action: %v", err)
	}
	action.Spec.Name = "1111"
	obj, err = target.Update(context.TODO(), action, metav1.UpdateOptions{})
	obj, err = target.Get(context.TODO(), "test", "demo-actions", "actions", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get action: %v", err)
	}

	err = target.Delete(context.TODO(), "test", "demo-actions", "actions", metav1.DeleteOptions{})
	_, name, err := informer.EncodeResourceType(obj)
	if err != nil {
		logs.Infof("Failed to encode action: %v", err)
	}
	fmt.Println(name)
}
