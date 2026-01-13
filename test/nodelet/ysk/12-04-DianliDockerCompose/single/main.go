package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
)

// 测试部署-123-12
// 3个group，3个Action，每个Action两个Runtime， 一共6个Runtime，其中第一个group为训练任务（debian1上处理），第二个任务为推理任务（pve2上处理），第三个任务为机器人任务（pve2上处理）
func main() {
	namespace := "HenanEP"
	moduleName := "testModule"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://120.220.95.189:48120", //http://suda801.wangwanu.com:11006
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

	//tasksClient := clientSet.Core().Tasks("test")
	//groupsClient := clientSet.Core().Groups("test")

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T1" // 第一个Task的Name

	// group
	group1_1Name := "G81" // 第一个Task下的第一个GroupName

	group1_1Replicas := []int32{0, 0}

	// action
	actionName := "A1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName

	// runtime
	runtime_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	runtime_2Name := "R2" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	runtime_3Name := "R3" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	runtime_4Name := "R4" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	runtime_5Name := "R5" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	runtime_6Name := "R6" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false

	group1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
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
		Desc: &apis.Description{
			Label: map[string]string{
				"scheduler": "CloudNode1",
			},
			Docs: "电力拉起容器工作流",
		},
		Parents:    make([]string, 0),
		Conditions: &group1_1Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: actionName,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime_1Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-d", "--name=front_v3", "--network", "host", "--restart", "always", "-v", "/znxs/conf/nginx/nginx.conf:/etc/nginx/nginx.conf", "-v", "/znxs/conf/nginx/ssl:/etc/nginx/ssl", "-v", "/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime", "192.168.102.228:8088/znxs/front_v3:1.4.18.1"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               nil,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime_2Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-d", "--name=admin", "--network", "znxs_default", "--restart", "always", "--privileged", "-p", "20001:20001", "-v", "/znxs/data/resource:/resource", "-v", "/run/docker.sock:/var/run/docker.sock", "-v", "/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime", "-v", "/znxs/logs/admin:/logs", "192.168.102.228:8088/znxs/admin:1.4.18.1"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               nil,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime_3Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-d", "--name=hardware", "--network", "znxs_default", "--restart", "always", "-p", "20005:20005", "-p", "8972:8972", "-v", "/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime", "-v", "/znxs/data/resource:/resource", "-v", "/znxs/logs/hardware:/logs", "192.168.102.228:8088/znxs/hardware:1.4.18.1"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               nil,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime_4Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-d", "--name=device", "--network", "znxs_default", "--restart", "always", "-p", "20002:20002", "-v", "/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime", "-v", "/run/docker.sock:/var/run/docker.sock", "-v", "/znxs/data/resource:/resource", "-v", "/znxs/logs/device:/logs", "192.168.102.228:8088/znxs/device:1.4.18.1"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               nil,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime_5Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-d", "--name=patrol", "--network", "znxs_default", "--restart", "always", "-p", "20003:20003", "-v", "/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime", "-v", "/znxs/data/resource:/resource", "-v", "/znxs/logs/patrol:/logs", "192.168.102.228:8088/znxs/patrol:1.4.18.1"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               nil,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime_6Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-d", "--name=protocol", "--network", "znxs_default", "--restart", "always", "-p", "20004:20004", "-p", "51001:10011", "-p", "9300:9300/udp", "-v", "/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime", "-v", "/znxs/data/resource:/resource", "-v", "/znxs/logs/protocol:/logs", "192.168.102.228:8088/znxs/protocol:1.4.18.1"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               nil,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
				},
			},
		},
	}

	ts := apis.TaskSpec{
		Name: task1Name,
		Desc: &apis.Description{
			Docs: "电力拉起容器工作流",
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
	tasksClient := clientSet.Core().Tasks(namespace)
	groupsClient := clientSet.Core().Groups(namespace)
	actionsClient := clientSet.Core().Actions(namespace)
	runtimesClient := clientSet.Core().Runtimes(namespace)
	eventsClient := clientSet.Core().Events(namespace)

	// Task资源
	logs.Info("======Task")
	list1, err := tasksClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, task := range list1.Items {
		err := tasksClient.Delete(context.TODO(), task.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Task删除成功: %v", task.Name)
	}
	// group资源
	logs.Info("======Group")
	list2, err := groupsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, group := range list2.Items {
		err := groupsClient.Delete(context.TODO(), group.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Group删除成功: %v", group.Name)
	}
	// action 资源
	logs.Info("======Action")
	list3, err := actionsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, action := range list3.Items {
		err := actionsClient.Delete(context.TODO(), action.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Action删除成功: %v", action.Name)
	}
	// runtime 资源
	logs.Info("======Runtime")
	list4, err := runtimesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, runtime := range list4.Items {
		err := runtimesClient.Delete(context.TODO(), runtime.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Runtime删除成功:%v", runtime.Name)
	}

	// Event资源
	logs.Info("======Event")
	list5, err := eventsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, event := range list5.Items {
		err := eventsClient.Delete(context.TODO(), event.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Event删除成功: %v", event.Name)
	}

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
