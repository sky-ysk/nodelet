package device

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/lib"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/rmf"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/utils"
	"time"
)

type DeviceRuntime struct {
	deviceClient core.DeviceInterface
	actionClient core.ActionInterface
	eventBus     *eventbus.EventBus
}

func NewDeviceRuntime(deviceClient core.DeviceInterface, actionClient core.ActionInterface, eventBus *eventbus.EventBus) *DeviceRuntime {
	return &DeviceRuntime{
		deviceClient: deviceClient,
		actionClient: actionClient,
		eventBus:     eventBus,
	}
}

func (dr *DeviceRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {

	// 获取Devices和executor
	logs.Infof("[DEVICE RUNTIME] Try to Obtain All Devices")
	deviceMap := make(map[string]*apis.Device)
	deviceSpecList := runtime.Spec.Devices
	var executor *apis.Device
	for name, obj := range runtime.Status.Devices {
		device, err := dr.deviceClient.Get(context.TODO(), obj.Name, metav1.GetOptions{})
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] Get Device %s error: %v", obj.Name, err)
			return err
		}
		logs.Infof("[DEVICE RUNTIME] Get Device %s successfully", obj.Name)
		deviceMap[name] = device
		if device.Status.Label == "executor" {
			executor = device
		}
	}
	// 进行一次检查 判断executor是否存在
	if executor == nil {
		logs.Error("[DEVICE RUNTIME] No executor")
		return fmt.Errorf("[DEVICE RUNTIME] No executor")
	}

	// 检查Device

	// 构造参数
	inst := ""
	params := runtime.Spec.Inputs
	// executor 发布指令
	if executor.Spec.AccessMethod.Type == apis.AccessByAbility {
		logs.Infof("[DEVICE RUNTIME] Try to Publish Ability Inst")
		taskId, err := ability.PublishAbilityInst(inst, executor, params)
		if err != nil { // 如果发布任务失败
			logs.Errorf("[DEVICE RUNTIME] Publish Ability Inst error: %s", err.Error())
			go dr.notifyRuntimeStartPhase(group.Name, action.Spec.Name, runtime.Spec.Name, "", apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
			return err
		}
		// 如果发布任务成功
		logs.Infof("[DEVICE RUNTIME] Publish Ability Inst successfully")
		// 更新runtime的phase
		dr.notifyRuntimeStartPhase(group.Name, action.Spec.Name, runtime.Spec.Name, "", apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
		// 更新device的状态
		err = utils.UpdateDeviceRunning(runtime, action, executor, dr.deviceClient)
		// 进入监听任务的协程
		err = dr.monitorDeviceAbility(taskId, executor, inst, runtime, group.Name, action.Spec.Name)
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] Monitor Ability error: %s", err.Error())
			return err
		}
	} else if executor.Spec.AccessMethod.Type == apis.AccessByRmf {

	}

}

func (dr *DeviceRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error { // 首先获取action
	logs.Infof("this is a test for kill")
	action, err := dr.actionClient.Get(context.TODO(), a.Name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("can not get action from etcd!")
	}
	// 检查action的运行状态，只有处在running状态时才能取消
	if action.Status.Phase == apis.Running {
		// 一个runtime可能涉及到多个或一个 device 获取全部的device
		err, devices := utils.ObtainDevices(runtime, action)
		if err != nil {
			logs.Errorf("Action[%s] Runtime[%s] CheckDevice failed\n", action.Spec.Name, runtime.Name)
			return err
		}
		// 遍历全部的device
		for _, device := range devices {
			fmt.Println("this is a string")
			// device的taskId存储在output中，同时取出
			output := runtime.Outputs[0]
			if device.Spec.AccessMethod.Type == apis.AccessByAbility {
				_, err = ability.PublishAbilityInst(runtime.Image, &device, "terminate")
				if err != nil {
					logs.Errorf("Action[%s] Runtime[%s] terminate Ability failed\n", action.Spec.Name, runtime.Name)
					return err
				}

				nowTime := apis.Time{Time: time.Now()}
				dr.notifyRuntimeEndPhase(group.Name, actionIndex, runtimeIndex, apis.Killed, nowTime, nowTime)
				action.Status.Phase = apis.Killed
				group.Spec.Actions[actionIndex] = *action
				group.Status.Phase = apis.Killed

			} else if device.Spec.AccessMethod.Type == apis.AccessByRmf {
				// 发布指令
				_, err = rmf.PublishCancelTaskInstruction(device, output.Value)
				if err != nil {
					logs.Errorf("Action[%s] Runtime[%s] PublishCancelTaskInstruction failed\n", action.Spec.Name, runtime.Name)
					return err
				}
			}

			////修改Device状态
			//logs.Infof("Action[%s] Runtime[%s] recover device status start\n", action.Spec.Name, runtime.Name)
			//err = utils.RecoverDeviceStatus(action)
			//if err != nil {
			//	logs.Errorf("Action[%s] Runtime[%s] recover device status failed", action.Spec.Name, runtime.Name)
			//	return fmt.Errorf("Action[%s] Runtime[%s] recover device status failed ", action.Spec.Name, runtime.Name)
			//}
			//logs.Infof("Action[%s] Runtime[%s] recover device status is finished\n", action.Spec.Name, runtime.Name)
			//
			////修改Resource状态
			//logs.Infof("Action[%s] Runtime[%s] recover resource status start\n", action.Spec.Name, runtime.Name)
			//err = utils.RecoverResourceStatus(runtime, action)
			//if err != nil {
			//	logs.Errorf("Action[%s] Runtime[%s] recover status failed", action.Spec.Name, runtime.Name)
			//	return fmt.Errorf("Action[%s] Runtime[%s] recover status failed ", action.Spec.Name, runtime.Name)
			//}
			//logs.Infof("Action[%s] Runtime[%s] recover resource status  is finished\n", action.Spec.Name, runtime.Name)
			//
			////修改Scene状态
			//logs.Infof("Action[%s] Runtime[%s] recover scene status  start\n", action.Spec.Name, runtime.Name)
			//err = utils.RecoverSceneStatus(runtime, action, len(devices), action.Status.ActionID)
			//if err != nil {
			//	logs.Errorf("Action[%s] Runtime[%s] recover scene status failed", action.Spec.Name, runtime.Name)
			//	return fmt.Errorf("Action[%s] Runtime[%s] recover scene status failed ", action.Spec.Name, runtime.Name)
			//}
			//logs.Infof("Action[%s] Runtime[%s] recover scene status is finished\n", action.Spec.Name, runtime.Name)

		}

		// etcd更改

	}
	logs.Infof("device runtime kill task: %s", group.Name)
	return nil
}

func (dr *DeviceRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) notifyRuntimeEndPhase(groupName string, actionName, runtimeName string, phase apis.Phase, finishTime, lastTime apis.Time) {
	event := events.RuntimeEndPhaseEvent1{
		GroupName:       groupName,
		ActionSpecName:  actionName,
		RuntimeSpecName: runtimeName,
		Phase:           phase,
		FinishAt:        finishTime,
		LastTime:        lastTime,
	}
	dr.eventBus.Publish(event)
}

func (dr *DeviceRuntime) notifyRuntimeStartPhase(groupName string, actionName, runtimeName string, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
	event := events.RuntimeStartPhaseEvent1{
		GroupName:       groupName,
		ActionSpecName:  actionName,
		RuntimeSpecName: runtimeName,
		ProcessId:       processId,
		Phase:           phase,
		StartAt:         startAt,
		LastTime:        lastTime,
	}
	dr.eventBus.Publish(event)
}

//func (dr *DeviceRuntime) monitorDeviceRMF(action *apis.Action, groupName string, actionIndex, runtimeIndex int, runtime *apis.Runtime, taskId string, device *apis.Device) {
//	for {
//		logs.Infof("Device %s is getting task state....", device.Name)
//		tr, err := rmf.GetTaskState(device, taskId)
//		if err != nil {
//			logs.Errorf("Device %s monitor GetTaskState failed\n", device.Name)
//			return
//		}
//		switch tr.Status {
//		case "failed":
//			//TODO:错误处理
//			logs.Errorf("Device %s is failed\n", device.Name)
//			dr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
//			if err = utils.UpdateDeviceStatusFailed(runtime, action, device, dr.deviceClient); err != nil {
//				logs.Errorf(err.Error())
//			}
//			if _, err = dr.deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
//				logs.Errorf(err.Error())
//			}
//			return
//		case "completed":
//			//TODO:锁操作
//			logs.Infof("Device %s is completed\n", device.Name)
//			dr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
//			if err = utils.UpdateDeviceStatusFailed(runtime, action, device, dr.deviceClient); err != nil {
//				logs.Errorf(err.Error())
//			}
//			if _, err = dr.deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
//				logs.Errorf(err.Error())
//			}
//			return
//		case "canceled":
//			logs.Infof("Device %s is canceled\n", device.Name)
//			dr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
//			if err = utils.UpdateDeviceStatusFailed(runtime, action, device, dr.deviceClient); err != nil {
//				logs.Errorf(err.Error())
//			}
//			if _, err = dr.deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
//				logs.Errorf(err.Error())
//			}
//			return
//		default:
//			logs.Infof("Device %s is %s\n", device.Name, tr.Status)
//		}
//		time.Sleep(time.Millisecond * 500)
//	}
//
//}

func (dr *DeviceRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) string {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	//TODO implement me
	panic("implement me")
}

// monitorDeviceAbility 监测业务的执行状态
func (dr *DeviceRuntime) monitorDeviceAbility(taskId string, device *apis.Device, inst string, runtime *apis.Runtime, groupName string, actionName string) error {
	url := device.Spec.AccessMethod.URL
	for {
		time.Sleep(2 * time.Second)
		resp, err := lib.GetTaskStatus(taskId, url)
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] Get Task[%s] Status failed", taskId)
			return err
		}
		switch resp.State {
		case lib.Running: // 处于running状态
			logs.Infof("[DEVICE RUNTIME] Task[%s] is Running", taskId)
		case lib.Error: // 处于错误状态
			logs.Errorf("[DEVICE RUNTIME] Task[%s] is Error", taskId)
			go dr.notifyRuntimeEndPhase(groupName, actionName, runtime.Spec.Name, apis.Failed, apis.Time{Time: time.Now()}, apis.Time{Time: time.Now()})
		case lib.Finished: // 处于完成状态
			logs.Infof("[DEVICE RUNTIME] Task[%s] is Finished", taskId)
			// 1.处理runtime
			// 解析payload
			var outputs []apis.Value
			outputs, err = lib.ParsePayLoad(inst, resp.Payload)
			if err != nil {
				logs.Errorf("[DEVICE RUNTIME] Parse Task[%s] Payload failed", taskId)
			}
			runtime.Spec.Outputs = outputs
			// TODO etcd更新runtime的信息

			go dr.notifyRuntimeEndPhase(groupName, actionName, runtime.Spec.Name, apis.Successed, apis.Time{Time: time.Now()}, apis.Time{Time: time.Now()})
			// 2.处理device

		}
	}

}
