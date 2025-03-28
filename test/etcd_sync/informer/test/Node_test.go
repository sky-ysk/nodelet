package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/test/etcd_sync/informer"
)

//	func init() {
//		go CreateWeb()
//	}
func CreateWeb() {
	apiserverhost := fmt.Sprintf("http://%s:%d", "0.0.0.0", 10000)
	serverhost := fmt.Sprintf("%s:%d", "0.0.0.0", 14399)

	// fmt.Println(apiserverhost)
	// fmt.Println(serverhost)
	// return
	logs.Init("sync_server")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    apiserverhost,
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8",
			ContentType:        "application/json; charset=UTF-8",
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
	}
	server := &informer.Sever_Config{
		Client: c,
	}
	logs.Infof("apisever run on %s", apiserverhost)
	server.CreateWebHandler(c, serverhost) //本地服务暴露的位置
}
func TestNode(t *testing.T) {
	target := informer.CreateTarget[*apis.Node]("broker", "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.120", 14399)
	target.URL = "http://0.0.0.0:14399" //用于域内初步测试，跨域时去掉
	node := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes",
			Namespace: "Test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
			HostName: "master",
		},
	}
	obj, err := target.Create(context.TODO(), node, metav1.CreateOptions{})
	obj, err = target.Get(context.TODO(), "Test", "demo-nodes", "nodes", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get node: %v", err)
	}
	node.Spec.NodeName = "1111"
	obj, err = target.Update(context.TODO(), node, metav1.UpdateOptions{})
	obj, err = target.Get(context.TODO(), "Test", "demo-nodes", "nodes", metav1.GetOptions{})
	if err != nil {
		logs.Infof("Failed to get node: %v", err)
	}

	err = target.Delete(context.TODO(), "Test", "demo-nodes", "nodes", metav1.DeleteOptions{})
	_, name, err := informer.EncodeResourceType(obj)
	if err != nil {
		logs.Infof("Failed to encode node: %v", err)
	}
	fmt.Println(name)

}
