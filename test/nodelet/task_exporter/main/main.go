package main

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task"
	"time"
)

func simulateBinaryAITask() []*apis.Group {
	r := apis.Runtime{Type: apis.ByBinary}
	runtimes := []apis.Runtime{r}
	action := apis.Action{TypeMeta: runtime.TypeMeta{Kind: "", APIVersion: ""}, Spec: apis.ActionSpec{Runtimes: runtimes}}
	actions := []apis.Action{action}
	groupInfo := &apis.Group{
		TypeMeta:   runtime.TypeMeta{Kind: "", APIVersion: ""},
		ObjectMeta: meta.ObjectMeta{Name: "AI", Namespace: "switch"},
		Spec:       apis.GroupSpec{Name: "AI", Parents: make([]string, 0), Desc: apis.Description{}, Type: apis.Norm, Conditions: apis.Conditions{}, Actions: actions},
	}
	groupInfos := []*apis.Group{groupInfo}
	return groupInfos
}

// deployment：grpc-client  namespace：switch  type：apis.ByDeployment  image:"registry.cn-hangzhou.aliyuncs.com/sudasuzhou/grpc_test:go-grpc-client-v5.0"  Name:"grpc-client"(container容器名) EnableFineGrainedControl: false(是否需要细粒度管理)
func simulateDeploymentGrpcClientTask() []*apis.Group {
	actionSpec := apis.ActionSpec{
		Name: "grpc-client-action",
		Runtimes: []apis.Runtime{
			{
				Replicas: 1,
				Selector: map[string]string{
					"app": "grpc-client",
				},
				Labels: map[string]string{ // 定义 Deployment 模板的标签
					"app": "grpc-client",
				},
				Type:    apis.ByDeployment,
				Name:    "grpc-client",                                                                // 对应yaml中的deployment name
				Image:   "registry.cn-hangzhou.aliyuncs.com/sudasuzhou/grpc_test:go-grpc-client-v5.0", // 对应yaml中的image
				Command: []string{},                                                                   // 如果需要
				EnvVar:  []apis.EnvVar{                                                                //{Name: "GRPC_SERVER", Value: "server-address"},
				},
				Resources: []apis.ResourceSpec{ // 定义请求的资源
				},
			},
		},
		EnableFineGrainedControl: false,
	}
	action := apis.Action{TypeMeta: runtime.TypeMeta{Kind: "", APIVersion: ""}, ObjectMeta: meta.ObjectMeta{}, Spec: actionSpec}
	actions := []apis.Action{action}
	groupInfo := &apis.Group{
		TypeMeta:   runtime.TypeMeta{Kind: "apps", APIVersion: "v1"},
		ObjectMeta: meta.ObjectMeta{Name: "grpc-client", Namespace: "switch"},
		Spec: apis.GroupSpec{
			Name:       "",
			Parents:    make([]string, 0),
			Desc:       apis.Description{},
			Type:       apis.Norm,
			Conditions: apis.Conditions{},
			Actions:    actions,
			Labels:     map[string]string{"app": "grpc-client"},
		},
	}
	groupInfos := []*apis.Group{groupInfo}
	return groupInfos
}

// 关键name：grpc-server
func simulateDeploymentGrpcServerTask() []*apis.Group {
	actionSpec := apis.ActionSpec{
		Name: "grpc-server-action",
		Runtimes: []apis.Runtime{
			{
				Replicas: 1,
				Selector: map[string]string{
					"app": "grpc-server",
				},
				Labels: map[string]string{ // 定义 Deployment 模板的标签
					"app": "grpc-server",
				},
				Type:    apis.ByDeployment,
				Name:    "grpc-server",                                                                // 对应yaml中的deployment name
				Image:   "registry.cn-hangzhou.aliyuncs.com/sudasuzhou/grpc_test:go-grpc-server-v5.0", // 对应yaml中的image
				Ports:   []apis.Port{apis.Port{TargetPort: 50051}},
				Command: []string{},    // 如果需要
				EnvVar:  []apis.EnvVar{ //{Name: "GRPC_SERVER", Value: "server-address"},
				},
				Resources: []apis.ResourceSpec{ // 定义请求的资源
				},
			},
		},
		EnableFineGrainedControl: false,
	}
	action := apis.Action{TypeMeta: runtime.TypeMeta{Kind: "", APIVersion: ""}, ObjectMeta: meta.ObjectMeta{}, Spec: actionSpec}
	actions := []apis.Action{action}
	groupInfo := &apis.Group{
		TypeMeta:   runtime.TypeMeta{Kind: "apps", APIVersion: "v1"},
		ObjectMeta: meta.ObjectMeta{Name: "grpc-server", Namespace: "switch"},
		Spec: apis.GroupSpec{
			Name:       "",
			Parents:    make([]string, 0),
			Desc:       apis.Description{},
			Type:       apis.Norm,
			Conditions: apis.Conditions{},
			Actions:    actions,
			Labels:     map[string]string{"app": "grpc-server"},
		},
	}
	groupInfos := []*apis.Group{groupInfo}
	return groupInfos
}

// pod
func simulatePodGrpcClientTask() []*apis.Group {
	actionSpec := apis.ActionSpec{
		Name: "grpc-client-action",
		Runtimes: []apis.Runtime{
			{
				Type:      apis.ByPod,
				Name:      "grpc-client",                                                                // 对应yaml中的deployment name
				Image:     "registry.cn-hangzhou.aliyuncs.com/sudasuzhou/grpc_test:go-grpc-client-v5.0", // 对应yaml中的image
				Command:   []string{},                                                                   // 如果需要
				EnvVar:    []apis.EnvVar{},
				Resources: []apis.ResourceSpec{},
			},
		},
		EnableFineGrainedControl: false,
	}
	action := apis.Action{TypeMeta: runtime.TypeMeta{Kind: "", APIVersion: ""}, ObjectMeta: meta.ObjectMeta{Name: "", Namespace: ""}, Spec: actionSpec}
	actions := []apis.Action{action}
	groupInfo := &apis.Group{
		TypeMeta:   runtime.TypeMeta{Kind: "apps", APIVersion: "1.0"},
		ObjectMeta: meta.ObjectMeta{Name: "grpc-client", Namespace: "switch"},
		Spec:       apis.GroupSpec{Name: "", Parents: make([]string, 0), Desc: apis.Description{}, Type: apis.Norm, Conditions: apis.Conditions{}, Actions: actions},
	}
	groupInfos := []*apis.Group{groupInfo}
	return groupInfos
	return nil
}

// service
func simulateServiceGrpcServerTask() []*apis.Group {
	actionSpec := apis.ActionSpec{
		Name: "grpc-server-action",
		Runtimes: []apis.Runtime{
			{
				Type:        apis.ByService,
				Selector:    map[string]string{"app": "grpc-server"},
				Ports:       []apis.Port{apis.Port{Protocol: "TCP", Port: 50051, TargetPort: 50051}},
				ServiceType: "ClusterIP",
			},
		},
		EnableFineGrainedControl: false,
	}
	action := apis.Action{TypeMeta: runtime.TypeMeta{Kind: "kind", APIVersion: "1.0"}, ObjectMeta: meta.ObjectMeta{Name: "grpc-client", Namespace: "switch"}, Spec: actionSpec}
	actions := []apis.Action{action}
	groupInfo := &apis.Group{
		TypeMeta:   runtime.TypeMeta{Kind: "apps", APIVersion: "1.0"},
		ObjectMeta: meta.ObjectMeta{Name: "grpc-server-service", Namespace: "switch"},
		Spec:       apis.GroupSpec{Name: "", Parents: make([]string, 0), Desc: apis.Description{}, Type: apis.Norm, Conditions: apis.Conditions{}, Actions: actions},
	}
	groupInfos := []*apis.Group{groupInfo}
	return groupInfos
}

func main() {
	//模拟的任务信息
	//Service:      groupInfos := simulateServiceGrpcServerTask()
	//Pod:  		groupInfos := simulatePodGrpcClientTask()
	//Deployment: groupInfos := simulateDeploymentGrpcServerTask()   或者  groupInfos := simulateDeploymentGrpcClientTask()
	groupInfos := simulatePodGrpcClientTask()
	taskexporter, err := task.NewTaskExporter(nil)
	if err != nil {
		logs.Error("fail to create task exporter")
	}
	go func() {
		err2 := taskexporter.Run(context.Background())
		if err2 != nil {
			logs.Error("fail to run task exporter")
		}
	}()

	time.Sleep(2 * time.Second)
	task.ReceiveGroupInfo(groupInfos, "create")
	select {}
}
