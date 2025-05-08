package device

import (
	"encoding/json"
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
		instance.updateMap()
	})
	return instance
}

// hasDevice 用来判断是否有某个设备
func (dw *DeviceWorker) hasDevice(deviceName string) bool {
	for name := range dw.MapTable {
		if name == deviceName {
			return true
		}
	}
	return false
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
	dw.mu.Lock()
	for _, device := range deviceList.Items {
		dw.MapTable[device.Name] = &device
	}
	dw.mu.Unlock()
}

// CheckSatisfaction 用来检查
func (dw *DeviceWorker) CheckSatisfaction(abilities []string) (bool, map[string]*apis.Device) {

	deviceTable := make(map[string]*apis.Device, 1)
	dw.updateMap()
	dw.mu.Lock()
	defer dw.mu.Unlock()
	// 遍历传进来的能力表
	for _, ability := range abilities {
		// 定义一个标志位
		flag := false
		for _, device := range dw.MapTable {
			// 当有这个能力 并且 设备没有上锁时
			if device.Status.Phase != apis.DeviceFrozen && contains(device.Spec.Abilities, ability) {
				if device.Status.Lock.IsLocked == false {
					flag = true
					deviceTable[ability] = device
					break
				}
			}
		}
		if !flag {
			return flag, nil
		}
	}
	for _, device := range deviceTable {
		patchDevice, _ := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"phase": apis.DeviceFrozen,
			},
		})
		_, err := dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Infof("[DEVICE WORKER] PATCH FAIL")
		}
	}

	return true, deviceTable
}

// LockDevices 对device进行上锁
func (dw *DeviceWorker) LockDevices(group *apis.Group, deviceTable map[string]*apis.Device) *apis.Group {
	dw.updateMap()
	dw.mu.Lock()
	defer dw.mu.Unlock()
	for ai, ac := range group.Spec.Actions {
		for ri, runtime := range ac.Runtimes {
			for di, ds := range runtime.Devices { // 遍历到runtime的device这一层
				for _, ab := range ds.Abilities { // 寻找runtime中的每一个ability
					device := deviceTable[ab]        // 根据ability找到device
					a := device.Status.Abilities[ab] // 找到对应能力

					// 修改能力
					if !device.Status.Abilities[ab].Lock.IsLocked { // 对应的能力加锁
						a.Lock.IsLocked = true
					}
					a.Status = apis.AbilityReadyStartUp // FIXME 这里暂定为准备拉起状态
					a.Lock.Ref += 1

					// 修改device
					if !device.Status.Lock.IsLocked {
						device.Status.Lock.IsLocked = true
					}
					device.Status.Lock.Ref += 1
					device.Status.Abilities[ab] = a

					deviceTable[ab] = device

					ds.ExpectedProperties["name"] = apis.Property{
						Value: deviceTable[ab].Name,
					}
				}
				runtime.Devices[di] = ds // 把填完的DeviceSpec放入Device中
			}
			ac.Runtimes[ri] = runtime
		}
		group.Spec.Actions[ai] = ac
	}

	for _, device := range deviceTable {
		// 更新能力状态
		patchDevice, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"abilities": device.Status.Abilities,
				"lock":      device.Status.Lock,
			},
		})
		_, err = dw.Manager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE WORKER] Patch device:%s fail", device.Name)
			return nil
		}
		dw.MapTable[device.Name] = device
	}
	return group
}

func contains(slice []string, target string) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}
