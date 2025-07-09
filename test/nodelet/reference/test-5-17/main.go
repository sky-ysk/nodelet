package main

import (
	"bufio"
	"fmt"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/nodelet/events"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
)

// 简单任务测试
var scheme = runtime.NewScheme()
var group1_1Name = "G1" // 第一个Task下的第一个GroupName
var clientSet = initClientSet(scheme)
var m = manager.NewManager(clientSet)

// 提交工作流测试
// 1个group，1个Action，每个Action1个Runtime， 一共1个Runtime
func main() {
	moduleName := "testModule"
	logs.Init(moduleName)

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T1" // 第一个Task的Name

	// group
	group1_1Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName

	// 程序依赖（requirements.txt）
	ProgramDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.ProgramDependency,
		LeftValue: apis.Value{
			Type:      apis.ResultsData,
			Name:      "ProgramDependency",
			Value:     "0",
			ValueType: "string",
			From:      "/home/public/goprojects/Combine-ysk-0102/adaptive-scheduling-framework/test/nodelet/task_exporter/dependency/requirements.txt",
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

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
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
						Name:       runtime1_1_1_1Name,
						Type:       apis.ByCommand,
						Command:    []string{"python"},
						Args:       []string{"/home/public/workspace/heongtong_yolo_linux/train.py"}, //20s
						Parents:    make([]string, 0),                                                // 加入Parents
						Conditions: &runtime1_1_1_1Condition,
						EnvVar:     []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
					},
				},
			},
		},
	}

	ts := apis.TaskSpec{
		Name: task1Name,
		Groups: []apis.GroupSpec{
			gs1,
		},
	}
	// 生成UUID
	u := uuid.Must(uuid.NewV7())

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
	postEventForKill_ForGroup()
	prompt()

}

// From K8s
func prompt() {
	fmt.Printf("-> Press Return key to continue.发送kill指令")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	logs.Info()
}

var group1 = &apis.Group{
	ObjectMeta: meta.ObjectMeta{Name: "Group1", Namespace: "test"},
	TypeMeta:   meta.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
	Spec:       apis.GroupSpec{Name: "G1"},
}

func postEventForKill_ForGroup() {
	logs.Info("发送kill事件======")
	// 这些配置实际在组件初始化时就已经完成

	groups, err := m.GetGroups("test")
	if err != nil {
		logs.Errorf("GetGroups err: %v", err)
	}
	var groupName string
	for i := range groups.Items {
		group := groups.Items[i]
		if group.Spec.Name == group1_1Name && !strings.Contains(group.Name, "copy") && group.Status.Phase == apis.Running {
			groupName = group.Name
		}
	}
	group, err := m.GetGroup(groupName, "test")
	if err != nil {
		logs.Errorf("GetGroup err: %v", err)
	}
	logs.Infof("group Name:%v,groupNameSpace:%v", group.Name, group.Namespace)
	// 通过 recorder.Event或 recorder.Eventf可以生成事件
	time.Sleep(10 * time.Millisecond)
	//group = group1
	m.LogEvent(group, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("The group %v is need to killing", group.Name), group.Namespace)
	logs.Infof("++++++++++++++++++++++++++++++++++++++++++++")
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
