package main

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/lib"
	"hit.edu/framework/pkg/scheduler/utils"
	"time"
)

func main() {
	logs.Init("test")
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)
	//deviceList, err := m.GetDevices("", apis.NamespaceTest)
	device, err := m.GetDevice("device2", apis.NamespaceTest)
	if err != nil {
		logs.Errorf("get device failed, err:%v", err)
	}
	// 构造URL
	ip := device.Status.Abilities["Grab"].Services["GrabInit"].Ip
	port := device.Status.Abilities["Grab"].Services["GrabInit"].Port
	api := device.Status.Abilities["Grab"].Services["GrabInit"].Interface
	url := fmt.Sprintf("http://%s:%s%s", *ip, *port, *api)

	taskId := ""
	taskId, err = lib.PublishGrabInitInst("up", url)
	if err != nil {
		logs.Errorf("publish grab init inst failed, err:%v", err)
		return
	}
	time.Sleep(5 * time.Second)
	for {
		time.Sleep(1 * time.Second)
		resp, err := lib.GetTaskStatus(taskId, device.Spec.AccessMethod.URL)
		if err != nil {
			logs.Errorf("get task status failed, err:%v", err)
			return
		}
		if resp.State == lib.Running {
			logs.Infof("grab init is running")
		} else if resp.State == lib.Finished {
			logs.Infof("device:%s, grab init is finished", device.Name)
			break
		}
	}
	logs.Infof("all init finished")
}
