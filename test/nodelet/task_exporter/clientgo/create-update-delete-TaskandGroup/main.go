package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"os"
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作
// 测试patch动词

func main() {
	moduleName := "testModule"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
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
		Timeout: 10 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Task为例，
	// 获取访问Task的客户端
	// 默认访问的Namespace是 ""

	groupsClient := clientSet.Core().Groups("")

	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: "TestGroup2", Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name: "TestGroup2",
			Actions: []apis.Action{
				apis.Action{
					Spec: apis.ActionSpec{
						Runtimes: []apis.Runtime{
							apis.Runtime{
								//Waiting: false,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			CheckDependencyCount: 0,
		},
	}

	//group.Spec.Actions[0].Spec.Runtimes[0].Waiting = true

	//runtimeJson, err := json.Marshal(group.Spec.Actions)
	//if err != nil {
	//	logs.Errorf("json marshal: RuntimeJson err:%v", err)
	//}
	patchGroupActionsRuntimes, err4 := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{
			"actions": group.Spec.Actions,
		},
	})
	if err4 != nil {
		logs.Errorf("json marshal:patchGroupActions err:%v", err)
	}
	result, err := groupsClient.Patch(context.TODO(), "TestGroup2", types.StrategicMergePatchType, patchGroupActionsRuntimes, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch patchGroupActionsRuntimes:group err:%v", err)
	}
	logs.Info(result)

}

// From K8s
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	logs.Info()
}
