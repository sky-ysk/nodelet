package device

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/rmf"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/utils"
	"net/http"
	"time"
)

type DeviceRuntime struct {
	deviceClient core.DeviceInterface
	actionClient core.ActionInterface
	eventBus     *eventbus.EventBus
	lockManager  *utils.LockManager
}

func NewDeviceRuntime(deviceClient core.DeviceInterface, actionClient core.ActionInterface, eventBus *eventbus.EventBus) *DeviceRuntime {
	return &DeviceRuntime{
		deviceClient: deviceClient,
		actionClient: actionClient,
		eventBus:     eventBus,
		lockManager:  utils.NewLockManager(),
	}
}

func (dr *DeviceRuntime) Run(group *apis.Group, a *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error {

	// 获取action
	action, err := dr.actionClient.Get(context.TODO(), a.Name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("can not get action from etcd!")
	}

	// 获取runtime(从action中获取)
	runtime = &action.Spec.Runtimes[runtimeIndex]
	// 构建一个device map
	devices := make(map[string]*apis.Device)
	for _, d := range runtime.Devices {
		device, err := dr.deviceClient.Get(context.TODO(), d.Name, metav1.GetOptions{})
		if err != nil {
			logs.Errorf("can not get device: %s from etcd!", d.Name)
		}
		devices[d.Name] = device
	}

	// 检查Device
	logs.Infof("Action[%s] Runtime[%s] CheckDevice start\n", action.Spec.Name, runtime.Name)
	err = utils.CheckDevice(devices)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] CheckDevice failed\n", action.Spec.Name, runtime.Name)
		return err
	}
	logs.Infof("Action[%s] Runtime[%s] CheckDevice is successful\n", action.Spec.Name, runtime.Name)

	// 遍历device
	for name, device := range devices {
		var taskId string
		// ability方式
		if device.Spec.AccessMethod.Type == apis.AccessByAbility {
			// 构造参数传入device的expected property中
			err = utils.ConstructParamAbility(devices, runtime)
			if err != nil {
				logs.Errorf("Action[%s] Runtime[%s] construct param failed\n", action.Spec.Name, runtime.Name)
				return err
			}

			// 发布指令
			logs.Infof("publish ability inst...")
			output, err := ability.PublishAbilityInst(runtime.Image, device, "")
			if err != nil { // 指令发布失败
				logs.Errorf("Action[%s] Runtime[%s] publish Ability failed\n", action.Spec.Name, runtime.Name)
				// processId 字段保留
				dr.notifyRuntimeStartPhase(group.Name, actionIndex, runtimeIndex, "", apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
				return err
			}
			// 指令发布成功
			dr.notifyRuntimeStartPhase(group.Name, actionIndex, runtimeIndex, "", apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
			time.Sleep(1 * time.Second)
			logs.Infof("publish ability successfully")
			// 这里暂时直接调用end方法更新phase
			dr.notifyRuntimeEndPhase(group.Name, actionIndex, runtimeIndex, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
			logs.Infof("publish ability inst output:%v", output)
			// 更新runtime的output
			runtime.Outputs = append(runtime.Outputs, output)
			action.Spec.Runtimes[runtimeIndex] = *runtime

			//go dr.monitorDeviceAbility(action, group.Name, actionIndex, runtimeIndex, runtime, taskId, device, abilityManager)
		} else if device.Spec.AccessMethod.Type == apis.AccessByRmf { // rmf方式

			// 构造任务的请求参数
			logs.Infof("Action[%s] Runtime[%s] ConstructParamRMF start\n", action.Spec.Name, runtime.Name)
			err = utils.ConstructParamRMF(devices, runtime)
			if err != nil {
				logs.Errorf("Action[%s] Runtime[%s] construct param failed\n", action.Spec.Name, runtime.Name)
				return err
			}
			logs.Infof("Action[%s] Runtime[%s] ConstructParamRMF is successful\n", action.Spec.Name, runtime.Name)

			logs.Infof("device %s execute %s task\n", name, runtime.Image)

			if taskId, err = rmf.PublishAbilityInstruction(device, runtime.Image); err != nil {
				dr.notifyRuntimeStartPhase(group.Name, actionIndex, runtimeIndex, "", apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
				return err
			}
			// 没有错误就将返回信息taskId存储到runtime的Outputs中
			output := apis.Output{
				Value:     taskId,
				Name:      "task_id",
				ValueType: "string",
				Type:      apis.ResultsData,
			}

			runtime.Outputs = append(runtime.Outputs, output)

			action.Spec.Runtimes[runtimeIndex] = *runtime

			//go dr.monitorDeviceRMF(action, group.Name, actionIndex, runtimeIndex, runtime, taskId, device)
		}

		//修改Device状态
		logs.Infof("Action[%s] Runtime[%s] update device  start\n", action.Spec.Name, runtime.Name)
		err = utils.UpdateDevice(runtime, device, taskId, dr.deviceClient)
		if err != nil {
			logs.Errorf("Action[%s] Runtime[%s] update device failed\n", action.Spec.Name, runtime.Name)
		}
		logs.Infof("Action[%s] Runtime[%s] update device finished\n", action.Spec.Name, runtime.Name)

		devices[name] = device
		action.Status.Devices[name] = device.Status
	}

	_, err = dr.actionClient.Update(context.TODO(), action, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] Update action failed err:%v \n", action.Spec.Name, runtime.Name, err)
	}

	return nil
}

func (dr *DeviceRuntime) Kill(group *apis.Group, a *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error { // 首先获取action
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

				_, err = dr.actionClient.Update(context.TODO(), action, metav1.UpdateOptions{})
				// 对group进行更新

				clientset, err := InitClient()
				if err != nil {
					logs.Errorf("error")
				}
				groupClient := clientset.Core().Groups("test")
				_, err = groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
				if err != nil {
					logs.Errorf("update group fail")
				}
				logs.Infof("..")
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

func (dr *DeviceRuntime) notifyRuntimeEndPhase(groupName string, actionIndex, runtimeIndex int, phase apis.Phase, finishTime, lastTime apis.Time) {
	event := events.RuntimeEndPhaseEvent1{
		GroupName:    groupName,
		ActionIndex:  actionIndex,
		RuntimeIndex: runtimeIndex,
		Phase:        phase,
		FinishAt:     finishTime,
		LastTime:     lastTime}
	dr.eventBus.Publish(event)
}

func (dr *DeviceRuntime) notifyRuntimeStartPhase(groupName string, actionIndex, runtimeIndex int, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
	event := events.RuntimeStartPhaseEvent1{
		GroupName:    groupName,
		ActionIndex:  actionIndex,
		RuntimeIndex: runtimeIndex,
		ProcessId:    processId,
		Phase:        phase,
		StartAt:      startAt,
		LastTime:     lastTime,
	}
	dr.eventBus.Publish(event)
}

func (dr *DeviceRuntime) monitorDeviceRMF(action *apis.Action, groupName string, actionIndex, runtimeIndex int, runtime *apis.Runtime, taskId string, device *apis.Device) {
	for {
		logs.Infof("Device %s is getting task state....", device.Name)
		tr, err := rmf.GetTaskState(device, taskId)
		if err != nil {
			logs.Errorf("Device %s monitor GetTaskState failed\n", device.Name)
			return
		}
		switch tr.Status {
		case "failed":
			//TODO:错误处理
			logs.Errorf("Device %s is failed\n", device.Name)
			dr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
			if err = utils.UpdateDeviceStatusFailed(runtime, action, device, dr.deviceClient); err != nil {
				logs.Errorf(err.Error())
			}
			if _, err = dr.deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
				logs.Errorf(err.Error())
			}
			return
		case "completed":
			//TODO:锁操作
			logs.Infof("Device %s is completed\n", device.Name)
			dr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
			if err = utils.UpdateDeviceStatusFailed(runtime, action, device, dr.deviceClient); err != nil {
				logs.Errorf(err.Error())
			}
			if _, err = dr.deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
				logs.Errorf(err.Error())
			}
			return
		case "canceled":
			logs.Infof("Device %s is canceled\n", device.Name)
			dr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
			if err = utils.UpdateDeviceStatusFailed(runtime, action, device, dr.deviceClient); err != nil {
				logs.Errorf(err.Error())
			}
			if _, err = dr.deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
				logs.Errorf(err.Error())
			}
			return
		default:
			logs.Infof("Device %s is %s\n", device.Name, tr.Status)
		}
		time.Sleep(time.Millisecond * 500)
	}

}

func (dr *DeviceRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) string {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	//TODO implement me
	panic("implement me")
}

func (dr *DeviceRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	//TODO implement me
	panic("implement me")
}

func InitClient() (*clients.ClientSet, error) {
	//初始化ClientSet客户端
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
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
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize clientSet: %v", err)
	}
	return clientSet, nil
}
