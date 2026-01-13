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

// 调度器代码: 触发CloudNode1资源不足事件,从CloudNode1迁移到CloudNode2
// if strings.Contains(group.ObjectMeta.Name, "G1") {
// host = "CloudNode1"
// }
// if strings.Contains(group.ObjectMeta.Name, "copy") {
// host = "CloudNode2"
// }
var scheme = runtime.NewScheme()

const NodeName = "cloudNode1" // 1$
var group1_1Name = "G81"      // 第一个Task下的第一个GroupName
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

	group1_1Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false
	IsHttpService := false
	IsHttpClient := false
	PlantformType := "linux"

	gs1 := apis.GroupSpec{
		Replicas: group1_1Replicas,
		Name:     group1_1Name,
		Desc: &apis.Description{
			Label: map[string]string{"scheduler": "EdgeNode1"},
			Docs:  "wasm-linux测试任务",
		},
		Parents: make([]string, 0),
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:    runtime1_1_1_1Name,
						Type:    apis.ByCommand,
						Command: []string{"./power_rust_infer_server_v260103"},
						//  /tmp/file/power_rust_infer_server/power_rust_infer_server_v260103 --port 20012 --config ./ftp.SECRET.yaml
						// Args:                     []string{"--port", "20013", "--config", "/tmp/file/power_rust_infer_server/ftp.SECRET.yaml", "--serviceName", "picAnalyse"},
						Args:                     []string{"--port", "20012", "--config", "./ftp.SECRET.yaml"},
						Data:                     []apis.DataSpec{},
						Conditions:               nil,
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
						IsHttpService:            IsHttpService,
						IsHttpClient:             IsHttpClient,
						PlantformType:            PlantformType,
						AbsDirectory: "/tmp/file/power_rust_infer_server",
					},
				},
			},
		},
	}

	ts := apis.TaskSpec{
		Name: task1Name,
		Desc: &apis.Description{
			Label: map[string]string{"scheduler": "EdgeNode1"},
			Docs:  "wasm-linux测试任务",
		},
		Groups: []apis.GroupSpec{
			gs1,
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
		Host:    "http://120.220.95.189:48120",
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
