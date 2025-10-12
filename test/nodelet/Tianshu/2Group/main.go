package main

import (
	"bufio"
	"fmt"
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
	"net/http"
	"os"
	"time"
)

var scheme = runtime.NewScheme()
var namespace = "HenanEP"

// 测试切换
// 1个group，1个Action，每个Action1个Runtime， 一共1个Runtime
func main() {
	moduleName := "testModule"
	logs.Init(moduleName)

	//创建ClientSet
	clientSet := initClientSet(scheme)

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T1" // 第一个Task的Name

	// group
	group1_1Name := "G91" // 第一个Task下的第一个GroupName
	group1_2Name := "G81" // 第一个Task下的第二个GroupName
	//group1_1Replicas := []int32{1, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_2_1Name := "A1" // 第一个Task下的第二个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_2_1_1Name := "R1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName
	// runtime是否细粒度控制
	runtime1_2_1_1FineGrainedControl := true
	runtime1_2_1_1FineGrainedControlPort := "5123"

	// 程序依赖（requirements.txt）
	ProgramDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.ProgramDependency,
		LeftValue: apis.Value{
			Type:      apis.ResultsData,
			Name:      "ProgramDependency",
			Value:     "0",
			ValueType: "string",
			From:      "/root/goprojects/reference/test/nodelet/task_exporter/dependency/requirements.txt", //2$
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
	runtime1_2_1_1Condition := apis.Conditions{
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
		Name: group1_1Name,
		Desc: &apis.Description{
			Label: map[string]string{
				"scheduler": "CloudNode1",
				"type":      "Train",
			},
			Docs: "Yolo训练任务", // 必须加上，不然前端报错
		},
		Parents: make([]string, 0),
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:       runtime1_1_1_1Name,
						Type:       apis.ByCommand,
						Command:    []string{"python"},
						Args:       []string{"/root/workspace/testTask/test.py"},
						Parents:    make([]string, 0),
						Conditions: &runtime1_1_1_1Condition,
					},
				},
			},
		},
	}

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
		Name: group1_2Name,
		Desc: &apis.Description{
			Label: map[string]string{
				"scheduler": "EdgeNode1",
				"type":      "Infer",
			},
			Docs: "推理任务", // 必须加上，不然前端报错
		},
		Parents: []string{group1_1Name},
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_2_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                         runtime1_2_1_1Name,
						Type:                         apis.ByCommand,
						Command:                      []string{"python"},
						Args:                         []string{"/root/workspace/yolo_projects/yolo-runner1.py"}, //20s   //3$
						Parents:                      make([]string, 0),                                         // 加入Parents
						Conditions:                   &runtime1_2_1_1Condition,
						EnableFineGrainedControl:     runtime1_2_1_1FineGrainedControl,
						EnableFineGrainedControlPort: &runtime1_2_1_1FineGrainedControlPort,
					},
				},
			},
		},
	}

	ts := apis.TaskSpec{
		Desc: &apis.Description{Docs: "Yolo训练推理工作流"}, // 必须加上，不然前端报错
		Name: task1Name,
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}
	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	m := manager.NewManager(clientSet)
	task, err := m.CreateTask(ts, nil, namespace, u.String(), "")
	if err != nil {
		panic(err)
	}
	str, err := analyzer.SerializeToJson(task)
	if err != nil {
		return
	}
	fmt.Println(str)

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
