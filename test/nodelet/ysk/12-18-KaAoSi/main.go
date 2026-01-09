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
	namespace := "Cosmo"
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
	group1_1Name := "G84" // 第一个Task下的第一个GroupName
	group1_2Name := "G85" // 第一个Task下的第二个GroupName
	group1_1Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_1_2Name := "A2" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_1_3Name := "A3" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_1_4Name := "A4" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_1_5Name := "A5" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_1_6Name := "A6" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"

	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false

	// action_Condition := apis.Conditions{
	// 	Formulas: []apis.ConditionFormula{
	// 		apis.ConditionFormula{
	// 			ConditionType: apis.DataDependency,
	// 			LeftValue:     apis.Value{Type: apis.LocalData, Value: "", From: "Action{A1}.Status{}"},
	// 			RightValue:    apis.Value{Value: "Running"},
	// 		},
	// 		apis.ConditionFormula{
	// 			ConditionType: apis.DataDependency,
	// 			LeftValue:     apis.Value{Type: apis.LocalData, Value: "", From: "Action{A2}.Status{}"},
	// 			RightValue:    apis.Value{Value: "Running"},
	// 		},
	// 	},
	// }
	runtime_condition_2 := apis.Conditions{Formulas: []apis.ConditionFormula{{ConditionType: apis.NodeDependency, LeftValue: apis.Value{NameSpace: namespace, From: "Group{" + group1_1Name + "}.Action{" + action1_1_1Name + "}.Runtime{" + runtime1_1_1_1Name + "}.Status{}"}}}}
	runtime_condition_3 := apis.Conditions{Formulas: []apis.ConditionFormula{{ConditionType: apis.NodeDependency, LeftValue: apis.Value{NameSpace: namespace, From: "Group{" + group1_1Name + "}.Action{" + action1_1_2Name + "}.Runtime{" + runtime1_1_1_1Name + "}.Status{}"}}}}
	runtime_condition_5 := apis.Conditions{Formulas: []apis.ConditionFormula{{ConditionType: apis.NodeDependency, LeftValue: apis.Value{NameSpace: namespace, From: "Group{" + group1_2Name + "}.Action{" + action1_1_4Name + "}.Runtime{" + runtime1_1_1_1Name + "}.Status{}"}}}}
	runtime_condition_6 := apis.Conditions{Formulas: []apis.ConditionFormula{{ConditionType: apis.NodeDependency, LeftValue: apis.Value{NameSpace: namespace, From: "Group{" + group1_2Name + "}.Action{" + action1_1_5Name + "}.Runtime{" + runtime1_1_1_1Name + "}.Status{}"}}}}
	action_Condition := apis.Conditions{}

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
		},
		Parents:    make([]string, 0),
		Conditions: &group1_1Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-it", "-d", "--name", "opcuanode", "-e", "RMW_IMPLEMENTATION=rmw_fastrtps_cpp", "-e", "ROS_DOMAIN_ID=41", "-e", "RMW_FASTRTPS_LOG_VERBOSITY=DEBUG", "-v", "/dev/shm:/dev/shm", "--ipc=host", "--network=host", "--restart", "always", "--entrypoint", "/ros_entrypoint.sh", "registry2-qingdao.cosmoplat.com/62_flexassemble/opcuadriver:v1.1"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               nil,
						EnvVar:                   []apis.EnvVar{},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
				},
			},
			apis.ActionSpec{
				Name: action1_1_2Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-it", "-d", "--name", "rosbridgeserver", "-e", "RMW_IMPLEMENTATION=rmw_fastrtps_cpp", "-e", "ROS_DOMAIN_ID=41", "-e", "RMW_FASTRTPS_LOG_VERBOSITY=DEBUG", "-v", "/dev/shm:/dev/shm", "--ipc=host", "--network=host", "--restart", "always", "--entrypoint", "/ros_entrypoint.sh", "registry2-qingdao.cosmoplat.com/62_flexassemble/rosbridge:v1.0"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               &runtime_condition_2,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
				},
			},
			apis.ActionSpec{
				Name:       action1_1_3Name,
				Parents:    []string{},
				Conditions: &action_Condition,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"bash"},
						Args:                     []string{"/home/cosmo/wangkun-projects/ts_platform/run_all.sh"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               &runtime_condition_3,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
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
		Desc: &apis.Description{
			Label: map[string]string{
				"scheduler": "CloudNode1",
			},
		},
		Parents:    make([]string, 0),
		Conditions: &group1_1Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name:       action1_1_4Name,
				Parents:    []string{},
				Conditions: &action_Condition,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "--entrypoint", "/2f_entrypoint.sh", "-d", "--network=host", "--name", "control_v9", "-e", "RMW_FASTRTPS_LOG_VERBOSITY=DEBUG", "-v", "/dev/shm:/dev/shm", "--ipc=host", "-e", "RMW_IMPLEMENTATION=rmw_fastrtps_cpp", "-e", "ROS_DOMAIN_ID=41", "-it", "control_v9"},
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
			apis.ActionSpec{
				Name:       action1_1_5Name,
				Parents:    []string{},
				Conditions: &action_Condition,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByDocker,
						Command:                  []string{"docker"},
						Args:                     []string{"run", "-itd", "--name", "mainflow", "--entrypoint", "bash", "--network=host", "--restart", "always", "mainflow_v1.8"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               &runtime_condition_5,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
				},
			},
			apis.ActionSpec{
				Name:       action1_1_6Name,
				Parents:    []string{},
				Conditions: &action_Condition,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"docker"},
						Args:                     []string{"exec", "-i", "mainflow", "bash", "-c", "python main1.py"},
						Inputs:                   []apis.Value{},    //20s
						Parents:                  make([]string, 0), // 加入Parents
						Data:                     []apis.DataSpec{}, // 依赖文件
						Image:                    "",
						Conditions:               &runtime_condition_6,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
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
