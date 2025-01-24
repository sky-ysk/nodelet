package rmf

import (
	apis "hit.edu/framework/pkg/apis/cores"
)

// 每个设备的状态
type DeviceInfo struct {
	// 设备需要监控的属性
	Infos []apis.Property
}

type Provider interface {
	GetDeviceInfo(devices []apis.Device) (map[string]*DeviceInfo, error)
}

type DeviceInfoProvider struct{}

var _ Provider = &DeviceInfoProvider{}

// 通过RMF的接口获取设备的相关参数
func (d *DeviceInfoProvider) GetDeviceInfo(devices []apis.Device) (map[string]*DeviceInfo, error) {
	deviceInfos := make(map[string]*DeviceInfo)
	for _, device := range devices {
		// 获取设备访问方式
		accessMethod := device.Spec.AccessMethod
		if accessMethod.Type != apis.AccessByRmf {
			// TODO: 错误处理
			continue
		}

		// TODO: 构造访问请求
		request := Request{
			URL:    accessMethod.URL,
			Fleets: accessMethod.Group,
			Name:   accessMethod.Alias,
			Params: make([]string, 0),
		}

		deviceInfo, err := GetDeviceInfo(request)
		if err != nil {
			// TODO: 输出错误信息
			continue
		}
		deviceInfos[device.Spec.Name] = deviceInfo
	}
	return deviceInfos, nil
}
