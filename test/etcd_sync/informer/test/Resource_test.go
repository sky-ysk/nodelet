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

func TestResource(t *testing.T) {
	// 跨域时需要将这个修改为跨域的域名
	target := informer.CreateTarget[*apis.Resource_Node]("broker", "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.120", 14399)
	target.URL = "http://0.0.0.0:14399" //用于域内初步测试，跨域时去掉
	resource := &apis.Resource_Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-resources",
			Namespace: "Test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Resource_Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.ResourceSpec{},
	}
	obj, err := target.Create(context.TODO(), resource, metav1.CreateOptions{})
	obj, err = target.Get(context.TODO(), "Test", "demo-resources", "resources", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get resource: %v", err)
	}
	resource.Spec.Name = "demo-resources1"
	obj, err = target.Update(context.TODO(), resource, metav1.UpdateOptions{})
	obj, err = target.Get(context.TODO(), "Test", "demo-resources", "resources", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get resource: %v", err)
	}

	err = target.Delete(context.TODO(), "Test", "demo-resources", "resources", metav1.DeleteOptions{})
	_, name, err := informer.EncodeResourceType(obj)
	if err != nil {
		logs.Infof("Failed to encode resource: %v", err)
	}
	fmt.Println(name)
}
