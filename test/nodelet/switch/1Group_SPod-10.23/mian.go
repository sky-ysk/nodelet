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
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"os"
	"time"
)

var scheme = runtime.NewScheme()

func main() {
	moduleName := "testModule"
	logs.Init(moduleName)

	//创建ClientSet
	clientSet := initClientSet(scheme)

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T1" // 第一个Task的Name
	// group1 - Client -A机器
	group1_1Name := "G1" // 第一个Task下的第一个GroupName
	group1_1Replicas := []int32{0, 0}
	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_1_1_1Input := []apis.Value{
		apis.Value{
			From: "/home/public/goprojects/reference/test/nodelet/switch/grpc-client-pod.yaml",
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
						Name:    runtime1_1_1_1Name,
						Type:    apis.ByPod,
						Parents: make([]string, 0), // 加入Parents
						Inputs:  runtime1_1_1_1Input,
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
	m := manager.NewManager(clientSet)
	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	_, err := m.CreateTask(ts, nil, "test", u.String(), "")
	if err != nil {
		panic(err)
	}
	logs.Info("Create task successfully======")

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
