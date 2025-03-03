package device

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/rmf"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/utils"
)

type DeviceRuntime struct {
	deviceClient core.DeviceInterface
	groupClient  core.GroupInterface
}

func NewDeviceRuntime(deviceClient core.DeviceInterface, groupClient core.GroupInterface) DeviceRuntime {
	return DeviceRuntime{
		deviceClient: deviceClient,
		groupClient:  groupClient,
	}
}

func (dr DeviceRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error {

	// 检查Device
	logs.Infof("Action[%s] Runtime[%s] CheckDevice start\n", action.Spec.Name, runtime.Name)
	err, devices := utils.CheckDevice(runtime, action)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] CheckDevice failed\n", action.Spec.Name, runtime.Name)
		return err
	}
	logs.Infof("Action[%s] Runtime[%s] CheckDevice is successful\n", action.Spec.Name, runtime.Name)

	//// 检查Resource
	//logs.Infof("Action[%s] Runtime[%s] CheckResource start\n", action.Spec.Name, runtime.Name)
	//err = utils.CheckResource(runtime, action)
	//if err != nil {
	//	logs.Errorf("Action[%s] Runtime[%s] check resource failed\n", action.Spec.Name, runtime.Name)
	//	return err
	//}
	//logs.Infof("Action[%s] Runtime[%s] CheckResource is successful\n", action.Spec.Name, runtime.Name)
	//
	//// 检查Scene
	//logs.Infof("Action[%s] Runtime[%s] CheckScene start\n", action.Spec.Name, runtime.Name)
	//err = utils.CheckScene(runtime, action)
	//if err != nil {
	//	logs.Errorf("Action[%s] Runtime[%s] check scene failed\n", action.Spec.Name, runtime.Name)
	//	return err
	//}
	//logs.Infof("Action[%s] Runtime[%s] CheckScene success\n", action.Spec.Name, runtime.Name)

	// 构造任务的请求参数
	logs.Infof("Action[%s] Runtime[%s] ConstructParam start\n", action.Spec.Name, runtime.Name)
	err = utils.ConstructParam(devices, runtime)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] construct param failed\n", action.Spec.Name, runtime.Name)
		return err
	}
	logs.Infof("Action[%s] Runtime[%s] ConstructParam is successful\n", action.Spec.Name, runtime.Name)

	// 拉起任务
	logs.Infof("Action[%s] Runtime[%s] ConstructDevice start", action.Spec.Name, runtime.Name)
	for name, device := range devices {
		//TODO:区分是rmf还是ability
		logs.Infof("device %s execute %s task\n", name, runtime.Image)
		taskId, err := rmf.PublishAbilityInstruction(device, runtime.Image)
		taskId = "this is a test id"
		device = apis.Device{Spec: device.Spec}
		//TODO: 错误处理
		if err != nil {
			return err
		}
		// TODO: 确定Type
		// 没有错误就将返回信息taskId存储到runtime的Outputs中
		output := apis.Output{
			Value:     taskId,
			Name:      "task_id",
			ValueType: "string",
			Type:      apis.ResultsData,
		}
		runtime.Outputs = append(runtime.Outputs, output)
		action.Spec.Runtimes[0] = *runtime
	}

	//修改Device状态
	logs.Infof("Action[%s] Runtime[%s] update device status start\n", action.Spec.Name, runtime.Name)
	err = utils.UpdateDeviceStatus(runtime, action, action.Status.ActionID, devices, dr.deviceClient)
	if err != nil {
		logs.Errorf(err.Error())
		logs.Errorf("Action[%s] Runtime[%s] update device status failed", action.Spec.Name, runtime.Name)
		return fmt.Errorf("Action[%s] Runtime[%s] update device status failed ", action.Spec.Name, runtime.Name)
	}
	logs.Infof("Action[%s] Runtime[%s] update device status is finished\n", action.Spec.Name, runtime.Name)

	// 更新group
	group.Status.ActionStatus[actionIndex] = action.Status
	_, err = dr.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf(err.Error())
		logs.Errorf("Action[%s] Runtime[%s] update group status failed", action.Spec.Name, runtime.Name)
		return fmt.Errorf("Action[%s] Runtime[%s] update group status failed ", action.Spec.Name, runtime.Name)
	}
	////修改Resource状态
	//logs.Infof("Action[%s] Runtime[%s] update resource status start\n", action.Spec.Name, runtime.Name)
	//err = utils.UpdateResourceStatus(runtime, action)
	//if err != nil {
	//	logs.Errorf("Action[%s] Runtime[%s] update status failed", action.Spec.Name, runtime.Name)
	//	return fmt.Errorf("Action[%s] Runtime[%s] update status failed ", action.Spec.Name, runtime.Name)
	//}
	//logs.Infof("Action[%s] Runtime[%s] update resource status  is finished\n", action.Spec.Name, runtime.Name)
	//
	////修改Scene状态
	//logs.Infof("Action[%s] Runtime[%s] update scene status  start\n", action.Spec.Name, runtime.Name)
	//err = utils.UpdateSceneStatus(runtime, action, len(devices), action.Status.ActionID)
	//if err != nil {
	//	logs.Errorf("Action[%s] Runtime[%s] update scene status failed", action.Spec.Name, runtime.Name)
	//	return fmt.Errorf("Action[%s] Runtime[%s] update scene status failed ", action.Spec.Name, runtime.Name)
	//}
	//logs.Infof("Action[%s] Runtime[%s] update scene status is finished\n", action.Spec.Name, runtime.Name)
	//
	//// etcd更改
	return nil
}

func (dr DeviceRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	// 检查action的运行状态，只有处在running状态时才能取消
	if action.Status.Phase == apis.Running {
		// 一个runtime可能涉及到多个或一个 device 获取全部的device
		err, devices := utils.ObtainDevices(runtime, action)
		if err != nil {
			logs.Errorf("Action[%s] Runtime[%s] CheckDevice failed\n", action.Spec.Name, runtime.Name)
			return err
		}
		// 遍历全部的device
		for index, device := range devices {
			fmt.Println("this is a string")
			// device的taskId存储在output中，同时取出
			output := runtime.Outputs[index]
			// 发布指令
			_, err = rmf.PublishCancelTaskInstruction(device, output.Value)
			if err != nil {
				logs.Errorf("Action[%s] Runtime[%s] PublishCancelTaskInstruction failed\n", action.Spec.Name, runtime.Name)
				return err
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

func (dr DeviceRuntime) CheckTaskStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {
	return "", nil
}
