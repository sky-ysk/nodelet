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

// 2个group，2个Action，每个Action一个Runtime， 一共2个Runtime，其中第一个group为训练任务，第二个任务为推理任务，依赖：推理任务需要接受训练任务的模型才能进行推理
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

	// Task  总共1个Task、2个Group、2个Action、2个runtime
	task1Name := "T1" // 第一个Task的Name

	// group
	group1_1Name := "G81" // 第一个Task下的第一个GroupName
	group1_2Name := "G71" // 第一个Task下的第二个GroupName

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName
	action1_2_1Name := "A2" // 第一个Task下的第二个Group下的第一个ActionName
	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_1_1_2Name := "R2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_2_1_1Name := "R1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName

	// 程序依赖（requirements.txt）
	ProgramDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.ProgramDependency,
		LeftValue: apis.Value{
			Type:      apis.ResultsData,
			Name:      "ProgramDependency",
			Value:     "0",
			ValueType: "string",
			From:      "requirements.txt",
			// From:      "/home/public/goprojects/test-0623/test/nodelet/task_exporter/dependency/requirements.txt", // 好像这里会卡主 原先；"requirements.txt"
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
	filePath := "/home/public/workspace/trainAndInfer/train_mnist.py"
	UploadFile(filePath)
	filePath = "/home/public/workspace/trainAndInfer/inference_mnist.py"
	UploadFile(filePath)
	filePath = "/home/public/workspace/trainAndInfer/upload.py"
	UploadFile(filePath)
	filePath = "/home/public/goprojects/Registry/Registry/cmd/registry/tmp/data/mnist_cnn.pt_v1.0.0"
	DeleteFile(filePath) // 删除上一轮跑的上传的模型
	filePath = "/home/public/goprojects/test-0623/test/nodelet/task_exporter/dependency/requirements.txt"
	UploadFile(filePath)
	//filePath = "/home/public/workspace/trainAndInfer/dataset"
	//UploadFiles(filePath)

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
			DataDependencyConditionFormula,
		},
	}
	runtime1_1_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
			DataDependencyConditionFormula,
		},
	}
	runtime1_2_1_1Condition := apis.Conditions{
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
			},
		},
		Name:    group1_1Name,
		Parents: make([]string, 0),
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:       runtime1_1_1_1Name,
						Type:       apis.ByCommand,
						Command:    []string{"python"},
						Args:       []string{"train_mnist.py"},                             // 10s
						Parents:    make([]string, 0),                                      // 加入Parents
						Data:       []apis.DataSpec{apis.DataSpec{Name: "train_mnist.py"}}, // 依赖文件
						Conditions: &runtime1_1_1_1Condition,
						Outputs:    []apis.Value{apis.Value{Type: apis.LocalData, Value: "mnist_cnn.pt", ValueType: apis.StringType}},
					},
					apis.RuntimeSpec{
						Name:       runtime1_1_1_2Name,
						Type:       apis.ByCommand,
						Command:    []string{"python"},
						Args:       []string{"upload.py"},
						Parents:    []string{runtime1_1_1_1Name},                      // 加入Parents
						Data:       []apis.DataSpec{apis.DataSpec{Name: "upload.py"}}, // 依赖文件
						Conditions: &runtime1_1_1_2Condition},
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
			},
		},
		Name:    group1_2Name,
		Parents: []string{group1_1Name},
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_2_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:       runtime1_2_1_1Name,
						Type:       apis.ByCommand,
						Inputs:     []apis.Value{apis.Value{Type: apis.LocalData, From: "Group{G81}.Action{A1}.Runtime{R1}.Outputs{0}"}},
						Command:    []string{"python"},
						Args:       []string{"inference_mnist.py"},                                                                  // 10s
						Data:       []apis.DataSpec{apis.DataSpec{Name: "inference_mnist.py"}, apis.DataSpec{Name: "mnist_cnn.pt"}}, // 依赖文件
						Conditions: &runtime1_2_1_1Condition,
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
	_, err1 := analyzer.SerializeToJson(task)
	if err1 != nil {
		return
	}
	//fmt.Println(str)
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

// 上传文件
func UploadFile(filePath string) (string, error) {
	// 调用 utils.UploadFile 函数上传文件
	// 这里的 filePath 是要上传的文件路径
	// 返回上传结果和错误信息
	url := "http://localhost:8919/upload"
	err := utils.UploadFile(filePath, "v1.0.0", url)
	if err != nil {
		fmt.Println("Upload failed:", err)
		return "", err
	} else {
		fmt.Println("Upload successful!")
		return "Upload successful!", nil
	}
}
func UploadFiles(filePath string) (string, error) {
	url := "http://localhost:8919/upload?filename=" + filePath
	err := utils.Traverse(filePath, url)
	if err != nil {
		fmt.Println("Upload failed:", err)
		return "", err
	} else {
		fmt.Println("Upload successful!")
		return "Upload successful!", nil
	}
}

func DeleteFile(filePath string) (string, error) {
	// 1. 验证文件路径是否有效
	if filePath == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}
	// 2. 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("文件不存在: %s", filePath)
		}
		return "", fmt.Errorf("无法获取文件信息: %w", err)
	}
	// 3. 检查路径是文件还是目录
	if fileInfo.IsDir() {
		return "", fmt.Errorf("路径是目录而非文件: %s", filePath)
	}
	// 4. 删除文件
	err = os.Remove(filePath)
	if err != nil {
		return "", fmt.Errorf("删除文件失败: %w", err)
	}
	return fmt.Sprintf("成功删除文件: %s", filePath), nil
}
