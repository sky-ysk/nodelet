package collector

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/nodelet/device/collector/rmf"
)

// TODO: 根据Config配置该选项 ，是否启用RMFCollector
var enableRMFInfo = true //改成从配置文件当中读取
const rmfCollectorName = "RMF"

// RMF Collector收集设备属性中的内容
type RMFCollector struct {
	// 访问方式为RMF的设备，对应的Item
	// 每个设备对应一组Item
	deviceInfos []*Item

	// provider 信息的提供者
	provider rmf.DeviceInfoProvider

	// 设备列表
	devices []apis.Device
}

var (
	// 是否启用status收集
	// TODO: 根据Config配置该选项
	enableStatusInfo *bool
)

func init() { //它在包级别的变量初始化之后，自动调用，不需要显式调用 ---也就是说main入口函数导入了collector包，他就会被调用
	// 向NodeCollector注册自身
	RegisterCollector(rmfCollectorName, enableRMFInfo, NewRMFCollector)
	fmt.Println("init RMF Collector")
}

func NewRMFCollector() (Collector, error) {
	var provider rmf.DeviceInfoProvider
	// 初始化时，设备列表中没有设备
	c := &RMFCollector{
		deviceInfos: []*Item{},
		provider:    provider,
		devices:     []apis.Device{},
	}
	return c, nil
}

func (c *RMFCollector) UpdateStaticInfo(ch chan<- Metric) error {
	//TODO implement me
	panic("implement me")
}

func (c *RMFCollector) UpdateDynamicInfo(ch chan<- Metric) error {
	if err := c.updateInfo(DYNAMIC); err != nil {
		return err
	}

	// 将动态CPU 信息逐个发送到通道
	for _, info := range c.deviceInfos {
		ch <- NewMetric(info)
	}
	return nil
}

func (c *RMFCollector) updateInfo(infoType string) error {
	//
	if infoType == DYNAMIC {
		// TODO: 获取设备动态字段
		DeviceInfos, err := c.provider.GetDeviceInfo(c.devices)
		if err != nil {
			return err
		}

		for _, device := range c.devices {
			// 构造Labels
			infos := DeviceInfos[device.Name].Infos
			labels := make([]string, len(infos))
			values := make(map[string]apis.Property)
			for i, info := range infos {
				labels[i] = info.Name
				values[info.Name] = info
			}
			item := NewItem(device.Name, "", labels)
			item.UpdateValues(values)

			c.deviceInfos = append(c.deviceInfos, item)
		}
	} else if infoType == STATIC {
		// TODO: 目前Device没有需要收集的静态字段
	} else {

	}
	return nil
}
