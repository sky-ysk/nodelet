package main

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task"
	"time"
)

// 拉取模型并进行推理-linux
func pullModelforPredictInLinux() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_pull_predict"},
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
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/pull_predict.py", "-v", "/home/public/workspace/heongtong_yolo_linux/*.jpg"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-pull-predict-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// 拉取模型并进行训练-linux
func pullModelforTrainInLinux() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_pull_train"},
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
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/pull_train.py"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-pull-train-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// 压缩数据文件
func compressDataSetInLinux() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "compress_data_sets"}, //暂时也得唯一
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
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/pack_files.sh"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-pull-train-group", //得唯一
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// 拉取模型并进行推理-linux
func pullModelAndDataforPredictInLinux() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_pull_predict"},
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
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/pull_all_predict.py", "-v", "/home/public/workspace/heongtong_yolo_linux/*.jpg"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-pull-predict-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func main() {
	//模拟的任务信息
	groupInfos1 := pullModelforPredictInLinux()
	groupInfos2 := pullModelforTrainInLinux()
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

	time.Sleep(30 * time.Second)
	task.ReceiveGroupInfo(groupInfos2, "create")

	select {}
}
