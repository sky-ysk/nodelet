package device

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"sort"
	"sync"
	"time"
)

// AllocateRequest 描述一个Group对DeviceWorker的请求
type AllocateRequest struct {
	RequestTime     time.Time
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

// hasDevice 判断映射表中是否含有某个device
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
	// 接受请求的通道
	RequestChan chan AllocateRequest
	// Ability到device的映射表
	DeviceMap *AbilityMapDevice
	// 访问Device的client
	DeviceClient core.DeviceInterface
	// 错误处理的通道
	errorChan chan error // 用于传递错误
	// 展示分配结果
	AllocateResult map[string]bool // 分配结果表
	// 存储请求的队列
	requestQueue []AllocateRequest
	// 互斥锁
	mu sync.Mutex
	// 设置分配策略
	Strategy Strategy
}
type Strategy string

const (
	StrategyPriority = "priority"
	StrategyFCFS     = "FCFS"
)

func NewDeviceWorker(deviceClient core.DeviceInterface, strategy ...Strategy) *DeviceWorker {
	deviceMap := &AbilityMapDevice{
		MapTable: make(map[string][]*apis.Device),
		NameList: make([]string, 0),
	}
	if len(strategy) == 0 {
		strategy[0] = StrategyPriority
	}
	return &DeviceWorker{
		RequestChan:    make(chan AllocateRequest, 10), // 创建一个缓冲的请求通道
		DeviceClient:   deviceClient,
		DeviceMap:      deviceMap,
		errorChan:      make(chan error, 1),   // 创建一个缓冲的错误通道
		AllocateResult: make(map[string]bool), // 初始化分配结果表
		Strategy:       strategy[0],
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
	w.mu.Lock()
	defer w.mu.Unlock()
	// 从数据库中拿去所有的device
	devices, err := w.DeviceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("[DEVICE WORKER] ")
		w.errorChan <- err
		return
	}
	logs.Infof("[DEVICE WORKER] ETCD has %d device", len(devices.Items))

	for _, device := range devices.Items {

		// 只负责维护Ability设备
		if device.Spec.AccessMethod.Type == apis.AccessByAbility {
			// 跳过离线的设备
			if device.Status.Phase == apis.DeviceDisconnected {
				logs.Infof("[DEVICE WORKER] Device[%s] is Disconnected", device.Name)
				continue
			}
			logs.Infof("[DEVICE WORKER] Update Device[%s]", device.Name)
			if w.DeviceMap.hasDevice(device.Name) { // 如果map中有这个device
				logs.Infof("[DEVICE WORKER] Device[%s] is exist in DeviceMap", device.Name)
				// 将device进行更新
				abilityName := device.Status.Abilities[0].Name

				for index, d := range w.DeviceMap.MapTable[abilityName] {
					if d.Name == device.Name {
						w.DeviceMap.MapTable[abilityName][index] = &device
					}
				}

				logs.Infof("[DEVICE WORKER] Update Device[%s] in DeviceMap", device.Name)
			} else { // 如果map中没有存放这个device
				logs.Infof("[DEVICE WORKER] Device[%s] is not exist in DeviceMap", device.Name)
				// 加入到map中

				deviceList := w.DeviceMap.MapTable[device.Status.Abilities[0].Name]
				deviceList = append(deviceList, &device)
				w.DeviceMap.MapTable[device.Status.Abilities[0].Name] = deviceList
				// 在nameList中注册
				nameList := w.DeviceMap.NameList
				nameList = append(nameList, device.Name)
				w.DeviceMap.NameList = nameList

				logs.Infof("[DEVICE WORKER] Register Device[%s] in DeviceMap", device.Name)
			}
		}

	}
}

// processRequests 进行请求处理
func (w *DeviceWorker) processRequests() {
	// 取出所有的请求
	var requests []AllocateRequest

	w.mu.Lock()
	for _, request := range w.requestQueue {
		requests = append(requests, request)
	}
	w.requestQueue = w.requestQueue[:0]
	w.mu.Unlock()
	// 根据策略进行处理
	switch w.Strategy {
	case StrategyFCFS:
		sort.Slice(requests, func(i, j int) bool {
			return requests[i].RequestTime.Before(requests[j].RequestTime)
		})
	case StrategyPriority:
		sort.Slice(requests, func(i, j int) bool {
			return requests[i].Priority > requests[j].Priority
		})
	default:

	}
	// 遍历request列表
	for _, request := range requests {
		// 检查每一个request是否可行
		var flag bool
		for ability, _ := range request.AbilityNumTable {
			if flag = !w.isAvailable(ability); flag { // 如果不可行
				w.mu.Lock()
				w.AllocateResult[request.RequestID] = false
				w.mu.Unlock()
				break
			}
		}
		// 正式分配
		if !flag {
			w.mu.Lock()
			for ability, num := range request.AbilityNumTable {
				// 从map中获取对应的device
				device := w.getDevice(ability)
				device.Status.GroupID = request.GroupID
				device.Status.Lock = apis.Lock{
					Ref:  num,
					Lock: true,
				}
				// 先更新etcd中的内容
				_, err := w.DeviceClient.Get(context.TODO(), device.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("[DEVICE WORKER] get device fail")
				}
				_, err = w.DeviceClient.Update(context.TODO(), device, metav1.UpdateOptions{})
				if err != nil {
					logs.Errorf("[DEVICE WORKER] update device fail")
					w.mu.Unlock()
					return
				}
				// 更新map中的内容
				w.updateDevice(ability, device)
			}
			w.AllocateResult[request.RequestID] = true

			w.mu.Unlock()

		}
	}
}

// isAvailable 判断ability对应的列表有无可用的设备
func (w *DeviceWorker) isAvailable(ability string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, device := range w.DeviceMap.MapTable[ability] {
		if device.Status.GroupID == "" {
			return true
		}
	}
	return false
}

func (w *DeviceWorker) getDevice(ability string) *apis.Device {
	for _, device := range w.DeviceMap.MapTable[ability] {
		if device.Status.GroupID == "" {
			return device
		}
	}
	return nil
}

func (w *DeviceWorker) updateDevice(ability string, device *apis.Device) {
	for index, d := range w.DeviceMap.MapTable[ability] {
		if d.Name == device.Name {
			w.DeviceMap.MapTable[ability][index] = device
		}
	}
}
