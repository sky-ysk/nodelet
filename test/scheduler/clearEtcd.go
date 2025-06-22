package main

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/utils"
)

func main() {
	moduleName := "testModule"
	logs.Init(moduleName)
	ctx := context.Background()
	cs, err := utils.CreateClientSetWithTimeOut(3600)
	if err != nil {
		logs.Error(err)
		return
	}

	//删group
	groupClient := cs.Core().Groups(apis.NamespaceTest)
	groups, err := groupClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, g := range groups.Items {
		logs.Infof("delete group %s ", g.Name)
		err := groupClient.Delete(ctx, g.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
			return
		}
	}

	//删actions
	actionClient := cs.Core().Actions(apis.NamespaceTest)
	acts, err := actionClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, a := range acts.Items {
		logs.Infof("delete act %s ", a.Name)
		err := actionClient.Delete(ctx, a.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}

	//删除device
	deviceClient := cs.Core().Devices("test")
	devices, err := deviceClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, d := range devices.Items {
		logs.Infof("delete act %s ", d.Name)
		err := deviceClient.Delete(ctx, d.Spec.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}

	//删events
	eventClient := cs.Core().Events(apis.NamespaceTest)
	events, err := eventClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range events.Items {
		logs.Infof("delete event %s ", e.Name)
		err := eventClient.Delete(ctx, e.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}

	//删task
	taskClient := cs.Core().Tasks(apis.NamespaceTest)
	tasks, err := taskClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, t := range tasks.Items {
		logs.Infof("delete task %s ", t.Name)
		err := taskClient.Delete(ctx, t.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}

	//删node
	//nc := cs.Core().Nodes(apis.NamespaceTest)
	//nodes, err := nc.List(ctx, metav1.ListOptions{})
	//if err != nil {
	//	return
	//}
	//for _, t := range nodes.Items {
	//	logs.Infof("delete task %s ", t.Name)
	//	err := nc.Delete(ctx, t.Name, metav1.DeleteOptions{})
	//	if err != nil {
	//		logs.Error(err)
	//		return
	//	}
	//}

	//删runtime
	rc := cs.Core().Runtimes(apis.NamespaceTest)
	runtimes, err := rc.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, t := range runtimes.Items {
		logs.Infof("delete task %s ", t.Name)
		err := rc.Delete(ctx, t.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}
}
