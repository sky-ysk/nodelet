package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
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

var scheme = runtime.NewScheme()

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作
// 1个group，1个Action，每个Action一个Runtime， 一共1个Runtime，其中第一个group为训练任务（debian1上处理），测试迁移到pve2上，提前预部署的情况，GroupSpec加上参数Replicas
func main() {
	moduleName := "testModule"
	logs.Init(moduleName)

	//创建ClientSet
	clientSet := initClientSet(scheme)

	// eventclient := clientSet.Core().Events("test")
	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T1" // 第一个Task的Name

	// group
	group1_1Name := "G1" // 第一个Task下的第一个GroupName
	group1_1Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	// runtime是否细粒度控制
	// runtime1_1_1_1FineGrainedControl := false
	// runtime1_1_1_1FineGrainedControlPort := "5123"

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	gs1 := apis.GroupSpec{
		Replicas: group1_1Replicas,
		Name:     group1_1Name,
		Parents:  make([]string, 0),
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Image:                    "/home/kcm/tmp/wasm/example.wasm", //暂时以文件本地地址进行测试
						Type:                     apis.ByWasm,
						Parents:                  make([]string, 0),
						Conditions:               &runtime1_1_1_1Condition,
						Command:                  []string{},
						Args:                     []string{},
						EnableFineGrainedControl: true,
						// EnableFineGrainedControlPort: "8080",
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

	// 删除事件
	// eventsClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})

	// time.Sleep(10 * time.Second)
	// logs.Info("after 10s , post event for migration")
	// postEventForMigrate(eventsClient, scheme, group_create)
	// prompt()

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
		Timeout: 100 * time.Second,
	}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return clientSet
}
