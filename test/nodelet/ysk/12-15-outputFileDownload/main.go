package main

import (
	"bufio"
	"context"
	"errors"
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
var namespace = "test1"

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

	group1_1Replicas := []int32{1, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_1_1_2Name := "R2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false

	// 数据依赖（../tmp/testFolder）
	DataDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.DataDependency,
		LeftValue: apis.Value{
			Type: apis.FileData,
		},
	}

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			DataDependencyConditionFormula,
		},
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
						Type:                     apis.ByCommand,
						Command:                  []string{"bash"},
						Args:                     []string{"./2.bash"},
						Data:                     []apis.DataSpec{apis.DataSpec{Name: "2.bash", FileFormat: "file"}, apis.DataSpec{Name: "1.txt", FileFormat: "file"}},
						Conditions:               &runtime1_1_1_1Condition,
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
						Outputs:                  []apis.Value{{Type: apis.FileData, Name: "2.txt"}},
					},
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_2Name,
						Type:                     apis.ByCommand,
						Parents:                  []string{runtime1_1_1_1Name},
						Command:                  []string{"bash"},
						Args:                     []string{"./3.bash"},
						Data:                     []apis.DataSpec{apis.DataSpec{Name: "3.bash", FileFormat: "file"}, apis.DataSpec{Name: "2.txt", FileFormat: "file"}},
						Conditions:               &runtime1_1_1_1Condition,
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
						Outputs:                  []apis.Value{{Type: apis.FileData, Name: "3.txt"}},
					},
				},
			},
		},
	}
	file1 := "/home/public/goprojects/ysk-1106/reference-626/test/nodelet/ysk/12-15-outputFileDownload/2.bash"
	file2 := "/home/public/goprojects/ysk-1106/reference-626/test/nodelet/ysk/12-15-outputFileDownload/3.bash"
	file3 := "/home/public/goprojects/ysk-1106/reference-626/test/nodelet/ysk/12-15-outputFileDownload/1.txt"
	_, err := UploadFile(file1)
	if err != nil {
		fmt.Println("upload file1 err!")
	}
	_, err = UploadFile(file2)
	if err != nil {
		fmt.Println("upload file2 err!")
	}
	_, err = UploadFile(file3)
	if err != nil {
		fmt.Println("upload file2 err!")
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

func UploadFolder(folderPath string) (string, error) {
	if folderPath == "" {
		return "", errors.New("dirPath is empty")
	}
	parts := strings.Split(folderPath, "/")
	if len(parts) == 0 {
		return "", errors.New("dirPath does not contain any parts")
	}
	// 获取最后一个部分作为文件夹名称
	// 如果最后一个部分是空字符串，说明路径以/结尾，可能是一个空目录
	// 例如 "/home/user/documents/"，最后一个部分是空
	if parts[len(parts)-1] == "" {
		if len(parts) < 2 {
			return "", errors.New("dirPath does not contain a valid folder name")
		}
		// 如果是空目录，使用倒数第二个部分作为文件夹名称
		folderName := parts[len(parts)-2]
		if folderName == "" {
			return "", errors.New("folder name is empty")
		}
		// fmt.Println("Folder Name:", folderName)
	}
	folderName := parts[len(parts)-1]
	if folderName == "" {
		return "", errors.New("folder name is empty")
	}
	fmt.Print("folderPath:%s, folderName:%s", folderPath, folderName)
	url := "http://localhost:8919/upload?filename=" + folderName
	err := utils.Traverse(folderPath, url)
	if err != nil {
		fmt.Println("Upload failed:", err)
	} else {
		fmt.Println("Upload successful!")
	}
	return "", nil
}
