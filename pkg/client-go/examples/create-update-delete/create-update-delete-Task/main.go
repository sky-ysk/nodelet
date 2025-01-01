package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"os"
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

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
		Host:    "http://localhost:10000",
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
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Task为例，
	// 获取访问Task的客户端
	// 默认访问的Namespace是 ""

	tasksClient := clientSet.Core().Tasks("")

	//task := &apis.Task{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "demo-tasks",
	//		Namespace: "",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Task",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.TaskSpec{
	//		Name: "demo-task",
	//		Groups: []apis.Group{
	//			apis.Group{
	//				ObjectMeta: metav1.ObjectMeta{Name: "TestGroup1", Namespace: ""},
	//				TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
	//				Spec: apis.GroupSpec{
	//					Name:    "TestGroup1",
	//					Parents: make([]string, 0),
	//					Actions: []apis.Action{
	//						apis.Action{
	//							ObjectMeta: metav1.ObjectMeta{Name: "cmd_yolo_predict_action"},
	//							Spec: apis.ActionSpec{
	//								Name: "cmd_yolo_predict_action",
	//								Runtimes: []apis.Runtime{
	//									apis.Runtime{
	//										Name:    "CMD",
	//										Type:    apis.ByCommand,
	//										Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
	//										Args:    []string{"D:\\Programming\\GoLand\\goProject\\all\\adaptive-scheduling-framework\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\predict.py"},
	//									},
	//								},
	//							},
	//							Status: apis.ActionStatus{
	//								ActionID: "cmd_yolo_predict_action:TestGroup1:test-task", // ActionID =ActionName + GroupID
	//								Phase:    apis.Unknown,
	//								RuntimeStatus: []apis.RuntimeStatus{
	//									apis.RuntimeStatus{
	//										RuntimeID: "CMD:cmd_yolo_predict_action:TestGroup1:test-task", // RuntimeID = RuntimeName + ActionID
	//										Phase:     apis.Unknown,
	//									},
	//								},
	//							},
	//						},
	//					},
	//				},
	//				Status: apis.GroupStatus{
	//					GroupID: "TestGroup1:test-task",
	//					ActionStatus: []apis.ActionStatus{
	//						apis.ActionStatus{
	//							ActionID: "cmd_yolo_predict_action:TestGroup1:test-task",
	//							RuntimeStatus: []apis.RuntimeStatus{
	//								apis.RuntimeStatus{
	//									RuntimeID: "CMD:cmd_yolo_predict_action:TestGroup1:test-task",
	//									Phase:     apis.Unknown,
	//								},
	//							},
	//							Phase: apis.Unknown,
	//						},
	//					},
	//				},
	//			},
	//			apis.Group{
	//				ObjectMeta: metav1.ObjectMeta{Name: "TestGroup2", Namespace: ""},
	//				TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
	//				Spec: apis.GroupSpec{
	//					Name:    "TestGroup2",
	//					Parents: []string{"TestGroup1"}, // 加入Parents
	//					Actions: []apis.Action{
	//						apis.Action{
	//							ObjectMeta: metav1.ObjectMeta{Name: "cmd_yolo_train_action"},
	//							Spec: apis.ActionSpec{
	//								Name: "cmd_yolo_train_action",
	//								Runtimes: []apis.Runtime{
	//									apis.Runtime{
	//										Name:    "ABC",
	//										Type:    apis.ByCommand,
	//										Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
	//										Args:    []string{"D:\\Programming\\GoLand\\goProject\\all\\adaptive-scheduling-framework\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\train.py"},
	//									},
	//								},
	//							},
	//							Status: apis.ActionStatus{
	//								ActionID: "cmd_yolo_train_action:TestGroup2:test-task",
	//								Phase:    apis.Unknown,
	//								RuntimeStatus: []apis.RuntimeStatus{
	//									apis.RuntimeStatus{
	//										RuntimeID: "ABC:cmd_yolo_train_action:TestGroup2:test-task",
	//										Phase:     apis.Unknown,
	//									},
	//								},
	//							},
	//						},
	//					},
	//				},
	//				Status: apis.GroupStatus{
	//					GroupID: "TestGroup2:test-task",
	//					ActionStatus: []apis.ActionStatus{
	//						apis.ActionStatus{
	//							ActionID: "cmd_yolo_train_action:TestGroup2:test-task",
	//							RuntimeStatus: []apis.RuntimeStatus{
	//								apis.RuntimeStatus{
	//									RuntimeID: "ABC:cmd_yolo_train_action:TestGroup2:test-task",
	//									Phase:     apis.Unknown,
	//								},
	//							},
	//							Phase: apis.Unknown,
	//						},
	//					},
	//				},
	//			},
	//		},
	//	},
	//	Status: apis.TaskStatus{
	//		TaskID: "test-task",
	//		Phase:  apis.Unknown,
	//		GroupStatus: []apis.GroupStatus{
	//			apis.GroupStatus{
	//				GroupID: "TestGroup1:test-task",
	//				ActionStatus: []apis.ActionStatus{
	//					apis.ActionStatus{
	//						ActionID: "cmd_yolo_predict_action:TestGroup1:test-task",
	//						RuntimeStatus: []apis.RuntimeStatus{
	//							apis.RuntimeStatus{
	//								RuntimeID: "CMD:cmd_yolo_predict_action:TestGroup1:test-task",
	//								Phase:     apis.Unknown,
	//							},
	//						},
	//						Phase: apis.Unknown,
	//					},
	//				},
	//			},
	//			apis.GroupStatus{
	//				GroupID: "TestGroup2:test-task",
	//				ActionStatus: []apis.ActionStatus{
	//					apis.ActionStatus{
	//						ActionID: "cmd_yolo_train_action:TestGroup2:test-task",
	//						RuntimeStatus: []apis.RuntimeStatus{
	//							apis.RuntimeStatus{
	//								RuntimeID: "ABC:cmd_yolo_train_action:TestGroup2:test-task",
	//								Phase:     apis.Unknown,
	//							},
	//						},
	//						Phase: apis.Unknown,
	//					},
	//				},
	//			},
	//		},
	//	},
	//}
	task := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-tasks",
			Namespace: "",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name:       "demo-tasks",
			Desc:       apis.Description{},
			Conditions: apis.Conditions{},
			Groups: []apis.Group{
				apis.Group{
					Spec: apis.GroupSpec{
						Name: "demo-groups",
					},
					Status: apis.GroupStatus{},
				},
				apis.Group{
					Spec: apis.GroupSpec{
						Name: "demo-groups2",
					},
					Status: apis.GroupStatus{},
				},
			},
		},
		Status: apis.TaskStatus{
			Belongs: apis.IDRef{},
		},
	}
	//task2 := &apis.Task{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-task2",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Task",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.TaskSpec{
	//		TaskName: "demo-task",
	//	},
	//}
	//task3 := &apis.Task{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-task3",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Task",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.TaskSpec{
	//		TaskName: "demo-task",
	//	},
	//}
	patchTask, err := json.Marshal(map[string]interface{}{
		"Spec": map[string]interface{}{
			"Name": "patch-task-name",
		},
	})

	//监听事件并打印  监听resources/v1/tasks
	go func() {
		logs.Infof("watching")
		watchOptions := metav1.ListOptions{}

		watcher, err := tasksClient.Watch(context.TODO(), watchOptions)
		if err != nil {
			panic(err)
		}
		defer watcher.Stop() // 确保 watcher 被停止

		// 获取事件通道
		watchChan := watcher.ResultChan()

		for {
			select {
			case event, ok := <-watchChan:
				if !ok {
					logs.Infof("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				logs.Infof("接收到事件类型: %v\n", event.Type)
				switch event.Type {
				case watch.Added:
					logs.Infof("资源被添加: ", event.Object)
				case watch.Modified:
					logs.Infof("资源被修改: ", event.Object)
				case watch.Deleted:
					logs.Infof("资源被删除: ", event.Object)
				case watch.Error:
					logs.Infof("发生错误: ", event.Object)
				default:
					logs.Infof("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()

	//如果已经存在，先删掉
	//err = tasksClient.Delete(context.TODO(), "demo-tasks", metav1.DeleteOptions{})

	// Create一个Task
	logs.Infof("creating")
	results, err := tasksClient.Create(context.TODO(), task, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	//_, _ = tasksClient.Create(context.TODO(), task2, metav1.CreateOptions{})
	//_, _ = tasksClient.Create(context.TODO(), task3, metav1.CreateOptions{})
	logs.Infof("Created task ", results)

	prompt()

	//Update一个Task

	logs.Info("updating")
	// 部分更改一个参数
	// 先Get一个Task ,更改Task的参数, UpdateTask

	result, getErr := tasksClient.Get(context.TODO(), "demo-tasks", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}

	logs.Infof("get result", result)
	logs.Infof("修改前的result.Spec.Name：", result.Spec.Name)

	result.Spec.Name = "updatedTaskName"
	_, updateErr := tasksClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		panic(fmt.Errorf("Update failed: %v", updateErr))
	}

	logs.Infof("修改后的result.Spec.Name：", result.Spec.Name)
	logs.Info("Updated task...")
	prompt()

	// List 所有Task
	logs.Info("listing")
	lstOpts := metav1.ListOptions{}
	list, err := tasksClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Infof("listing done")
	prompt()

	//Patch 一个Task
	logs.Infof("patching")
	patchResult, err := tasksClient.Patch(context.TODO(), "demo-tasks", types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
	logs.Infof("patchResult: ", patchResult)
	logs.Infof("patch Done")

	// List 所有Task
	logs.Info("listing")
	lstOpts = metav1.ListOptions{}
	list, err = tasksClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Infof("listing done")
	prompt()

	// Delete一个Task
	logs.Info("deleting")
	err = tasksClient.Delete(context.TODO(), "demo-tasks", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	logs.Info("Deleted task...")
	prompt()

	// Delete 之后再次 List所有Task
	logs.Info("listing")
	lstOpts = metav1.ListOptions{}
	list, err = tasksClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Infof("listing done")

	select {}

	////DeleteCollection 删除所有Spec.TaskName=demo-task的Task
	//logs.Infof("deleting collection")
	//lstOpts = metav1.ListOptions{
	//	FieldSelector: "Spec.TaskName=demo-task",
	//}
	//err = tasksClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//logs.Infof("Deleted collection...")
	//prompt()
	//
	//// DeleteCollection 之后再次 List所有Task
	//logs.Infof("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = tasksClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Infof(d)
	//}
	//logs.Infof("listing done")
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
