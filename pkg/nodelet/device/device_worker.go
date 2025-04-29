package device

import (
	apis "hit.edu/framework/pkg/apis/cores"
	m "hit.edu/framework/pkg/client-go/util/manager"
	"sync"
)

var instance *DeviceWorker
var once sync.Once

type DeviceWorker struct {
	MapTable map[string]*apis.Device
	// etcd的client
	Manager *m.Manager
	// 互斥锁
	mu sync.Mutex
}

// GetInstance 获取DeviceWorker的单例实例
func GetInstance() *DeviceWorker {
	once.Do(func() {
		clientSet, _ := InitClient()
		instance = &DeviceWorker{
			MapTable: make(map[string]*apis.Device),
			Manager:  m.NewManager(clientSet),
		}
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

func (dw *DeviceWorker) Run() {
	deviceList, err := dw.Manager.GetDevices("", "test")
	if err != nil {

	}
}
