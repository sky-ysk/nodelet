package device

import (
	"encoding/json"
	"errors"
	apis "hit.edu/framework/pkg/apis/cores"
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
			deviceName, ok := ds.ExpectedProperties["name"]
			if !ok || deviceName.Value == "" {
				logs.Errorf("strategy is nominate but no device name ! group %s, device spec %s", group.Name, ds.Name)
				return false, nil, errors.New("strategy is nominate but no device name")
			}
			device, ok := dw.MapTable[ds.Name]
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
		deviceSpec.ExpectedProperties["name"] = apis.Property{
			Value: deviceTable[deviceSpec.Name].Name,
		}
		group.Spec.Devices[index] = deviceSpec
	}

	// 遍历Group的actions
	for ai, a := range group.Spec.Actions {
		for ri, r := range a.Runtimes {
			for di, ds := range r.Devices {
				ds.ExpectedProperties["name"] = apis.Property{
					Value: deviceTable[ds.Name].Name,
				}
				for _, ab := range ds.Abilities {
					deviceTable[ds.Name].Status.Lock.Ref += 1
					ability := deviceTable[ds.Name].Status.Abilities[ab]
					ability.Lock.Ref += 1
					deviceTable[ds.Name].Status.Abilities[ab] = ability
				}
				r.Devices[di] = ds
			}

			a.Runtimes[ri] = r
		}
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
	_, err := dw.Manager.UpdateGroup(group.Name, apis.NamespaceTest, group)
	if err != nil {
		logs.Errorf("[DEVICE WORKER] Patch group %s failed", group.Name)
		return false, nil
	}
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

func (dw *DeviceWorker) ReleaseAbility(device *apis.Device, ability string) bool {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	dw.updateMap()
	d := dw.MapTable[device.Name]
	if d.Status.Abilities[ability].Lock.IsLocked != true {
		return false
	} else {
		a := d.Status.Abilities[ability]
		a.Lock.IsLocked = false // 解锁
		a.Lock.Ref -= 1         // 减引用
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
