package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"net/http"
	"os"
	"time"
)

var scheme = runtime.NewScheme()

const NodeName = "debian1"

// 测试切换
// 1个group，1个Action，每个Action1个Runtime， 一共1个Runtime
func main() {
	moduleName := "testModule"
	logs.Init(moduleName)

	//创建ClientSet
	clientSet := initClientSet(scheme)

	eventclient := clientSet.Core().Events("test")
	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T1" // 第一个Task的Name

	// group1 - Client -A机器
	group1_1Name := "G1" // 第一个Task下的第一个GroupName
	group1_1Replicas := []int32{1, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := true
	runtime1_1_1_1FineGrainedControlPort := "30052"
	runtime1_1_1_1FineGrainedControlService := "172.110.0.103" //A机器IP地址

	runtime1_1_1_1Input := []apis.Value{
		apis.Value{
			From: "/home/public/goprojects/reference/test/nodelet/switch/grpc-client-pod.yaml",
		},
	}

	// group2 -Server-B机器
	group1_2Name := "G2" // 第一个Task下的第一个GroupName
	group1_2Replicas := []int32{0, 0}

	// action
	action1_2_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_2_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	// runtime是否细粒度控制
	runtime1_2_1_1FineGrainedControl := true
	runtime1_2_1_1FineGrainedControlPort := "30051"
	runtime1_2_1_1FineGrainedControlService := "172.110.0.104" //B机器ip地址

	runtime1_2_1_1Input := []apis.Value{
		apis.Value{
			From: "/home/public/goprojects/reference/test/nodelet/switch/grpc-server-pod.yaml",
		},
	}

	gs1 := apis.GroupSpec{
		ResourceRequirements: []apis.ResourceRequirement{
			apis.ResourceRequirement{
				Name:       "CPU",
				Lowbound:   "2",
				Upperbound: "4",
			},
			apis.ResourceRequirement{
				Name:       "RAM",
				Lowbound:   "2",
				Upperbound: "4",
			},
		},
		Replicas: group1_1Replicas,
		Name:     group1_1Name,
		Parents:  make([]string, 0),
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                            runtime1_1_1_1Name,
						Type:                            apis.ByPod,
						Command:                         []string{},
						Args:                            []string{},        //20s
						Parents:                         make([]string, 0), // 加入Parents
						EnvVar:                          []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						Inputs:                          runtime1_1_1_1Input,
						EnableFineGrainedControl:        runtime1_1_1_1FineGrainedControl,
						EnableFineGrainedControlService: &runtime1_1_1_1FineGrainedControlService,
						EnableFineGrainedControlPort:    &runtime1_1_1_1FineGrainedControlPort,
					},
				},
			},
		},
	}
	// Server -B 机器
	gs2 := apis.GroupSpec{
		ResourceRequirements: []apis.ResourceRequirement{
			apis.ResourceRequirement{
				Name:       "CPU",
				Lowbound:   "2",
				Upperbound: "4",
			},
			apis.ResourceRequirement{
				Name:       "RAM",
				Lowbound:   "2",
				Upperbound: "4",
			},
		},
		Replicas: group1_2Replicas,
		Name:     group1_2Name,
		Parents:  make([]string, 0),
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_2_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                            runtime1_2_1_1Name,
						Type:                            apis.ByPod,
						Command:                         []string{},
						Args:                            []string{},        //20s
						Parents:                         make([]string, 0), // 加入Parents
						EnvVar:                          []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						Inputs:                          runtime1_2_1_1Input,
						EnableFineGrainedControl:        runtime1_2_1_1FineGrainedControl,
						EnableFineGrainedControlService: &runtime1_2_1_1FineGrainedControlService,
						EnableFineGrainedControlPort:    &runtime1_2_1_1FineGrainedControlPort,
					},
				},
			},
		},
	}

	ts := apis.TaskSpec{
		Name: task1Name,
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}
	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	m := manager.NewManager(clientSet)
	task, err := m.CreateTask(ts, nil, "test", u.String(), "")
	if err != nil {
		panic(err)
	}
	str, err := analyzer.SerializeToJson(task)
	if err != nil {
		return
	}
	fmt.Println(str)

	prompt()
	postEventForMigrate(eventclient)
	prompt()

}

// From K8s
func prompt() {
	fmt.Printf("-> Press Return key to continue.发送一个迁移事件")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	logs.Info()
}

func GetNodeDepencyConditionFormula(parentName string) apis.ConditionFormula {
	return apis.ConditionFormula{
		LeftValue: apis.Value{
			Type:      apis.ResultsData,
			Name:      "NodeDependency",
			Value:     "0",
			ValueType: "string",
			From:      parentName,
		},
		RightValue: apis.Value{
			Type:      apis.ConstData,
			Name:      "NodeDependency",
			Value:     "1",
			ValueType: "string",
			From:      "",
		},
		Signal: apis.Equal,
		Join:   "",
		Result: apis.False,
	}
}

var node = &apis.Node{
	ObjectMeta: meta.ObjectMeta{Name: NodeName, Namespace: "test"},
	TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
	Spec:       apis.NodeSpec{NodeName: NodeName},
}

func postEventForMigrate(client core.EventInterface) {
	// 这些配置实际在组件初始化时就已经完成
	ctx := context.Background()
	eventBroadcaster := recorder.NewBroadcaster(recorder.WithContext(ctx))
	defer eventBroadcaster.Shutdown()
	eventBroadcaster.StartRecordingToSink(ctx, &core.EventSinkImpl{Interface: client})
	recorder := eventBroadcaster.NewRecorder(scheme, "test-controller")

	// 通过 recorder.Event或 recorder.Eventf可以生成事件
	recorder.EventForMigration(node, apis.EventTypeNormal, events.TriggerLocalMigration, fmt.Sprintf("Node Name:\t %s is shortage", node.Name), "")
	// recorder.Eventf(group, apis.EventTypeNormal, events.ReadyToMigrate, fmt.Sprintf("The task %v is ready for migration", group.Spec.Actions[0].Name))
}
func initClientSet(scheme *runtime.Scheme) *clients.ClientSet {
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	// 创建ClientSet
	c := &rest.Config{
		Host:    "http://127.0.0.1:10000",
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
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return clientSet
}
