package device

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
	"time"
)

type AllocateRequest struct {
	RequestTime     apis.Time
	Priority        int
	AbilityNumTable map[string]int
	GroupID         string
	RequestID       string
}

// AbilityMapDevice 维护一个Ability到Map的映射表 存储所有非Disconnected的设备
type AbilityMapDevice struct {
	// Key是能力名 相同能力的设备放在一个列表中
	MapTable map[string][]*apis.Device
	// 用来存放所有的device名字
	NameList []string
}

func (a *AbilityMapDevice) hasDevice(deviceName string) bool {
	for _, name := range a.NameList {
		if name == deviceName {
			return true
		}
	}
	return false
}

type Worker interface {
	Run() error
}

type DeviceWorker struct {
	RequestChan    chan AllocateRequest
	DeviceMap      *AbilityMapDevice
	DeviceClient   core.DeviceInterface
	errorChan      chan error      // 用于传递错误
	AllocateResult map[string]bool // 分配结果表
	requestQueue   []AllocateRequest
	mu             sync.Mutex // 互斥锁，用于保护 requestQueue
}

func NewDeviceWorker(deviceClient core.DeviceInterface) *DeviceWorker {
	deviceMap := &AbilityMapDevice{
		MapTable: make(map[string][]*apis.Device),
		NameList: make([]string, 0),
	}
	return &DeviceWorker{
		RequestChan:    make(chan AllocateRequest, 10), // 创建一个缓冲的请求通道
		DeviceClient:   deviceClient,
		DeviceMap:      deviceMap,
		errorChan:      make(chan error, 1),   // 创建一个缓冲的错误通道
		AllocateResult: make(map[string]bool), // 初始化分配结果表
	}

}
func (w *DeviceWorker) Run() error {
	// 更新设备状态的定时器
	deviceUpdateTicker := time.NewTicker(time.Second * 5)
	// 处理请求的定时器
	requestProcessingTicker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-deviceUpdateTicker.C: // 设备更新
			go w.updateDeviceMap()
		case req := <-w.RequestChan: // 处理请求
			// 接收到请求，将其加入队列
			w.mu.Lock()
			w.requestQueue = append(w.requestQueue, req)
			w.mu.Unlock()
		case err := <-w.errorChan: // 错误反馈
			logs.Infof("Error in DeviceWorker: %v", err)
			return err
		case <-requestProcessingTicker.C:

		}
	}
}

// updateDeviceMap 负责定期更新AbilityMapDevice
func (w *DeviceWorker) updateDeviceMap() {
	devices, err := w.DeviceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("[DEVICE WORKER] ")
		w.errorChan <- err
		return
	}

	for _, device := range devices.Items {
		if device.Status.Phase == apis.DeviceDisconnected {
			continue
		}
		if w.DeviceMap.hasDevice(device.Name) {

		} else { // 如果map中没有存放这个device

		}

	}
}
