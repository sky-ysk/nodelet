package main

import (
	"bufio"
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
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
		Host:    "http://localhost:10000", //http://suda801.wangwanu.com:10000
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
	groupsClient := clientSet.Core().Groups("")
	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: "TrainGroup", Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    "TrainGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: "cmd_yolo_train_action"},
					Spec: apis.ActionSpec{
						Name: "cmd_yolo_train_action",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_train_action:TrainGroup:test-task", // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "CMD:cmd_yolo_train_action:TrainGroup:test-task", // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "TrainGroup:test-task",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_train_action:TrainGroup:test-task",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "CMD:cmd_yolo_train_action:TrainGroup:test-task",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: "test-task"},
		},
	}
	group1 := &g1
	g2 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: "ReasonGroup", Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    "ReasonGroup",
			Parents: []string{"TrainGroup"}, // 加入Parents
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: "cmd_yolo_predict_action"},
					Spec: apis.ActionSpec{
						Name: "cmd_yolo_predict_action",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "ABC",
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
							},
							apis.Runtime{
								Name:    "XYZ",
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"//home/public/workspace/heongtong_yolo_linux/pull_robot.py"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_predict_action:ReasonGroup:test-task",
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "ABC:cmd_yolo_predict_action:ReasonGroup:test-task",
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: "XYZ:cmd_yolo_predict_action:ReasonGroup:test-task",
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "ReasonGroup:test-task",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_predict_action:ReasonGroup:test-task",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "ABC:cmd_yolo_predict_action:ReasonGroup:test-task",
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: "XYZ:cmd_yolo_predict_action:ReasonGroup:test-task",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: "test-task"},
		},
	}
	group2 := &g2
	task := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TrainInferTask",
			Namespace: "",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: "TrainInferTask",
			Groups: []apis.Group{
				g1, g2,
			},
		},
		Status: apis.TaskStatus{
			TaskID: "test-task",
			Phase:  apis.Unknown,
			GroupStatus: []apis.GroupStatus{
				apis.GroupStatus{
					GroupID: "TrainGroup:test-task",
					ActionStatus: []apis.ActionStatus{
						apis.ActionStatus{
							ActionID: "cmd_yolo_train_action:TrainGroup:test-task",
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: "CMD:cmd_yolo_train_action:TrainGroup:test-task",
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: "test-task"},
				},
				apis.GroupStatus{
					GroupID: "ReasonGroup:test-task",
					ActionStatus: []apis.ActionStatus{
						apis.ActionStatus{
							ActionID: "cmd_yolo_predict_action:ReasonGroup:test-task",
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: "ABC:cmd_yolo_predict_action:ReasonGroup:test-task",
									Phase:     apis.Unknown,
								},
								apis.RuntimeStatus{
									RuntimeID: "XYZ:cmd_yolo_predict_action:ReasonGroup:test-task",
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: "test-task"},
				},
			},
		},
	}
	//task := &apis.Task{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "TrainInferTask",
	//		Namespace: "",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Task",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.TaskSpec{
	//		Name:       "TrainInferTask",
	//		Desc:       apis.Description{},
	//		Conditions: apis.Conditions{},
	//		Groups: []apis.Group{
	//			apis.Group{
	//				ObjectMeta: metav1.ObjectMeta{Name: "demo-group1", Namespace: ""},
	//				Spec: apis.GroupSpec{
	//					Name: "demo-groups",
	//				},
	//				Status: apis.GroupStatus{},
	//			},
	//			apis.Group{
	//				ObjectMeta: metav1.ObjectMeta{Name: "demo-group2", Namespace: ""},
	//				Spec: apis.GroupSpec{
	//					Name: "demo-groups2",
	//				},
	//				Status: apis.GroupStatus{},
	//			},
	//		},
	//	},
	//	Status: apis.TaskStatus{
	//		Belongs: apis.IDRef{},
	//	},
	//}

	//task2 := &apis.Task{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "TrainInferTask2",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Task",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.TaskSpec{
	//		TaskName: "TrainInferTask",
	//	},
	//}
	//task3 := &apis.Task{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "TrainInferTask3",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Task",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.TaskSpec{
	//		TaskName: "TrainInferTask",
	//	},
	//}
	//patchTask, err := json.Marshal(map[string]interface{}{
	//	"Spec": map[string]interface{}{
	//		"Name": "patch-task-name",
	//	},
	//})

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
	////err = tasksClient.Delete(context.TODO(), "TrainInferTask", metav1.DeleteOptions{})

	err = tasksClient.Delete(context.TODO(), "TrainInferTask", metav1.DeleteOptions{})
	err1 := groupsClient.Delete(context.TODO(), "TrainGroup", metav1.DeleteOptions{})
	err2 := groupsClient.Delete(context.TODO(), "ReasonGroup", metav1.DeleteOptions{})
	// Create一个Task
	logs.Infof("creating")
	results, err := tasksClient.Create(context.TODO(), task, metav1.CreateOptions{})
	results2, err1 := groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})
	results3, err2 := groupsClient.Create(context.TODO(), group2, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	if err1 != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	if err2 != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	//_, _ = tasksClient.Create(context.TODO(), task2, metav1.CreateOptions{})
	//_, _ = tasksClient.Create(context.TODO(), task3, metav1.CreateOptions{})
	logs.Infof("Created task ", results)
	logs.Infof("Created group1 ", results2)
	logs.Infof("Created group2 ", results3)
	//prompt()

	//Update一个Task

	//logs.Info("updating")
	//// 部分更改一个参数
	//// 先Get一个Task ,更改Task的参数, UpdateTask
	//
	//result, getErr := tasksClient.Get(context.TODO(), "TrainInferTask", metav1.GetOptions{})
	//if getErr != nil {
	//	panic(fmt.Errorf("Failed to get : %v", getErr))
	//}
	//
	//logs.Infof("get result", result)
	//logs.Infof("修改前的result.Spec.Name：", result.Spec.Name)
	//
	//result.Spec.Name = "updatedTaskName"
	//_, updateErr := tasksClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	//if updateErr != nil {
	//	panic(fmt.Errorf("Update failed: %v", updateErr))
	//}
	//
	//logs.Infof("修改后的result.Spec.Name：", result.Spec.Name)
	//logs.Info("Updated task...")
	//prompt()

	//// List 所有Task
	//logs.Info("listing")
	//lstOpts := metav1.ListOptions{}
	//list, err := tasksClient.List(context.TODO(), lstOpts)
	//list1, err1 := groupsClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//if err1 != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Info(d)
	//}
	//for _, d := range list1.Items {
	//	logs.Info(d)
	//}
	//logs.Infof("listing done")
	//prompt()

	////Patch 一个Task
	//logs.Infof("patching")
	//patchResult, err := tasksClient.Patch(context.TODO(), "TrainInferTask", types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
	//logs.Infof("patchResult: ", patchResult)
	//logs.Infof("patch Done")
	//
	//// List 所有Task
	//logs.Info("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = tasksClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Info(d)
	//}
	//
	//logs.Infof("listing done")
	//prompt()

	// Delete一个Task 和两个Group
	//logs.Info("deleting")
	//err = tasksClient.Delete(context.TODO(), "TrainInferTask", metav1.DeleteOptions{})
	//err1 = groupsClient.Delete(context.TODO(), "demo-group1", metav1.DeleteOptions{})
	//err2 = groupsClient.Delete(context.TODO(), "demo-group2", metav1.DeleteOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//if err1 != nil {
	//	panic(err)
	//}
	//if err2 != nil {
	//	panic(err)
	//}
	//logs.Info("Deleted task...")
	//logs.Info("Deleted group1、group2...")
	//prompt()
	//
	//// Delete 之后再次 List所有Task
	//logs.Info("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = tasksClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Info(d)
	//}
	//
	//logs.Infof("listing done")
	//
	//select {}

	////DeleteCollection 删除所有Spec.TaskName=TrainInferTask的Task
	//logs.Infof("deleting collection")
	//lstOpts = metav1.ListOptions{
	//	FieldSelector: "Spec.TaskName=TrainInferTask",
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
