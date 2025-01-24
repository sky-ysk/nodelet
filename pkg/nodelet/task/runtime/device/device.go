package device

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/rmf"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/utils"
)

type DeviceRuntime struct {
}

func NewDeviceRuntime() DeviceRuntime {
	return DeviceRuntime{}
}

func (dr DeviceRuntime) Run(group *apis.Group, action *apis.Action) error {

	//从action中获取runtime TODO:暂时设置为第1个runtime
	runtime := &action.Spec.Runtimes[0]

	// 检查Device
	logs.Infof("Action[%s] Runtime[%s] CheckDevice start\n", action.Spec.Name, runtime.Name)
	err, devices := utils.CheckDevice(runtime, action)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] CheckDevice failed\n", action.Spec.Name, runtime.Name)
		return err
	}
	logs.Infof("Action[%s] Runtime[%s] CheckDevice is successful\n", action.Spec.Name, runtime.Name)

	// 检查Resource
	logs.Infof("Action[%s] Runtime[%s] CheckResource start\n", action.Spec.Name, runtime.Name)
	err = utils.CheckResource(runtime, action)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] check resource failed\n", action.Spec.Name, runtime.Name)
		return err
	}
	logs.Infof("Action[%s] Runtime[%s] CheckResource is successful\n", action.Spec.Name, runtime.Name)

	// 检查Scene
	logs.Infof("Action[%s] Runtime[%s] CheckScene start\n", action.Spec.Name, runtime.Name)
	err = utils.CheckScene(runtime, action)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] check scene failed\n", action.Spec.Name, runtime.Name)
		return err
	}
	logs.Infof("Action[%s] Runtime[%s] CheckScene success\n", action.Spec.Name, runtime.Name)

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
	}

	//修改Device状态
	logs.Infof("Action[%s] Runtime[%s] update device status start\n", action.Spec.Name, runtime.Name)
	err = utils.UpdateDeviceStatus(action, action.Status.ActionID)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] update device status failed", action.Spec.Name, runtime.Name)
		return fmt.Errorf("Action[%s] Runtime[%s] update device status failed ", action.Spec.Name, runtime.Name)
	}
	logs.Infof("Action[%s] Runtime[%s] update device status is finished\n", action.Spec.Name, runtime.Name)

	//修改Resource状态
	logs.Infof("Action[%s] Runtime[%s] update resource status start\n", action.Spec.Name, runtime.Name)
	err = utils.UpdateResourceStatus(runtime, action)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] update status failed", action.Spec.Name, runtime.Name)
		return fmt.Errorf("Action[%s] Runtime[%s] update status failed ", action.Spec.Name, runtime.Name)
	}
	logs.Infof("Action[%s] Runtime[%s] update resource status  is finished\n", action.Spec.Name, runtime.Name)

	//修改Scene状态
	logs.Infof("Action[%s] Runtime[%s] update scene status  start\n", action.Spec.Name, runtime.Name)
	err = utils.UpdateSceneStatus(runtime, action, len(devices), action.Status.ActionID)
	if err != nil {
		logs.Errorf("Action[%s] Runtime[%s] update scene status failed", action.Spec.Name, runtime.Name)
		return fmt.Errorf("Action[%s] Runtime[%s] update scene status failed ", action.Spec.Name, runtime.Name)
	}
	logs.Infof("Action[%s] Runtime[%s] update scene status is finished\n", action.Spec.Name, runtime.Name)

	return nil
}

func (dr DeviceRuntime) Kill(group *apis.Group, action *apis.Action) error {
	// TODO: 检查Action的运行状态，只有在Running状态的任务，才能停止部署

	logs.Infof("device runtime kill task: %s", group.Name)
	return nil
}
