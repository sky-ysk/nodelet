package device

import (
	"context"
	"encoding/json"
	"errors"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	m "hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
)

var instance *DeviceWorker
var once sync.Once

type DeviceWorker struct {
	MapTable map[string]*apis.Device
	// etcd的client
	Manager *m.Manager
	errChan chan error
	// 互斥锁
	mu sync.Mutex
}

// GetDeviceWorker 获取DeviceWorker的单例实例
func GetDeviceWorker() *DeviceWorker {
	once.Do(func() {
		clientSet, _ := InitClient()
		instance = &DeviceWorker{
			MapTable: make(map[string]*apis.Device),
			Manager:  m.NewManager(clientSet),
			errChan:  make(chan error, 1),
		}
		ctx := context.Background()
		go instance.monitorDiscardRuntimes(ctx)
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
	}

}

// ChooseDevices 用来筛选Group需要的设备
func (dw *DeviceWorker) ChooseDevices(group *apis.GroupSpec) (bool, map[string]*apis.Device, error) {

	deviceTable := make(map[string]*apis.Device, len(group.Devices))
	dw.mu.Lock()
	defer dw.mu.Unlock()
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
				if device.Status.Lock.IsLocked != true { // 没有上锁的设备
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

	dw.mu.Lock()
	defer dw.mu.Unlock()
	dw.updateMap()
	// 首先检查所需的设备有没有被占用
	for _, device := range deviceTable {
		if dw.MapTable[device.Name].Status.Lock.IsLocked != false {
			logs.Warnf("device %s is locked, group %s", device.Name, group.Name)
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

func (dw *DeviceWorker) LockAbility(device *apis.Device, ability string) bool {
	dw.mu.Lock()
	defer dw.mu.Unlock()
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
func (dw *DeviceWorker) ReleaseAbilityRef(device string, ability string) error {

	dw.mu.Lock()
	defer dw.mu.Unlock()
	dw.updateMap()

	// 获取device
	d := dw.MapTable[device]
	// 获取ability
	a := d.Status.Abilities[ability]
	a.Lock.Ref -= 1        // 减引用
	d.Status.Lock.Ref -= 1 // 减引用
	d.Status.Abilities[ability] = a

	aByte, err := json.Marshal(d.Status.Abilities)
	patchDevice, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"abilities": json.RawMessage(aByte),
			"lock":      d.Status.Lock,
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
	return nil
}

// ReleaseAbility 正常释放能力锁
func (dw *DeviceWorker) ReleaseAbility(device *apis.Device, ability string) bool {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	dw.updateMap()
	d := dw.MapTable[device.Name]
	if d.Status.Abilities[ability].Lock.IsLocked != true {
		return false
	}
	a := d.Status.Abilities[ability]
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
	logs.Infof("[DEVICE RUNTIME] RELEASE device:%s ability:%s successfully", device.Name, ability)
	return true

}

func (dw *DeviceWorker) monitorDiscardRuntimes(ctx context.Context) {
	//eventClient := dw.Manager.EventClients[apis.NamespaceTest]
	eventClient := dw.Manager.ClientSet.Core().Events(apis.NamespaceTest)
	if eventClient == nil {
		logs.Error("No such event client!")
		return
	}
	watch, err := eventClient.Watch(ctx, metav1.ListOptions{})
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
				logs.Infof("[DEVICE RUNTIME] EVENT IS %v", event)
				logs.Infof("[DEVICE RUNTIME] Receive event:%s ", event.Name)
				dw.handleRuntimeDiscardEvent(ctx, event)
			} else {
				logs.Errorf("[device worker] cannot tranform to event")
			}
		}
	}

}

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
			err = dw.ReleaseAbilityRef(dname, ability)
			if err != nil {
				logs.Errorf("[DEVICE WORKER] ReleaseAbilityRef %s failed, err is %s", dname, err.Error())
				return
			}
		}
	}
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
