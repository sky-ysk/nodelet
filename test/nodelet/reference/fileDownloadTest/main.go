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
	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
)

// 测试部署一个Task，一个Group，一个Action，每个Action一个Runtime
// 用于测试DataDependency的条件和NodeDependency的条件是否正确运行
func main() {
	moduleName := "testModule"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://localhost:10000", //http://suda801.wangwanu.com:11006
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
	group1_1Name := "G1" // 第一个Task下的第一个GroupName

	group1_1Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_1_1_2Name := "R2" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false
	runtime1_1_1_2FineGrainedControl := false

	// 数据依赖（../tmp/testFolder）
	DataDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.DataDependency,
		LeftValue: apis.Value{
			Type:      apis.FileData,
			Name:      "asdasd",
			Value:     "0",
			ValueType: "string",
			From:      "",
		},
		RightValue: apis.Value{
			Type:      apis.ConstData,
			Name:      "asdasd",
			Value:     "1",
			ValueType: "string",
			From:      "",
		},
		Signal: apis.Equal,
		Join:   "",
		Result: apis.False,
	}
	//上传文件，runtime的Data[]里面的每一个文件都需要上传
	//filePath := "/home/public/goprojects/Combine-ysk-0102/tmp/testFolder/upload.py"
	//filePath := "/home/public/goprojects/Combine-ysk-0102/tmp/testFolder/test.txt"
	filePath := "/home/public/goprojects/Combine-ysk-0102/tmp/testFolder/test.txt"
	UploadFile(filePath)
	filePath = "/home/public/goprojects/Combine-ysk-0102/tmp/testFolder/upload.py"
	UploadFile(filePath)

	// node依赖
	NodeDependencyCondition := GetNodeDepencyConditionFormula("R1")

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			DataDependencyConditionFormula,
		},
	}
	runtime1_1_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			NodeDependencyCondition,
			DataDependencyConditionFormula,
		},
	}
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
		Replicas:   group1_1Replicas,
		Name:       group1_1Name,
		Parents:    make([]string, 0),
		Conditions: &group1_1Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Data:                     []apis.DataSpec{apis.DataSpec{Name: "upload.py"}, apis.DataSpec{Name: "test.txt"}}, // 需要下载的文件
						Args:                     []string{"upload.py"},                                                              //20s
						Parents:                  make([]string, 0),                                                                  // 加入Parents
						Conditions:               &runtime1_1_1_1Condition,
						Inputs:                   []apis.Value{apis.Value{Value: "test.txt"}, apis.Value{Value: "success.txt"}},
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_2Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Data:                     []apis.DataSpec{apis.DataSpec{Name: "upload.py"}, apis.DataSpec{Name: "test.txt"}}, // 需要下载的文件
						Args:                     []string{"upload.py"},                                                              //20s
						Parents:                  []string{runtime1_1_1_1Name},                                                                  // 加入Parents
						Conditions:               &runtime1_1_1_2Condition,
						Inputs:                   []apis.Value{apis.Value{Value: "test.txt"}, apis.Value{Value: "success.txt"}},
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_2FineGrainedControl,
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
	tasksClient := clientSet.Core().Tasks("test")
	groupsClient := clientSet.Core().Groups("test")
	actionsClient := clientSet.Core().Actions("test")
	runtimesClient := clientSet.Core().Runtimes("test")
	eventsClient := clientSet.Core().Events("test")

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

func GetNodeDepencyConditionFormula(parentName string) apis.ConditionFormula {
	return apis.ConditionFormula{
		ConditionType: apis.NodeDependency,
		LeftValue: apis.Value{
			Type:      apis.LocalData,
			Name:      "NodeDependency",
			Value:     "0",
			ValueType: "string",
			From:      "runtime{" + parentName + "}",
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

// 上传文件
func UploadFile(filePath string) (string, error) {
	// 调用 utils.UploadFile 函数上传文件
	// 这里的 filePath 是要上传的文件路径
	// 返回上传结果和错误信息
	url := "http://localhost:10000/apis/resources/v1/upload"
	err := utils.UploadFile(filePath, "v1.0.0", url)
	if err != nil {
		fmt.Println("Upload failed:", err)
		return "", err
	} else {
		fmt.Println("Upload successful!")
		return "Upload successful!", nil
	}
}
