package rmf

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"log"
	"testing"
)

func TestGetDeviceInfo(t *testing.T) {
	done := make(chan struct{})

	go func() {
		var RMFProvider Provider = &DeviceInfoProvider{}
		// 构造一个DeviceList供测试使用
		DeviceList := ConstructDeviceListForTest()
		DeviceInfoMap, err := RMFProvider.GetDeviceInfo(DeviceList)
		if err != nil {
			log.Fatal(err)
		}
		// 遍历返回的map
		t.Logf("------deviceinfo的测试结果------\n")
		for DeviceName, DeviceInfo := range DeviceInfoMap {
			// 打印key
			t.Logf("DeviceName: %s\n", DeviceName)
			// 遍历value
			for _, property := range DeviceInfo.Infos {
				if len(property.SubProperty) > 0 {
					t.Logf("\t|name: %s, type: %s, value: %s\n", property.Name, property.Type, property.Value)
					for _, subProperty := range property.SubProperty {
						t.Logf("\t\t|name: %s, type: %s, value: %s\n", subProperty.Name, subProperty.Type, subProperty.Value)
					}
				} else {
					t.Logf("\t|name: %s, type: %s, value: %s\n", property.Name, property.Type, property.Value)
				}
				//fmt.Printf("\talias: %s, property: %s, type: %s, value: %s\n", property.Alias, property.Name, property.Type, property.Value)
			}
			t.Log("\n")
		}

		done <- struct{}{}
	}()

	<-done
	close(done)
}

func ConstructDeviceListForTest() []apis.Device {

	FixedArmAccessMethod := apis.AccessMethod{
		Type:  apis.AccessByRmf,
		URL:   "http://192.168.1.225:8000",
		Group: "tinyRobot",
		Alias: "fixedArm",
	}
	MobileArmAccessMethod := apis.AccessMethod{
		Type:  apis.AccessByRmf,
		URL:   "http://192.168.1.225:8000",
		Group: "tinyRobot",
		Alias: "mobileArm",
	}
	device1 := apis.Device{
		Spec: apis.DeviceSpec{
			Name:         "device1",
			AccessMethod: FixedArmAccessMethod,
		},
	}
	device2 := apis.Device{
		Spec: apis.DeviceSpec{
			Name:         "device2",
			AccessMethod: MobileArmAccessMethod,
		},
	}
	return []apis.Device{device1, device2}
}
