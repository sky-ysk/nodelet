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

func TestGroup(t *testing.T) {

	// 跨域时需要将这个修改为跨域的域名
	target := informer.CreateTarget[*apis.Group]("broker", "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.120", 14399)
	target.URL = "http://0.0.0.0:14399" //用于域内初步测试，跨域时去掉
	action := apis.Action{
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
			Runtimes: []apis.Runtime{
				apis.Runtime{
					Name: "demo-runtime",
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: []apis.RuntimeStatus{
				apis.RuntimeStatus{
					NodeName: "demo-runtime",
					Phase:    "running",
				},
			},
		},
	}
	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-groups",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Name: "demo-group",
			Actions: []apis.Action{
				action,
			},
		},
	}
	obj, err := target.Create(context.TODO(), group, metav1.CreateOptions{})
	obj, err = target.Get(context.TODO(), "test", "demo-groups", "groups", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get group: %v", err)
	}
	group.Spec.Name = "1111"
	obj, err = target.Update(context.TODO(), group, metav1.UpdateOptions{})
	obj, err = target.Get(context.TODO(), "test", "demo-groups", "groups", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get group: %v", err)
	}

	err = target.Delete(context.TODO(), "test", "demo-groups", "groups", metav1.DeleteOptions{})
	_, name, err := informer.EncodeResourceType(obj)
	if err != nil {
		logs.Infof("Failed to encode group: %v", err)
	}
	fmt.Println(name)
}
