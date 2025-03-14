package main

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task"
	"time"
)

// 打开yolo训练任务的Groups信息-适用于Windows
func yoloTrainTaskGroupInWindows() []*apis.Group {
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
								Args:    []string{"D:\\Programming\\GoLand\\goProject\\new2-task\\resourcelet\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\train.py"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-train-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// 打开yolo推理任务的Groups信息-适用于Windows
func yoloPredictTaskGroupInWindows() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_predict"},
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
			GroupID: "test-predict-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// 打开yolo训练任务的Groups信息-适用于Linux
func yoloTrainTaskGroupInLinux() []*apis.Group {
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
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-tarin-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// 打开yolo推理任务的Groups信息-适用于Linux
func yoloPredictTaskGroupInLinux() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_predict"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:  "CMD",
								Image: "",
								Type:  apis.ByCommand,
								// Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Command: []string{"/home/ysk/miniconda3/envs/py38/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-predict-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// 将推理所需的数据压缩到指定的位置上
func yoloCompressValidDataGroupInLinux() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_compress_valid"},
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
								Command: []string{"bash"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/pack_files-2.sh"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "compress-valid-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func main() {
	//模拟的任务信息
	groupInfos1 := yoloTrainTaskGroupInLinux()
	groupInfos2 := yoloPredictTaskGroupInLinux()
	//groupInfos2 := closeYoloTrainTaskGroup()
	taskexporter, err := task.NewTaskExporter(nil)
	if err != nil {
		logs.Error("fail to create task exporter")
	}
	go func() {
		err2 := taskexporter.Run(context.Background())
		if err2 != nil {
			logs.Error("fail to run task exporter")
		}
	}()

	time.Sleep(2 * time.Second)
	task.ReceiveGroupInfo(groupInfos1, "create")

	time.Sleep(60 * time.Second)
	task.ReceiveGroupInfo(groupInfos1, "kill")

	time.Sleep(5 * time.Second)
	task.ReceiveGroupInfo(groupInfos2, "create")
	select {}
}
