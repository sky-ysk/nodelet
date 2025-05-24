package device

import (
	"context"
	"encoding/json"
	"errors"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	m "hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/utils"
	"sync"
	"time"
)

var instance *DeviceWorker
var once sync.Once

type DeviceWorker struct {
	MapTable map[string]*apis.Device
	// etcd的client
	Manager *m.Manager
	errChan chan error
	// 互斥锁
	Mu sync.Mutex
}

// GetDeviceWorker 获取DeviceWorker的单例实例
func GetDeviceWorker() *DeviceWorker {
	once.Do(func() {

		//clientSet, _ := InitClient()
		clientSet, err := utils.CreateClientSetWithTimeOut(3600 * 3)
		if err != nil {
			logs.Fatalf("create client set failed, err:%v", err)
		}
		instance = &DeviceWorker{
			MapTable: make(map[string]*apis.Device),
			Manager:  m.NewManager(clientSet),
			errChan:  make(chan error, 1),
		}
		ctx := context.Background()
		go instance.monitorDiscardRuntimes(ctx)
		//go instance.errorCycle()
	})
	return instance
}

// updateMapCycle 负责更新一下内存中的map
func (dw *DeviceWorker) updateMap() {
	logs.Infof("[DEVICE WORKER] update all devices")
	deviceList, err := dw.Manager.GetDevices("", "test")
	if err != nil {
		logs.Infof("[DEVICE WORKER] List All Devices failed")
		return
	}
	logs.Infof("[DEVICE WORKER] HAS %d devices", len(deviceList.Items))

	for _, device := range deviceList.Items {
		dw.MapTable[device.Name] = &device
		logs.Warnf("[DEVICE WORKER] Device %s has been updated, ref is %d", device.Name, device.Status.Lock.Ref)
		for name, ability := range device.Status.Abilities {
			logs.Warnf("[DEVICE WORKER] Device %s ability %s ref is %d", device.Name, name, ability.Lock.Ref)
		}
	}

}

// ChooseDevices 用来筛选Group需要的设备
func (dw *DeviceWorker) ChooseDevices(group *apis.GroupSpec) (bool, map[string]*apis.Device, error) {

	deviceTable := make(map[string]*apis.Device, len(group.Devices))
	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()
	// 遍历device需求表
	for _, ds := range group.Devices {
		//区分filter策略和Nominate策略, 默认是filter策略
		strategy := ""
		strategyVal, exists := ds.ExpectedProperties["strategy"]
		if !exists || strategyVal.Value == "" {
			strategy = "filter"
		} else {
			strategy = strategyVal.Value
		}
		switch strategy {
		case "filter":
			flag := false // 定义一个标志位，标识这个device需求是否满足
			for _, device := range dw.MapTable {
				// 找到满足能力需求的device
				if device.Status.Lock.IsLocked != true && device.Status.Phase == apis.DeviceIdle { // 没有上锁的设备
					if _, ok := deviceTable[device.Name]; ok {
						continue
					}
					if contains(device.Spec.Abilities, ds.Abilities) {
						// 将device存入deviceTable中
						deviceTable[ds.Name] = device
						//delete(dw.MapTable, device.Name)
						flag = true // 满足设置为true
						break
					}
				}
			}
			if !flag { // device需求未满足，返回false
				logs.Warnf("no device statisfied for group %s", group.Name)
				return false, nil, nil
			}
		//最好不要有一些是nominate有一些是filter ，单个Group中的DeviceSpec中的strategy最好保持一致，要不然可能会出错
		case "nominate":
			// 获取真实的名字
			deviceName, ok := ds.ExpectedProperties["name"]
			if !ok || deviceName.Value == "" {
				logs.Errorf("strategy is nominate but no device name ! group %s, device spec %s", group.Name, ds.Name)
				return false, nil, errors.New("strategy is nominate but no device name")
			}
			// 查看table中的数据，能不能根据真实的名字找到这个device
			device, ok := dw.MapTable[deviceName.Value]
			if !ok || device == nil {
				logs.Errorf("nominated device %s not exist", ds.Name)
				return false, nil, errors.New("nominated device not exist")
			}
			if device.Status.Lock.IsLocked {
				logs.Warnf("nominated device %s is locked", ds.Name)
				return false, nil, nil
			}
			if device.Status.Phase != apis.DeviceIdle {
				logs.Warnf("nominated device %s is error", ds.Name)
				return false, nil, nil
			}
			deviceTable[ds.Name] = device

		default:
			logs.Errorf("invalid strtegy %s", strategy)
			return false, nil, errors.New("invalid strategy")
		}

	}
	return true, deviceTable, nil
}

// LockDevices 对device进行上锁
func (dw *DeviceWorker) LockDevices(group *apis.Group, deviceTable map[string]*apis.Device) (bool, *apis.Group) {

	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()
	// 首先检查所需的设备有没有被占用
	for _, device := range deviceTable {
		if dw.MapTable[device.Name].Status.Lock.IsLocked != false {
			logs.Warnf("device %s is locked, group %s", device.Name, group.Name)
			return false, nil
		}
		if dw.MapTable[device.Name].Status.Phase != apis.DeviceIdle {
			logs.Warnf("device %s is not idle, group %s", device.Name, group.Name)
			return false, nil
		}
	}

	// 对所有设备上锁
	for name, device := range deviceTable {
		device.Status.Lock.IsLocked = true
		deviceTable[name] = device
	}
	// 填写group的devices字段

	for index, deviceSpec := range group.Spec.Devices {
		if deviceSpec.ExpectedProperties == nil {
			deviceSpec.ExpectedProperties = make(map[string]apis.Property)
		}
		deviceSpec.ExpectedProperties["name"] = apis.Property{
			Value: deviceTable[deviceSpec.Name].Name,
		}
		group.Spec.Devices[index] = deviceSpec
	}

	// 遍历Group的actions
	for ai, a := range group.Spec.Actions {
		// 获取action的真实信息
		aRealName := group.Status.Actions[a.Name].Name
		aNamespace := group.Status.Actions[a.Name].Namespace
		action, err := dw.Manager.GetAction(aRealName, aNamespace)
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] GET action fail")
			return false, nil
		}
		// 遍历runtime
		for ri, r := range a.Runtimes {
			// 获取runtime的真实信息
			rRealName := action.Status.Runtimes[r.Name].Name
			rNamespace := action.Status.Runtimes[r.Name].Namespace
			for di, ds := range r.Devices {
				// 对runtime的device字段进行操作
				ds.ExpectedProperties["name"] = apis.Property{
					Value: deviceTable[ds.Name].Name,
				}
				// 对设备加锁
				for _, ab := range ds.Abilities {

					deviceTable[ds.Name].Status.Lock.Ref += 1
					ability := deviceTable[ds.Name].Status.Abilities[ab]
					ability.Lock.Ref += 1
					deviceTable[ds.Name].Status.Abilities[ab] = ability
					logs.Infof("Group:%s Device:%s Ref:%d", group.Name, deviceTable[ds.Name].Name, deviceTable[ds.Name].Status.Lock.Ref)
				}

				r.Devices[di] = ds
			}

			// 更新runtime
			patchRuntime, _ := json.Marshal(map[string]interface{}{
				"spec": map[string]interface{}{
					"devices": r.Devices,
				},
			})
			_, err = dw.Manager.PatchRuntime(rRealName, rNamespace, patchRuntime)
			if err != nil {
				logs.Errorf("[DEVICE WORKER] Patch runtime %s failed", rRealName)
				return false, nil
			}
			logs.Infof("[DEVICE WORKER] Patch runtime:%s successfully", rRealName)
			a.Runtimes[ri] = r
		}
		// 更新runtime
		patchAction, _ := json.Marshal(map[string]interface{}{
			"spec": map[string]interface{}{
				"runtimes": a.Runtimes,
			},
		})
		_, err = dw.Manager.PatchAction(aRealName, aNamespace, patchAction)
		if err != nil {
			logs.Errorf("[DEVICE WORKER] Patch action %s failed", aRealName)
			return false, nil
		}
		logs.Infof("[DEVICE WORKER] Patch action:%s successfully", aRealName)
		group.Spec.Actions[ai] = a
	}

	// 更新设备
	for _, device := range deviceTable {
		patchDevice, _ := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"lock":      device.Status.Lock,
				"abilities": device.Status.Abilities,
				"group":     group.Name,
			},
		})
		_, err := dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE WORKER] Patch device %s failed", device.Name)
			return false, nil
		}
		logs.Infof("[DEVICE WORKER] LOCK device:%s successfully", device.Name)
	}

	// 更新group

	//groupSpecStr, err := json.Marshal(group.Spec)
	//if err != nil {
	//	logs.Errorf("[DEVICE WORKER] Marshal Group Spec failed: %s", err.Error())
	//	return false, nil
	//}
	//patchGroup, _ := json.Marshal(map[string]interface{}{
	//	"spec": string(groupSpecStr),
	//})
	logs.Infof("group is %s", group.Spec.Actions[0].Runtimes[0].Devices[0].ExpectedProperties["name"].Value)
	logs.Infof("action is %v", group.Spec.Actions[0])
	_, err := dw.Manager.UpdateGroup(group.Name, apis.NamespaceTest, group)
	if err != nil {
		logs.Errorf("[DEVICE WORKER] update group %s failed", group.Name)
		return false, nil
	}
	logs.Infof("[DEVICE RUNTIME] update group successfully")
	return true, group
}

// LockAbility 对能力进行上锁
func (dw *DeviceWorker) LockAbility(device *apis.Device, ability string) bool {
	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()
	d := dw.MapTable[device.Name]
	if d.Status.Abilities[ability].Lock.IsLocked != false {
		return false
	} else {
		a := d.Status.Abilities[ability]
		a.Lock.IsLocked = true
		d.Status.Abilities[ability] = a
	}
	patchDevice, _ := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"abilities": d.Status.Abilities,
		},
	})
	_, err := dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
	if err != nil {
		logs.Errorf("[DEVICE WORKER] Patch device %s failed", device.Name)
		return false
	}
	return true
}

// ReleaseAbilityRef 在非正常情况下减少引用
func (dw *DeviceWorker) ReleaseAbilityRef(device string, ability string, runtime *apis.Runtime) error {

	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()

	// 获取device
	d := dw.MapTable[device]
	// 获取ability
	a := d.Status.Abilities[ability]
	a.Lock.Ref -= 1 // 减引用
	logs.Warnf("[DEVICE worker] ABNORMAL")
	logs.Warnf("[DEVICE WORKER] REF IS %d, runtime is %s(before)", d.Status.Lock.Ref, runtime.Name)
	d.Status.Lock.Ref -= 1 // 减引用

	d.Status.Abilities[ability] = a

	aByte, err := json.Marshal(d.Status.Abilities)
	lByte, err := json.Marshal(d.Status.Lock)
	patchDevice, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"abilities": json.RawMessage(aByte),
			"lock":      json.RawMessage(lByte),
		},
	})

	if err != nil {
		logs.Errorf("json marshal fail err：%s", err.Error())
		return err
	}
	_, err = dw.Manager.PatchDevice(device, d.Namespace, string(patchDevice))
	if err != nil {
		logs.Errorf("[DEVICE WORKER] Patch device %s failed", device)
		return err
	}
	d1, err := dw.Manager.GetDevice(device, d.Namespace)
	if err != nil {
		logs.Errorf("[DEVICE WORKER] Get device %s failed", device)
	}
	logs.Warnf("[DEVICE WORKER] REF IS %d, runtime is %s(after)", d1.Status.Lock.Ref, runtime.Name)
	return nil
}

// ReleaseAbility 正常释放能力锁
func (dw *DeviceWorker) ReleaseAbility(device *apis.Device, ability string) bool {
	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()
	logs.Warnf("[DEVICE RUNTIME] NORMAL")
	d := dw.MapTable[device.Name]
	if d.Status.Abilities[ability].Lock.IsLocked != true {
		return false
	}
	a := d.Status.Abilities[ability]
	logs.Warnf("[DEVICE WORKER] REF IS  ABILITY %s ref is %d(BEFORE)", ability, a.Lock.Ref)
	a.Lock.IsLocked = false // 解锁
	a.Lock.Ref -= 1         // 减引用
	d.Status.Abilities[ability] = a

	aByte, err := json.Marshal(d.Status.Abilities)
	patchDevice, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"abilities": json.RawMessage(aByte),
		},
	})
	if err != nil {
		logs.Errorf("json marshal fail err：%s", err.Error())
	}
	_, err = dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
	if err != nil {
		logs.Errorf("[DEVICE WORKER] Patch device %s failed, err:%s", device.Name, err.Error())
		return false
	}
	dd, err := dw.Manager.GetDevice(device.Name, device.Namespace)
	if err != nil {
		logs.Errorf("[DEVICE WORKER] Get device %s failed", device.Name)
		return false
	}
	logs.Warnf("[DEVICE WORKER] REF IS  ABILITY %s ref is %d(after)", ability, dd.Status.Abilities[ability].Lock.Ref)
	logs.Infof("[DEVICE RUNTIME] RELEASE device:%s ability:%s successfully", device.Name, ability)
	return true

}

// monitorDiscardRuntimes 监听丢弃的runtime
func (dw *DeviceWorker) monitorDiscardRuntimes(ctx context.Context) {
	//eventClient := dw.Manager.EventClients[apis.NamespaceTest]
	eventClient := dw.Manager.ClientSet.Core().Events(apis.NamespaceTest)
	if eventClient == nil {
		logs.Error("No such event client!")
		return
	}
	var timeOut int64 = 3600 * 3
	watch, err := eventClient.Watch(ctx, metav1.ListOptions{
		TimeoutSeconds: &timeOut,
	})
	watchChan := watch.ResultChan()
	if err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			logs.Errorf("[device worker] ctx is done")
		case e, ok := <-watchChan:
			if !ok {
				logs.Error("[device worker] watch channel closed")
				return
			}

			if event, ok := e.Object.(*apis.Event); ok {
				//logs.Infof("[DEVICE RUNTIME] EVENT IS %v", event)
				//logs.Infof("[DEVICE RUNTIME] Receive event:%s ", event.Name)
				if event.InvolvedObject.Kind != "Runtime" || event.Reason != "ExecuteDiscard" {
					continue
				}
				dw.handleRuntimeDiscardEvent(ctx, event)
			} else {
				logs.Errorf("[device worker] cannot tranform to event")
			}
		}
	}

}

// handleRuntimeDiscardEvent 处理丢弃事件
func (dw *DeviceWorker) handleRuntimeDiscardEvent(ctx context.Context, event *apis.Event) {
	name := event.InvolvedObject.Name
	namespace := event.InvolvedObject.Namespace
	//logs.Infof("[DEVICE RUNTIME] handleRuntimeDiscardEvent runtime IS %v", name)
	// 获取runtime
	runtime, err := dw.Manager.GetRuntime(name, namespace)
	if err != nil {
		logs.Errorf("[DEVICE WORKER] get runtime %s failed", name)
		return
	}
	// 释放不正常的引用
	for _, ds := range runtime.Spec.Devices {
		dname := ds.ExpectedProperties["name"].Value
		for _, ability := range ds.Abilities {
			logs.Errorf("runtime%s 被丢弃，不正常释放device%s的ability%s", runtime.Name, dname, ability)
			err = dw.ReleaseAbilityRef(dname, ability, runtime)
			if err != nil {
				logs.Errorf("[DEVICE WORKER] ReleaseAbilityRef %s failed, err is %s", dname, err.Error())
				return
			}
		}
	}
}

// UpdateDeviceFinished 任务完成阶段进行设备状态更新
func (dw *DeviceWorker) UpdateDeviceFinished(deviceMap map[string]*apis.Device, runtime *apis.Runtime) error {
	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()
	logs.Warnf("[DEVICE WORKER] NORMAL")
	for _, d := range deviceMap {
		device := dw.MapTable[d.Name]
		logs.Infof("[DEVICE RUNTIME] Update Device[%s] stage[FINISHED]", device.Name)
		logs.Warnf("[DEVICE RUNTIME] REF IS %d, runtime is %s(before)", device.Status.Lock.Ref, runtime.Name)
		device.Status.Lock.Ref -= 1

		if device.Status.Lock.Ref == 0 {
			device.Status.Lock.IsLocked = false
		}
		// 将phase更改为running
		device.Status.Phase = apis.DeviceIdle
		// 设置更新时间
		device.Status.LastTime = apis.Time{Time: time.Now()}
		patchDevice, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"lock":      device.Status.Lock,
				"phase":     device.Status.Phase,
				"last_time": device.Status.LastTime,
			},
		})

		_, err = dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE Worker] Update Device[%s] stage[FINISHED], err:%s", device.Name, err)
			return err
		}
		d1, err := dw.Manager.GetDevice(device.Name, device.Namespace)
		if err != nil {
			logs.Errorf("[DEVICE WORKER] Get device %s failed", device.Name)
		}
		logs.Warnf("[DEVICE RUNTIME] REF IS %d, runtime is %s(after)", d1.Status.Lock.Ref, runtime.Name)
		logs.Infof("[DEVICE Worker] Update Device[%s] successfully stage [FINISHED]\n", device.Name)
	}

	return nil
}

func (dw *DeviceWorker) UpdateDeviceRunning(deviceMap map[string]*apis.Device) error {
	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()
	for name, d := range deviceMap {
		device := dw.MapTable[d.Name]
		logs.Infof("[DEVICE RUNTIME] Update Device[%s] stage[RUNNING]", name)
		// 将phase更改为running
		device.Status.Phase = apis.DeviceRunning
		// 设置更新时间
		device.Status.LastTime = apis.Time{Time: time.Now()}
		patchDevice, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"phase":     device.Status.Phase,
				"last_time": device.Status.LastTime,
			},
		})
		_, err = dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE WORKER] Update Device[%s] Failed stage [RUNNING], err:%s", device.Name, err.Error())
			return err
		}
		logs.Infof("[DEVICE WORKER] Update Device[%s] Successfully stage [RUNNING]\n", device.Name)
	}

	return nil
}

// UpdateDeviceError 任务执行失败阶段 更新设备状态
func (dw *DeviceWorker) UpdateDeviceError(message string, d *apis.Device) error {
	dw.Mu.Lock()
	defer dw.Mu.Unlock()
	dw.updateMap()

	logs.Infof("[DEVICE WORKER] Device:%s is Error, message: %s", d.Name, message)
	device := dw.MapTable[d.Name]
	errEvent := apis.DeviceEvent{
		Desc: message,
	}

	// 填写错误信息
	device.Status.Events = append(device.Status.Events, errEvent)

	// 改写状态为error
	device.Status.Phase = apis.DeviceError
	device.Status.LastTime = apis.Time{Time: time.Now()}

	// 清空每一个能力的锁信息
	for name, ability := range device.Status.Abilities {
		ability.Lock.IsLocked = false
		ability.Lock.Ref = 0
		device.Status.Abilities[name] = ability
	}

	// 清空设备的锁信息
	device.Status.Lock.IsLocked = false
	device.Status.Lock.Ref = 0
	device.Status.Group = ""
	aByte, err := json.Marshal(device.Status.Abilities)
	if err != nil {
		logs.Errorf("[DEVICE WORKER] Marshal device ability fail, err:%s", err.Error())
		return err
	}

	var patchDevice []byte
	patchDevice, err = json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"group":     device.Status.Group,
			"events":    device.Status.Events,
			"abilities": json.RawMessage(aByte),
			"lock":      device.Status.Lock,
			"phase":     device.Status.Phase,
			"last_time": device.Status.LastTime,
		},
	})
	_, err = dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
	if err != nil {
		logs.Errorf("[DEVICE Worker] Update Device[%s] stage[ERROR], err:%s", device.Name, err)
		return err
	}
	logs.Infof("[DEVICE WORKER] Patch Device%s Successfully\n", device.Name)
	return nil
}

// 定时错误检测
func (dw *DeviceWorker) errorCycle() {
	for {
		time.Sleep(5 * time.Second)
		dw.checkError()
	}
}

// 进行一次错误检查
func (dw *DeviceWorker) checkError() {
	dw.Mu.Lock()
	// 获取这段时间中处于ERROR状态的device
	dw.updateMap()
	errDeviceList := make([]*apis.Device, 0)
	for _, d := range dw.MapTable {
		if d.Status.Phase == apis.DeviceError {
			errDeviceList = append(errDeviceList, d)
		}
	}
	dw.Mu.Unlock()
	//
	//// 进行错误处理
	//for _, device := range errDeviceList {
	//	errEvent :=
	//}

}

// handleSimpleError 处理简单错误
func (dw *DeviceWorker) handleSimpleError() {

}

func contains(slice []string, target []string) bool {
	for _, t := range target {
		found := false
		for _, s := range slice {
			if s == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
