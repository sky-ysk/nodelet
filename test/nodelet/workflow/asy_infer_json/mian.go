package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
)

var scheme = runtime.NewScheme()

const NodeName = "cloudNode1" // 1$
var group1_1Name = "G91"      // 第一个Task下的第一个GroupName
var namespace = "test"
var filefold = "yolo_projects_asy"
var filename = "yolo_projects_asy/yolo-asy-running1.py"
var command = "python"
var require = "yolo_projects_asy/requirements.txt"
var folderPath = "/home/public/workspace/yolo_projects_asy"

// 定义JSON解析结构体
type AppConfig struct {
	Data []struct {
		Name string `json:"name"`
	} `json:"data"`
	Type          string   `json:"type"`
	Command       []string `json:"command"`
	Args          []string `json:"args"`
	EnableControl bool     `json:"enable_control"`
	Conditions    struct {
		Formulas []struct {
			ConditionType string `json:"condition_type"`
			From          string `json:"from"`
		} `json:"formulas"`
	} `json:"conditions"`
	HasReplca bool `json:"hasReplca"`
}

// 测试切换
// 1个group，1个Action，每个Action1个Runtime， 一共1个Runtime
func main() {
	// 第一步：处理命令行参数
	args := os.Args
	if len(args) > 1 {
		folderPath = args[1] // 使用第一个命令行参数覆盖folderPath
	}

	// 第二步：解析application.json
	configFile := "application.json" // JSON文件路径
	if len(args) > 2 {
		configFile = args[2] // 可选的第二个参数指定JSON路径
	}

	jsonData, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Error reading JSON file: %v", err))
	}

	var config AppConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		panic(fmt.Sprintf("Error parsing JSON: %v", err))
	}

	// 第三步：从JSON提取参数
	if len(config.Data) > 0 {
		filefold = config.Data[0].Name // 映射 data[0].name
	}
	if len(config.Command) > 0 {
		command = config.Command[0] // 映射 command[0]
	}
	if len(config.Args) > 0 {
		// 直接使用整个参数字符串作为filename
		filename = config.Args[0]
	}
	if len(config.Conditions.Formulas) > 0 {
		require = config.Conditions.Formulas[0].From // 映射 conditions.formulas[0].from
	}

	// 如果命令行未提供folderPath，使用默认值
	if folderPath == "" {
		fmt.Println("foldPath is null,please add")
		return
	}
	moduleName := "testModule"
	logs.Init(moduleName)

	//创建ClientSet
	clientSet := initClientSet(scheme)

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T1" // 第一个Task的Name

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := true

	// 程序依赖（requirements.txt）
	ProgramDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.ProgramDependency,
		LeftValue: apis.Value{
			From: require, //2$
		},
	}
	// 数据依赖（../tmp/testFolder）
	DataDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.DataDependency,
		LeftValue: apis.Value{
			Type: apis.FileData,
		},
	}

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
			DataDependencyConditionFormula,
		},
	}

	gs1 := apis.GroupSpec{
		//Replicas: group1_1Replicas,
		HasReplca: true,
		Name:      group1_1Name,
		Parents:   make([]string, 0),
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:    runtime1_1_1_1Name,
						Type:    apis.ByCommand,
						Command: []string{command},
						Args:    []string{filename},
						Data:    []apis.DataSpec{apis.DataSpec{Name: filefold}},
						// Data:                     []apis.DataSpec{{Name: "requirements.txt"}},
						Conditions:               &runtime1_1_1_1Condition,
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
				},
			},
		},
	}

	_, err = UploadFolder(folderPath) //---这个方法改传递文件夹
	if err != nil {
		fmt.Println("upload folder err!")
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
	_, err = m.CreateTask(ts, nil, namespace, u.String(), "")
	if err != nil {
		panic(err)
	}
	fmt.Println("Application description submitted successfully")
	//str, err := analyzer.SerializeToJson(task)
	//if err != nil {
	//	return
	//}
	//fmt.Println(str)

	prompt()
	postEventForMigrate_ForGroup()
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

func postEventForMigrate_ForGroup() {
	logs.Info("发送迁移事件======")
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
