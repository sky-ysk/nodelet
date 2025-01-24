package collector

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
)

const namespace = "device"

const (
	defaultEnable  = true
	defaultDisable = false
	STATIC         = "static"
	DYNAMIC        = "dynamic"
)

const (
	// 通道内最大的Metric缓存数量
	capMetricChan = 1000
)

type NewCollectorFactory func() (Collector, error)

var (
	factories = make(map[string]NewCollectorFactory)
	cache     = make(map[string]Metric)
	cacheLock sync.RWMutex
)

// XXCollector模块在注册时，调用该函数，将XXCollector（cpu、storage、memory）的New函数 注册到StateMap中
func RegisterCollector(name string, isDefaultEnable bool, factory NewCollectorFactory) {
	if isDefaultEnable { //这个参数好像可以删除，已经在NewNodeCollector进行了XXCollector的选择性开启
		factories[name] = factory
	}
}

// XXCollect的接口 例如CPUCollector都得实现这个接口
type Collector interface {
	// 获取新的资源信息
	UpdateStaticInfo(ch chan<- Metric) error
	UpdateDynamicInfo(ch chan<- Metric) error
}

type DeviceCollector struct {
	// 所有注册的Collector
	Collectors map[string]Collector
}

// NewDeviceCollector 创建新的DeviceCollector
func NewDeviceCollector(enabledCollectors []string) (*DeviceCollector, error) {
	collectors := make(map[string]Collector)
	var err error
	var failedCollectors []string //如果感觉没有可以删除，记录无法初始化的XXCollector
	// TODO: 创建新的Collector
	// TODO: 实现Collector的注册逻辑
	// 遍历所有的构造函数，添加Collector
	for name, f := range factories {
		if contains(enabledCollectors, name) {
			fmt.Println("name:" + name)
			collectors[name], err = f()
			if err != nil {
				logs.V2().Errorf("Warning: failed to initialize collector %s: %v", name, err)
				failedCollectors = append(failedCollectors, name)
				continue
			}
			if collectors[name] == nil {
				logs.V2().Errorf("Warning: failed to initialize %s collector: collector does not exist", name)
			}
		}
	}
	if len(failedCollectors) > 0 {
		logs.V2().Errorf("Warning: failed to initialize the following collectors: %v", failedCollectors)
	}
	return &DeviceCollector{Collectors: collectors}, nil
}

// contains 判断slice中是否含有item
func contains(slice []string, item string) bool {
	for _, value := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// GatherStaticData 进行静态信息数据收集
func (n *DeviceCollector) GatherStaticData() error {
	
	var metricChan = make(chan Metric, capMetricChan)
	var wg sync.WaitGroup
	
	//wg.Add(len(n.Collectors))
	
	for name, c := range n.Collectors {
		if c == nil {
			logs.V2().Error("collector %s does not exist", name)
			continue
		}
		wg.Add(1)
		go func(name string, c Collector) {
			if err := c.UpdateStaticInfo(metricChan); err != nil {
				logs.V2().Error("failed to update static metrics: %v", err)
			}
			wg.Done()
		}(name, c)
	}
	
	// 所有的数据收集完后，清除Chan
	wg.Wait()
	close(metricChan)
	
	// 在主线程中接收 metricChan 中的数据
	for metric := range metricChan { //循环会阻塞，直到有数据进入 metricChan
		if err := processMetric(metric); err != nil {
			return err // 处理 metric 时出错，返回错误
		}
	}
	return nil // 处理成功完成
}

// GatherDynamicData 进行动态的数据收集
func (n *DeviceCollector) GatherDynamicData() error {
	
	var metricChan = make(chan Metric, capMetricChan)
	var wg sync.WaitGroup
	
	//wg.Add(len(n.Collectors))
	
	for name, c := range n.Collectors {
		wg.Add(1)
		go func(name string, c Collector) {
			if err := c.UpdateDynamicInfo(metricChan); err != nil {
				logs.V2().Error("failed to update static metrics: %v", err)
			}
			wg.Done()
		}(name, c)
	}
	
	// 所有的数据收集完后，清除Chan
	wg.Wait()
	close(metricChan)
	
	// 在主线程中接收 metricChan 中的数据
	for metric := range metricChan { //循环会阻塞，直到有数据进入 metricChan
		if err := processMetric(metric); err != nil {
			return err // 处理 metric 时出错，返回错误
		}
	}
	return nil // 处理成功完成
}

func processMetric(metric Metric) error {
	// TODO：处理Metric
	// TODO: 存入本地Cache或同步到manager中？待定
	
	//暂时先实现写入本地Cache
	key := fmt.Sprintf("%s-%v", metric.Item.name, metric.Item.labels)
	cacheLock.Lock()
	defer cacheLock.Unlock()
	cache[key] = metric // 将 metric 存储到缓存中
	logs.V2().Info("successfully processed metric: ", key, metric.toString())
	// TODO: 将Cache放到nodelet.go中
	// TODO: 数据预处理，然后放入Node中，然后写入API-Server中
	
	return nil
}
