package plugins

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"testing"
	"time"
)

func TestBuildGroupsRequest(t *testing.T) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
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
		fmt.Println(err)
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Task为例，
	// 获取访问Task的客户端
	// 默认访问的Namespace是 ""

	tasksClient := clientSet.Core().Tasks("test")
	task := mockGetTask()
	results, err := tasksClient.Create(context.TODO(), &task, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Created task ", results)

}

func TestAddGroupsIndatabus(t *testing.T) {
	task := mockGetTask()
	ctx := context.Background()
	req := BuildSendGroupsRequest(ctx, &task)
	plugin := &ScorePluginDBY{
		pluginClient: NewScorePluginClient(),
	}
	//NewScorePluginClient()
	plugin.SendGroups(ctx, &task)
	fmt.Println(req)
	fmt.Println(len(req.TopInfo))
	//TODO Fill the node name
	//plugin.Score(ctx, &task.Spec.Groups[0], "")
}

func TestListAllNodes(t *testing.T) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
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
			MaxIdleConns:        10000,            // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Node为例，
	// 获取访问Node的客户端
	// 默认访问的Namespace是 ""

	nodesClient := clientSet.Core().Nodes("test")
	logs.Info("listing 筛选的node")
	lstOpts := metav1.ListOptions{}
	list, err := nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}
}

func TestSendGroupsRequest(t *testing.T) {
	task := mockGetTask()
	ctx := context.Background()
	req := BuildSendGroupsRequest(ctx, &task)
	plugin := &ScorePluginDBY{
		pluginClient: NewScorePluginClient(),
	}
	//NewScorePluginClient()
	plugin.SendGroups(ctx, &task)
	fmt.Println(req)
	fmt.Println(len(req.TopInfo))
	//TODO Fill the node name
	//plugin.Score(ctx, &task.Spec.Groups[0], "")
}

func mockGetTask() apis.Task {
	reqs := make([]apis.ResourceRequirement, 0)
	req1 := apis.ResourceRequirement{
		Name:       "CPU",
		Lowbound:   "1",
		Upperbound: "1",
	}
	req2 := apis.ResourceRequirement{
		Name:       "RAM",
		Lowbound:   "1",
		Upperbound: "1",
	}
	req3 := apis.ResourceRequirement{
		Name:       "IO",
		Lowbound:   "1",
		Upperbound: "1",
	}
	reqs = append(reqs, req1, req2, req3)
	g1 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup1",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			//Parents:              [] string {1,2,3 },
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup1",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}
	g2 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup2",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			Parents:              []string{"testGroup1"},
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup2",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}
	g3 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup3",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			Parents:              []string{"testGroup2"},
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup3",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}

	return apis.Task{
		ObjectMeta: meta.ObjectMeta{
			Name: "testTask",
		},
		Spec: apis.TaskSpec{
			Groups: []apis.Group{g1, g2, g3},
		},
		Status: apis.TaskStatus{
			TaskID: "task1",
		},
	}
}
