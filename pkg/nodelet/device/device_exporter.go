package device

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	m "hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/device/collector"
	"hit.edu/framework/pkg/nodelet/device/collector/ability"
	"net/http"
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
	//TODO 添加设备 device.json  replace 设备注册
	return de, nil
}

// 想改成每隔60秒收集一次静态信息，每隔1s收集一次动态信息
func (n *DeviceExporter) Run() error {
	// TODO: 开始运行
	clientSet, err := InitClient()
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] init client failed")
		return err
	}
	clientManager := m.NewManager(clientSet)

	//// 定期Gather一次数据
	//err := n.deviceCollector.GatherStaticData()
	//if err != nil {
	//	return err
	//}
	//
	//err = n.deviceCollector.GatherDynamicData()
	//if err != nil {
	//	return err
	//}
	//
	//staticTicker := time.NewTicker(time.Second * 30)
	//defer staticTicker.Stop()
	//dynamicTicker := time.NewTicker(time.Second * 5)
	//defer dynamicTicker.Stop()

	go func() {
		for {
			err1 := manager.MonitorAllDevices(clientManager)
			if err1 != nil {
				logs.Errorf("[DEVICE EXPORTER] monitor err: %v", err)
			}
			time.Sleep(20 * time.Second)
		}
	}()

	go func() {
		for {
			err2 := manager.MonitorAllAbilities(clientManager)
			if err2 != nil {
				logs.Errorf("[DEVICE EXPORTER] monitor err: %v", err)
			}
			time.Sleep(20 * time.Second)
		}
	}()
	// 加入关于Ability的信息收集
	//deviceMonitorTicker := time.NewTicker(time.Second * 20)
	//abilityMonitorTicker := time.NewTicker(time.Second * 20)
	//defer deviceMonitorTicker.Stop()
	//
	//for {
	//	select {
	//	//case <-staticTicker.C:
	//	//	err := n.deviceCollector.GatherStaticData()
	//	//	if err != nil {
	//	//		return err
	//	//	}
	//	//case <-dynamicTicker.C:
	//	//	err := n.deviceCollector.GatherDynamicData()
	//	//	if err != nil {
	//	//		return err
	//	//	}
	//	case <-deviceMonitorTicker.C:
	//		logs.Infof("[DEVICE EXPORTER] monitor all device....")
	//		go func() {
	//			err := manager.MonitorAllDevices(clientManager)
	//			if err != nil {
	//				logs.Errorf("[DEVICE EXPORTER] monitor err: %v", err)
	//			}
	//		}()
	//	case <-abilityMonitorTicker.C:
	//		logs.Infof("[DEVICE EXPORTER] monitor all device....")
	//		go func() {
	//			err := manager.MonitorAllAbilities(clientManager)
	//			if err != nil {
	//				logs.Errorf("[DEVICE EXPORTER] monitor err: %v", err)
	//			}
	//		}()
	//	}

	//}
	// TODO: 监控Node,查看是否有与Node相关联的新的Device
	// 更新deviceCollector中的DeviceList

	// 将收集的数据更新到API-Server中
	// 更新Node

	// TODO: 停止运行
	return nil
}

func InitClient() (*clients.ClientSet, error) {
	//初始化ClientSet客户端
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize clientSet: %v", err)
	}
	return clientSet, nil
}
