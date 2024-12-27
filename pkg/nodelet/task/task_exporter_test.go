package task

import (
	"context"
	"testing"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

func yoloTrainTaskGroup() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train_action"},
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Image:   "",
								Type:    apis.ByCommand,
								Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
								Args:    []string{"D:\\Programming\\GoLand\\goProject\\new2-task\\resourcelet\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\train.py"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_train_action:test-group",
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "CMD:cmd_yolo_train_action:test-group",
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_train_action:test-group",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "CMD:cmd_yolo_train_action:test-group",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}
func yoloPredictTaskGroup() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Image:   "",
								Type:    apis.ByCommand,
								Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
								Args:    []string{"D:\\Programming\\GoLand\\goProject\\new2-task\\resourcelet\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\predict.py"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func simpleTaskGroup() []*apis.Group {
	// 测试任务是否正确部署
	// 创建一个任务
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_test"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0), // 当前Group没有Parents
			Actions: []apis.Action{
				apis.Action{
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name: "CMD",
								Type: apis.ByCommand,
								Command: []string{
									// "python",
									"ls",
								},
								Args: []string{
									// "/home/ysk/Desktop/datafolder/predict.py",
									"-a",
								},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func TestTaskExporter(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	// 构造Task Exporter
	tc := NewConfig("test-node")
	te, err := NewTaskExporter(tc)
	if err != nil {
		panic(err)
	}

	//updateCh := make(chan types.GroupUpdate)
	//groupToUpdate := types.GroupUpdate{
	//	Groups: []*apis.Group{&newGroup},
	//	Op:     types.ADD,
	//}
	//go te.groupHandler.Loop(ctx, updateCh)

	// 部署一个任务
	go func() {
		err2 := te.Run(ctx)
		if err2 != nil {
			logs.Error("fail to run task exporter")
		}
	}()
	time.Sleep(2 * time.Second)
	Groups := yoloTrainTaskGroup()
	ReceiveGroupInfo(Groups, "create")
	time.Sleep(10 * time.Second)
	ReceiveGroupInfo(Groups, "kill")
	select {}
	// 部署多个任务
	//Groups[0].Spec.Name = "Test-Group2"
	//Groups[0].Status.GroupID = "Test-Group2"
	//Groups[0].Spec.Actions[0].Name = "Test-Actions"
	//Groups[0].Spec.Actions[0].Spec.Name = "Test-Actions"
	//Groups[0].Spec.Actions[0].Status.ActionID = "Test-Actions"
	//
	//time.Sleep(5 * time.Second)
	//ReceiveGroupInfo(Groups, "create")
}
