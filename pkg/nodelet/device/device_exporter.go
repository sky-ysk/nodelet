package device

import (
	"hit.edu/framework/pkg/nodelet/device/collector"
	"time"
)

type Exporter interface {
	// TODO: 定义接口
	Run() error
}

var _ Exporter = &DeviceExporter{}

type DeviceExporter struct {
	// Register
	// TODO: 扫描节点信息，并注册到API-Server中
	// TODO: 定义组件
	deviceCollector *collector.DeviceCollector
}

func NewDeviceExporter(cfg *Config) (*DeviceExporter, error) {
	// TODO：参数配置
	// 创建NodeCollector 读取配置信息
	dc, err := collector.NewDeviceCollector(cfg.EnabledCollectors)
	if err != nil {
		return nil, err
	}
	de := &DeviceExporter{
		deviceCollector: dc,
	}
	return de, nil
}

// 想改成每隔60秒收集一次静态信息，每隔1s收集一次动态信息
func (n *DeviceExporter) Run() error {
	// TODO: 开始运行
	
	// 定期Gather一次数据
	err := n.deviceCollector.GatherStaticData()
	if err != nil {
		return err
	}
	
	err = n.deviceCollector.GatherDynamicData()
	if err != nil {
		return err
	}
	
	staticTicker := time.NewTicker(time.Second * 30)
	defer staticTicker.Stop()
	dynamicTicker := time.NewTicker(time.Second * 5)
	defer dynamicTicker.Stop()
	
	for {
		select {
		case <-staticTicker.C:
			err := n.deviceCollector.GatherStaticData()
			if err != nil {
				return err
			}
		case <-dynamicTicker.C:
			err := n.deviceCollector.GatherDynamicData()
			if err != nil {
				return err
			}
		}
		
	}
	// TODO: 监控Node,查看是否有与Node相关联的新的Device
	// 更新deviceCollector中的DeviceList
	
	// 将收集的数据更新到API-Server中
	// 更新Node
	
	// TODO: 停止运行
	return nil
}
