package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	"hit.edu/framework/pkg/nodelet/events"
	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
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
var group1_1Name = "G91"      // 第一个Task下的第一个GroupName
var group1_2Name = "G71"      // 第一个Task下的第一个GroupName
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
	runtime1_1_1_1FineGrainedControl := true
	runtime1_1_1_1FineGrainedControlPort := "5123"

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}
	filePath := "/home/public/goprojects/ysk-1110/serviceProxy_test/client/client"
	UploadFile(filePath)
	filePath = "/home/public/goprojects/ysk-1110/serviceProxy_test/newServer/server"
	UploadFile(filePath)

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
						Command:                      []string{"./server"},
						Args:                         []string{"--ip xx --port 9091 --serviceName add"}, //20s   //3$
						Parents:                      make([]string, 0),                                 // 加入Parents
						Conditions:                   &runtime1_1_1_1Condition,
						EnvVar:                       []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl:     runtime1_1_1_1FineGrainedControl,
						EnableFineGrainedControlPort: &runtime1_1_1_1FineGrainedControlPort,
						IsHttpService:                true,
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
		Replicas: group1_1Replicas,
		Name:     group1_2Name,
		Parents:  []string{group1_1Name},
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                         runtime1_1_1_1Name,
						Type:                         apis.ByCommand,
						Command:                      []string{"./client"},
						Args:                         []string{"--ip xx --port xx"}, //20s   //3$
						Parents:                      make([]string, 0),             // 加入Parents
						Conditions:                   &runtime1_1_1_1Condition,
						EnvVar:                       []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl:     runtime1_1_1_1FineGrainedControl,
						EnableFineGrainedControlPort: &runtime1_1_1_1FineGrainedControlPort,
					},
				},
			},
		},
	}

	ts := apis.TaskSpec{
		Name: task1Name,
		Groups: []apis.GroupSpec{
			gs1,
			gs2,
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
	postEventForMigrate_ForGroup()
	prompt()

	tasksClient := clientSet.Core().Tasks("HenanEP")
	groupsClient := clientSet.Core().Groups("HenanEP")
	actionsClient := clientSet.Core().Actions("HenanEP")
	runtimesClient := clientSet.Core().Runtimes("HenanEP")
	eventsClient := clientSet.Core().Events("HenanEP")

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

func postEventForMigrate_ForGroup() {
	logs.Info("发送跨域迁移事件======")
	clientSet := initClientSet(scheme)
	m := manager.NewManager(clientSet)
	groups, err := m.GetGroups(namespace)
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
	group, err := m.GetGroup(groupName, namespace)
	if err != nil {
		logs.Errorf("GetGroup err: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	m.LogEvent(group, apis.EventTypeNormal, events.TriggerLocalMigration, fmt.Sprintf("The group %v is need to migrate", group.Name), group.Namespace)
}

func initClientSet(scheme *runtime.Scheme) *clients.ClientSet {
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	// 创建ClientSet
	c := &rest.Config{
		Host:    "http://172.150.0.24:10000",
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
	url := "http://172.150.0.24:8919/upload"
	err := utils.UploadFile(filePath, "v1.0.0", url)
	if err != nil {
		fmt.Println("Upload failed:", err)
		return "", err
	} else {
		fmt.Println("Upload successful!")
		return "Upload successful!", nil
	}
}
