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
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
	"net/http"
	"os"
	"time"
)

// 适配从pve2 迁移到 broker
// 修改1：
// 调度器代码: 触发k8s-master节点资源不足事件,从k8s-master迁移到broker--k8s-master节点的调度器代码可能要修改一下，这里为了和2Group_SPod_ahead_switch_kuayu统一
// k8s-master上的调度器修改成下面这样，broker下面的调度器不用修改代码，因为只有一个broker节点，不存在节点选择
// host, _, err := selectHost(priorityList, numberOfHighestScoredNodesToReport)
//
//	if strings.Contains(group.ObjectMeta.Name, "Train") {
//		host = "CloudNode1"
//	}
//
//	if strings.Contains(group.ObjectMeta.Name, "Reason") {
//		host = "EdgeNode1"
//	}
//
// 修改2：调度器关闭score插件
// 修改3：const NodeName = "k8s-master"
// 修改4：group1_1Replicas := []int32{0, 1}
// 修改5：recorder.EventForMigration(node, apis.EventTypeNormal, events.TriggerCrossMigration
var scheme = runtime.NewScheme()

const NodeName = "k8s-master"

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

	// group
	group1_1Name := "G1" // 第一个Task下的第一个GroupName
	group1_1Replicas := []int32{0, 1}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := true
	runtime1_1_1_1FineGrainedControlPort := "5123"

	// 程序依赖（requirements.txt）
	ProgramDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.ProgramDependency,
		LeftValue: apis.Value{
			Type:      apis.ResultsData,
			Name:      "ProgramDependency",
			Value:     "0",
			ValueType: "string",
			From:      "requirements1.txt",
		},
		RightValue: apis.Value{
			Type:      apis.ConstData,
			Name:      "ProgramDependency",
			Value:     "1",
			ValueType: "string",
			From:      "",
		},
		Signal: apis.Equal,
		Join:   "",
		Result: apis.False,
	}

	// 数据依赖（../tmp/testFolder）
	DataDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.DataDependency,
		LeftValue: apis.Value{
			Type:      apis.FileData,
			Name:      "python",
			Value:     "0",
			ValueType: "string",
			From:      "",
		},
		RightValue: apis.Value{
			Type:      apis.ConstData,
			Name:      "python",
			Value:     "1",
			ValueType: "string",
			From:      "",
		},
		Signal: apis.Equal,
		Join:   "",
		Result: apis.False,
	}

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
			DataDependencyConditionFormula,
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
						Name:                         runtime1_1_1_1Name,
						Type:                         apis.ByCommand,
						Command:                      []string{"python"},
						Data:                         []apis.DataSpec{apis.DataSpec{Name: "yolo-runner1.py"}, apis.DataSpec{Name: "requirements1.txt"}},
						Args:                         []string{"yolo-runner1.py"}, //20s
						Parents:                      make([]string, 0),           // 加入Parents
						Conditions:                   &runtime1_1_1_1Condition,
						EnvVar:                       []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl:     runtime1_1_1_1FineGrainedControl,
						EnableFineGrainedControlPort: &runtime1_1_1_1FineGrainedControlPort,
					},
				},
			},
		},
	}
	// 首先上传文件到文件仓库
	filePath := "/home/public/goprojects/reference/test/nodelet/task_exporter/dependency/requirements1.txt"
	UploadFile(filePath)
	filePath = "/home/public/workspace/yolo_projects/yolo-runner1.py"
	UploadFile(filePath)

	ts := apis.TaskSpec{
		Name: task1Name,
		Groups: []apis.GroupSpec{
			gs1,
		},
	}
	// 生成UUID
	m := manager.NewManager(clientSet)
	u := uuid.Must(uuid.NewV7())
	_, err := m.CreateTask(ts, nil, "test", u.String(), "")
	if err != nil {
		panic(err)
	}

	logs.Info("下发一个任务======")
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

var node = &apis.Node{
	ObjectMeta: meta.ObjectMeta{Name: NodeName, Namespace: "test"},
	TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
	Spec:       apis.NodeSpec{NodeName: NodeName},
}

func postEventForMigrate(client core.EventInterface) {
	logs.Info("发送跨域迁移事件======")
	// 这些配置实际在组件初始化时就已经完成
	ctx := context.Background()
	eventBroadcaster := recorder.NewBroadcaster(recorder.WithContext(ctx))
	defer eventBroadcaster.Shutdown()
	eventBroadcaster.StartRecordingToSink(ctx, &core.EventSinkImpl{Interface: client})
	recorder := eventBroadcaster.NewRecorder(scheme, "test-controller")

	// 通过 recorder.Event或 recorder.Eventf可以生成事件
	time.Sleep(10 * time.Millisecond)
	recorder.EventForMigration(node, apis.EventTypeNormal, events.TriggerCrossMigration, fmt.Sprintf("The node %vresource is shorted", NodeName), "")
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

// 上传文件
func UploadFile(filePath string) (string, error) {
	// 调用 utils.UploadFile 函数上传文件
	// 这里的 filePath 是要上传的文件路径
	// 返回上传结果和错误信息
	url := "http://localhost:8888/upload" //url := "http://localhost:8888/apis/resources/v1/upload"
	err := utils.UploadFile(filePath, "v1.0.0", url)
	if err != nil {
		fmt.Println("Upload failed:", err)
		return "", err
	} else {
		fmt.Println("Upload successful!")
		return "Upload successful!", nil
	}
}
