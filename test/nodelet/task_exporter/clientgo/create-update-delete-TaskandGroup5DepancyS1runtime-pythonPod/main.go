package main

import (
	"bufio"
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// 创建一个Rest Client-123
// 验证xxx动词
// 与API Server通信，并执行基础操作
// 1个group，1个Action，每个Action一个Runtime， 一共1个Runtime，为Pod
func randomSuffix(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
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

	tasksClient := clientSet.Core().Tasks("test")
	groupsClient := clientSet.Core().Groups("test")
	actionClient := clientSet.Core().Actions("test")
	eventclient := clientSet.Core().Events("test")

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := fmt.Sprintf("Task1-%s", randomSuffix(10)) // 第一个Task的Name
	task1ID := fmt.Sprintf("TaskID1-%s", randomSuffix(10)) // 第一个Task的ID

	// group
	group1_1Name := fmt.Sprintf("Group1-%s", randomSuffix(10)) // 第一个Task下的第一个GroupName
	//group1_2Name := "ReasonGroup-2"           // 第一个Task下的第二个GroupName
	//group1_3Name := "RobotDestinationGroup-3" // 第一个Task下的第三个GroupName  "RobotDestinationGroup"
	group1_1ID := fmt.Sprintf("GroupID1-%s", randomSuffix(10)) // 第一个Task下的第一个GroupID
	//group1_2ID := "GroupID-2"                 // 第一个Task下的第二个GroupID
	//group1_3ID := "GroupID-3"                 // 第一个Task下的第三个GroupID

	// action
	action1_1_1Name := fmt.Sprintf("Action1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	//action1_2_1Name := "Action2-1" // 第一个Task下的第二个Group下的第一个ActionName
	//action1_3_1Name := "Action3-1" // 第一个Task下的第三个Group下的第一个ActionName
	action1_1_1ID := fmt.Sprintf("ActionID1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionID
	//action1_2_1ID := "ActionID2-1" // 第一个Task下的第二个Group下的第一个ActionID
	//action1_3_1ID := "ActionID3-1" // 第一个Task下的第三个Group下的第一个ActionID

	// runtime
	runtime1_1_1_1Name := fmt.Sprintf("Runtime1-1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_1_1_2Name := "Runtime1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_2_1_1Name := "Runtime2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_2_1_2Name := "Runtime2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_3_1_1Name := "Runtime3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_3_1_2Name := "Runtime3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeName

	runtime1_1_1_1ID := fmt.Sprintf("RuntimeID1-1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_1_1_2ID := "RuntimeID1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_2_1_1ID := "RuntimeID2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_2_1_2ID := "RuntimeID2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_3_1_1ID := "RuntimeID3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_3_1_2ID := "RuntimeID3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeID

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false
	//runtime1_1_1_2FineGrainedControl := true
	//runtime1_2_1_1FineGrainedControl := false
	//runtime1_2_1_2FineGrainedControl := false
	//runtime1_3_1_1FineGrainedControl := false
	//runtime1_3_1_2FineGrainedControl := false
	// 统一地规定： Belongs：填的是ID
	//            Parents: 填的也是ID吧--改为Name
	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}
	podName := "python-pod" // fmt.Sprintf("python-pod1-%s", randomSuffix(10))
	podNamespace := "switch"

	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_1Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:     group1_1Name,
			Parents:  make([]string, 0),
			Replicas: []int32{0, 0},
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_1_1Name},
					Spec: apis.ActionSpec{
						Name: action1_1_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{ // nodeSelector版本
								Name:                     runtime1_1_1_1Name,
								Type:                     apis.ByPod,
								Command:                  []string{},
								Args:                     []string{},
								Parents:                  []string{runtime1_1_1_1Name}, // 加入Parents
								Conditions:               runtime1_1_1_1Condition,
								Image:                    "",
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
								Pod: apis.Pod{
									TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
									ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: podNamespace},
									Spec: apis.PodSpec{
										RestartPolicy: apis.RestartPolicyNever,
										NodeSelector:  map[string]string{"kubernetes.io/hostname": "server2"},
										Containers: []apis.Container{
											apis.Container{
												Name:  podName,
												Image: "crpi-f433c61sz0xmdl3o.cn-hangzhou.personal.cr.aliyuncs.com/hit-kcm/mnist",
												Ports: []apis.ContainerPort{
													apis.ContainerPort{
														ContainerPort: 50051,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_1_1ID, // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_1_1ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: group1_1ID,
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: action1_1_1ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_1_1_1ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group1 := &g1

	task := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      task1Name,
			Namespace: "",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: task1Name,
			Groups: []apis.Group{
				g1,
			},
		},
		Status: apis.TaskStatus{
			TaskID: task1ID,
			Phase:  apis.Unknown,
			GroupStatus: []apis.GroupStatus{
				apis.GroupStatus{
					GroupID: group1_1ID,
					ActionStatus: []apis.ActionStatus{
						apis.ActionStatus{
							ActionID: action1_1_1ID,
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: runtime1_1_1_1ID,
									Phase:     apis.Unknown,
								},
								//apis.RuntimeStatus{
								//	RuntimeID: runtime1_1_1_2ID,
								//	Phase:     apis.Unknown,
								//},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: task1ID},
					Phase:   apis.Unknown,
				},
			},
		},
	}

	// 删除事件
	logs.Info("delete events")
	eventList, err := eventclient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range eventList.Items {
		err := eventclient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	// 删除Task
	logs.Info("delete tasks")
	taskList, err := tasksClient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range taskList.Items {
		err := tasksClient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	// 删除Group
	logs.Info("delete group")
	groupList, err := groupsClient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range groupList.Items {
		err := groupsClient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	// 删除Action
	logs.Info("delete group")
	actionList, err := actionClient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range actionList.Items {
		err := actionClient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	//如果已经存在，先删掉
	clientset := config.LoadConfig()
	DeletePod(clientset, podName, podNamespace)

	// Create一个Task
	logs.Infof("creating")
	_, err = tasksClient.Create(context.TODO(), task, metav1.CreateOptions{})
	_, err1 := groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})
	//results2, err2 := groupsClient.Create(context.TODO(), group2, metav1.CreateOptions{})
	//results3, err3 := groupsClient.Create(context.TODO(), group3, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	if err1 != nil {
		logs.Errorf("Failed to create group1: %v", err1)
		panic(err)
	}
	prompt()
	// 删除事件
	// 删除事件
	logs.Info("delete events")
	eventList, err = eventclient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range eventList.Items {
		err := eventclient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	// 删除Task
	logs.Info("delete tasks")
	taskList, err = tasksClient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range taskList.Items {
		err := tasksClient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	// 删除Group
	logs.Info("delete group")
	groupList, err = groupsClient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range groupList.Items {
		err := groupsClient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	// 删除Action
	logs.Info("delete group")
	actionList, err = actionClient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range actionList.Items {
		err := actionClient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}
	//如果已经存在，先删掉
	DeletePod(clientset, podName, podNamespace)

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
func DeletePod(clientset *kubernetes.Clientset, podName, podNamespace string) {
	// 使用与创建时一致的日志记录风格
	err := clientset.CoreV1().Pods(podNamespace).Delete(
		context.TODO(),
		podName, // 直接从Pod对象获取名称
		k8smetav1.DeleteOptions{
			GracePeriodSeconds: pointer.Int64Ptr(5), // 可选：优雅删除等待时间
		},
	)

	if err != nil {
		// 带上下文的错误日志，保持与你的风格一致
		logs.Error(err, "删除Pod失败", "Pod名称", podName, "命名空间", podNamespace)
	} else {
		// 成功日志包含结构化参数
		logs.Info("Pod删除成功",
			"Pod名称", podName,
			"命名空间", podNamespace,
			"删除时间", time.Now().Format(time.RFC3339))
	}

	// 如果EM需要清理，可以在此处调用
	// EM.GetInstance().RemovePod(pod.Namespace, pod.Name)
}
